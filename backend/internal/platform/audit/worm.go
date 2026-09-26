package audit

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yurythx/projeto-nexus/internal/platform/database"
	"github.com/yurythx/projeto-nexus/internal/platform/storage"
)

// wormInterval é de quanto em quanto tempo o worker verifica se há dias
// completos ainda não exportados. Diário seria suficiente; 6h dá margem
// para o worker ter ficado fora do ar por um tempo sem acumular atraso.
const wormInterval = 6 * time.Hour

type wormExporter struct {
	db            database.DBTX
	store         storage.Provider
	worm          storage.WORMWriter // != nil quando store suporta object-lock
	bucket        string
	retentionDays int
	logger        *slog.Logger
	bucketReady   bool
	interval      time.Duration
	now           func() time.Time
}

// WORMExporter é um processor do worker. Serializa cada dia COMPLETO de audit_logs ainda não
// exportado, encadeia o SHA-256 ao do dia anterior (evidência de
// adulteração) e sobe arquivo + digest para o object storage. F2.6.
//
// Se o provider implementa storage.WORMWriter, o bucket é criado com
// object-lock e cada objeto recebe retenção Compliance de retentionDays;
// caso contrário, cai para Put comum e registra o aviso (a cadeia de
// hash ainda dá evidência de adulteração).
func WORMExporter(pool *pgxpool.Pool, store storage.Provider, bucket string, retentionDays int, logger *slog.Logger) func(ctx context.Context) error {
	return newWORMExporter(pool, store, bucket, retentionDays, logger).loop
}

func newWORMExporter(db database.DBTX, store storage.Provider, bucket string, retentionDays int, logger *slog.Logger) *wormExporter {
	e := &wormExporter{db: db, store: store, bucket: bucket, retentionDays: retentionDays, logger: logger,
		interval: wormInterval, now: time.Now}
	if w, ok := store.(storage.WORMWriter); ok {
		e.worm = w
	}
	return e
}

func (e *wormExporter) loop(ctx context.Context) error {
	e.run(ctx)
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			e.run(ctx)
		}
	}
}

func (e *wormExporter) ensureBucket(ctx context.Context) bool {
	if e.bucketReady {
		return true
	}
	if e.worm != nil {
		if err := e.worm.EnsureImmutableBucket(ctx, e.bucket, e.retentionDays); err != nil {
			e.logger.Error("audit worm: não foi possível preparar o bucket imutável — object-lock indisponível; a cadeia de hash segue como evidência",
				slog.String("bucket", e.bucket), slog.Any("error", err))
			e.worm = nil // cai para Put comum daqui pra frente
		}
	}
	if e.worm == nil {
		if be, ok := e.store.(interface {
			EnsureBucket(context.Context, string) error
		}); ok {
			if err := be.EnsureBucket(ctx, e.bucket); err != nil {
				e.logger.Error("audit worm: não foi possível garantir o bucket", slog.Any("error", err))
				return false
			}
		}
		e.logger.Warn("audit worm: bucket SEM object-lock — configure um bucket dedicado com lock na infraestrutura para a garantia WORM completa",
			slog.String("bucket", e.bucket))
	}
	e.bucketReady = true
	return true
}

// run exporta, em ordem, todos os dias completos (até ontem, UTC) ainda
// não exportados. As exportações são sempre contíguas — um dia só sai
// depois do anterior —, então basta retomar do dia seguinte ao último
// exportado (a marca d'água), sem revisitar todo o histórico a cada tick.
func (e *wormExporter) run(ctx context.Context) {
	if !e.ensureBucket(ctx) {
		return
	}
	day, prevSHA, err := e.nextDay(ctx)
	if err != nil {
		e.logger.Error("audit worm: consulta de marca d'água falhou", slog.Any("error", err))
		return
	}
	if day.IsZero() {
		return // trilha vazia
	}
	yesterday := e.now().UTC().Truncate(24 * time.Hour).Add(-24 * time.Hour)
	for ; !day.After(yesterday); day = day.Add(24 * time.Hour) {
		sha, err := e.exportOneDay(ctx, day, prevSHA)
		if err != nil {
			e.logger.Error("audit worm: falha ao exportar o dia",
				slog.String("day", day.Format("2006-01-02")), slog.Any("error", err))
			return // tenta de novo no próximo tick; a cadeia precisa ser sequencial
		}
		e.logger.Info("audit worm: dia exportado", slog.String("day", day.Format("2006-01-02")))
		prevSHA = sha
	}
}

