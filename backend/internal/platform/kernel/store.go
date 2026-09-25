package kernel

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yurythx/projeto-nexus/internal/platform/audit"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
)

// Store persiste o estado de ativação dos plugins.
type Store interface {
	// Ensure cria a linha do módulo com o estado default, se ainda não existir.
	Ensure(ctx context.Context, key string, enabled bool) error
	// LoadAll devolve o estado persistido de todos os módulos.
	LoadAll(ctx context.Context) (map[string]bool, error)
	// Set grava o novo estado e a entrada de auditoria na MESMA transação.
	Set(ctx context.Context, key string, enabled bool, actor string, entry audit.Entry) error
	// Listen bloqueia até ctx acabar, chamando onChange a cada NOTIFY.
	Listen(ctx context.Context, onChange func()) error
}

// ModulesChannel é o canal LISTEN/NOTIFY disparado pelo gatilho de
// system_modules (migration 000100).
const ModulesChannel = "nexus_modules_channel"

// PostgresStore implementa Store sobre system_modules.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewPostgresStore cria o Store.
func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Ensure(ctx context.Context, key string, enabled bool) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO system_modules (key, enabled) VALUES ($1, $2) ON CONFLICT (key) DO NOTHING`, key, enabled)
	if err != nil {
		return fmt.Errorf("kernel: ensure module %s: %w", key, err)
	}
	return nil
}

func (s *PostgresStore) LoadAll(ctx context.Context) (map[string]bool, error) {
	rows, err := s.pool.Query(ctx, `SELECT key, enabled FROM system_modules`)
	if err != nil {
		return nil, fmt.Errorf("kernel: load modules: %w", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var k string
		var e bool
		if err := rows.Scan(&k, &e); err != nil {
			return nil, err
		}
		out[k] = e
	}
	return out, rows.Err()
}

func (s *PostgresStore) Set(ctx context.Context, key string, enabled bool, actor string, entry audit.Entry) error {
	return database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			INSERT INTO system_modules (key, enabled, updated_at, updated_by) VALUES ($1, $2, now(), $3)
			ON CONFLICT (key) DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = now(), updated_by = EXCLUDED.updated_by`,
			key, enabled, actor); err != nil {
			return fmt.Errorf("kernel: set module %s: %w", key, err)
		}
		return audit.NewWriter(tx).Record(ctx, entry)
	})
}

// Listen mantém uma conexão dedicada em LISTEN e reconecta com backoff.
func (s *PostgresStore) Listen(ctx context.Context, onChange func()) error {
	backoff := time.Second
	for {
		err := s.listenOnce(ctx, onChange)
		if ctx.Err() != nil {
			return nil
		}
		_ = err
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func (s *PostgresStore) listenOnce(ctx context.Context, onChange func()) error {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, "LISTEN "+ModulesChannel); err != nil {
		return err
	}
	// Recarrega logo após (re)conectar: pode ter havido mudança enquanto
	// a conexão estava fora.
	onChange()
	for {
		if _, err := conn.Conn().WaitForNotification(ctx); err != nil {
			return err
		}
		onChange()
	}
}