// nextDay devolve o primeiro dia ainda não exportado e o SHA-256 do último
// exportado (ao qual o próximo se encadeia). Sem exportação alguma, começa
// no dia do registro mais antigo; trilha vazia devolve o tempo zero.
func (e *wormExporter) nextDay(ctx context.Context) (time.Time, string, error) {
	var last time.Time
	var sha string
	err := e.db.QueryRow(ctx, `SELECT day, sha256 FROM audit_worm_exports ORDER BY day DESC LIMIT 1`).Scan(&last, &sha)
	if err == nil {
		return last.UTC().Truncate(24 * time.Hour).Add(24 * time.Hour), sha, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, "", err
	}
	var earliest *time.Time
	if err := e.db.QueryRow(ctx, `SELECT min(created_at) FROM audit_logs`).Scan(&earliest); err != nil {
		return time.Time{}, "", err
	}
	if earliest == nil {
		return time.Time{}, "", nil
	}
	return earliest.UTC().Truncate(24 * time.Hour), "", nil
}

func (e *wormExporter) put(ctx context.Context, key string, body []byte, contentType string) error {
	if e.worm != nil {
		return e.worm.PutImmutable(ctx, e.bucket, key, bytes.NewReader(body), int64(len(body)), contentType, e.retentionDays)
	}
	return e.store.Put(ctx, e.bucket, key, bytes.NewReader(body), int64(len(body)), contentType)
}

// exportOneDay grava o arquivo do dia (encadeado a prevSHA), o digest e a
// marca d'água; devolve o SHA-256 do arquivo.
func (e *wormExporter) exportOneDay(ctx context.Context, day time.Time, prevSHA string) (string, error) {
	next := day.Add(24 * time.Hour)
	rows, err := e.db.Query(ctx, `
		SELECT id, COALESCE(actor_id::text,''), action, COALESCE(resource_type,''),
		       COALESCE(resource_id,''), COALESCE(metadata::text,'{}'),
		       COALESCE(correlation_id::text,''), COALESCE(host(ip_address),''), created_at,
		       chain_pos, prev_hash, hash
		FROM audit_logs
		WHERE created_at >= $1 AND created_at < $2
		ORDER BY chain_pos ASC`, day, next)
	if err != nil {
		return "", fmt.Errorf("query day: %w", err)
	}
	defer rows.Close()

	var buf bytes.Buffer
	header, _ := json.Marshal(map[string]any{
		"_worm_header": true,
		"day":          day.Format("2006-01-02"),
		"prev_sha256":  prevSHA,
		"generated_at": e.now().UTC().Format(time.RFC3339),
	})
	buf.Write(header)
	buf.WriteByte('\n')

	count := 0
	for rows.Next() {
		var id, actorID, action, rType, rID, meta, corr, ip, prevHash, hash string
		var createdAt time.Time
		var chainPos int64
		if err := rows.Scan(&id, &actorID, &action, &rType, &rID, &meta, &corr, &ip, &createdAt, &chainPos, &prevHash, &hash); err != nil {
			return "", fmt.Errorf("scan: %w", err)
		}
		// O hash da cadeia vai junto em cada linha: a cópia WORM permite
		// provar, fora do banco, que a sequência não foi alterada.
		line, _ := json.Marshal(map[string]any{
			"id": id, "actor_id": actorID, "action": action, "resource_type": rType,
			"resource_id": rID, "metadata": json.RawMessage(meta), "correlation_id": corr,
			"ip_address": ip, "created_at": createdAt.UTC().Format(time.RFC3339Nano),
			"chain_pos": chainPos, "prev_hash": prevHash, "hash": hash,
		})
		buf.Write(line)
		buf.WriteByte('\n')
		count++
	}
	if rows.Err() != nil {
		return "", fmt.Errorf("rows: %w", rows.Err())
	}

	sum := sha256.Sum256(buf.Bytes())
	digest := hex.EncodeToString(sum[:])
	objectKey := fmt.Sprintf("%s/%s.jsonl", day.Format("2006/01"), day.Format("2006-01-02"))

	if err := e.put(ctx, objectKey, buf.Bytes(), "application/x-ndjson"); err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}
	digestBody := []byte(digest + "  " + day.Format("2006-01-02") + ".jsonl\n")
	if err := e.put(ctx, objectKey+".sha256", digestBody, "text/plain"); err != nil {
		return "", fmt.Errorf("put digest: %w", err)
	}

	if _, err := e.db.Exec(ctx, `
		INSERT INTO audit_worm_exports (day, row_count, sha256, prev_sha256, object_key)
		VALUES ($1, $2, $3, $4, $5)`, day, count, digest, prevSHA, objectKey); err != nil {
		return "", fmt.Errorf("insert watermark: %w", err)
	}

	if err := NewWriter(e.db).Record(ctx, Entry{
		Action:       "audit.worm.exported",
		ResourceType: "audit_worm_exports",
		ResourceID:   day.Format("2006-01-02"),
		Metadata: map[string]any{
			"rows": count, "sha256": digest, "object_key": objectKey,
			"bucket": e.bucket, "immutable": e.worm != nil,
		},
	}); err != nil {
		// O dia já está exportado e marcado; só a trilha do próprio ato falhou.
		e.logger.Warn("audit worm: falha ao registrar a exportação na trilha", slog.Any("error", err))
	}
	return digest, nil
}
