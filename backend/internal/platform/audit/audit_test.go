package audit

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping live audit integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func TestWriter_Record(t *testing.T) {
	pool := testPool(t)
	writer := NewWriter(pool)
	corrID := uuid.New()

	err := writer.Record(context.Background(), Entry{
		Action:        "test.chain",
		ResourceType:  "test_resource",
		ResourceID:    "example",
		Metadata:      map[string]any{"result": "online"},
		CorrelationID: &corrID,
		IPAddress:     "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	var action, resourceType, resourceID string
	err = pool.QueryRow(context.Background(),
		`SELECT action, resource_type, resource_id FROM audit_logs WHERE correlation_id = $1`, corrID,
	).Scan(&action, &resourceType, &resourceID)
	if err != nil {
		t.Fatalf("query row: %v", err)
	}

	if action != "test.chain" {
		t.Errorf("action = %q", action)
	}
	if resourceType != "test_resource" {
		t.Errorf("resource_type = %q", resourceType)
	}
	if resourceID != "example" {
		t.Errorf("resource_id = %q", resourceID)
	}
}

// Prova em cima do banco real (não apenas documentado em comentário) que
// a migration 000008 realmente bloqueia UPDATE/DELETE/TRUNCATE em
// audit_logs — uma trilha de auditoria que pode ser editada não prova
// nada.
func TestAuditLogs_AreImmutable(t *testing.T) {
	pool := testPool(t)
	writer := NewWriter(pool)
	ctx := context.Background()
	corrID := uuid.New()

	if err := writer.Record(ctx, Entry{Action: "test.immutability", CorrelationID: &corrID}); err != nil {
		t.Fatalf("Record: %v", err)
	}

	t.Run("UPDATE é rejeitado", func(t *testing.T) {
		_, err := pool.Exec(ctx, `UPDATE audit_logs SET action = 'tampered' WHERE correlation_id = $1`, corrID)
		if err == nil {
			t.Fatal("esperava erro ao tentar UPDATE em audit_logs, mas foi bem-sucedido")
		}
	})

	t.Run("DELETE é rejeitado", func(t *testing.T) {
		_, err := pool.Exec(ctx, `DELETE FROM audit_logs WHERE correlation_id = $1`, corrID)
		if err == nil {
			t.Fatal("esperava erro ao tentar DELETE em audit_logs, mas foi bem-sucedido")
		}
	})

	t.Run("TRUNCATE é rejeitado", func(t *testing.T) {
		_, err := pool.Exec(ctx, `TRUNCATE audit_logs`)
		if err == nil {
			t.Fatal("esperava erro ao tentar TRUNCATE em audit_logs, mas foi bem-sucedido")
		}
	})

	// A linha original precisa continuar exatamente como foi gravada —
	// nenhuma das tentativas acima deve ter alterado nada.
	var action string
	err := pool.QueryRow(ctx, `SELECT action FROM audit_logs WHERE correlation_id = $1`, corrID).Scan(&action)
	if err != nil {
		t.Fatalf("query row: %v", err)
	}
	if action != "test.immutability" {
		t.Errorf("action = %q, want unchanged %q", action, "test.immutability")
	}
}

func TestWriter_Record_NilUserAndEmptyIP(t *testing.T) {
	pool := testPool(t)
	writer := NewWriter(pool)
	corrID := uuid.New()

	err := writer.Record(context.Background(), Entry{
		Action:        ActionLogin,
		CorrelationID: &corrID,
	})
	if err != nil {
		t.Fatalf("Record with nil user/empty IP: %v", err)
	}
}

// TestAuditChain_VerifiesAndDetectsTampering prova, sobre o banco real,
// que cada linha nova estende a cadeia SHA-256 e que audit_verify_chain
// acusa uma adulteração feita por quem contorna o gatilho de imutabilidade
// (ex.: um superusuário que desabilita o trigger).
func TestAuditChain_VerifiesAndDetectsTampering(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	w := NewWriter(pool)
	for i := 0; i < 3; i++ {
		if err := w.Record(ctx, Entry{Action: "test.chain.link", Before: map[string]int{"n": i}, After: map[string]int{"n": i + 1}}); err != nil {
			t.Fatal(err)
		}
	}
	res, err := NewReader(pool).Verify(ctx, 1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Valid || res.Checked < 3 {
		t.Fatalf("cadeia íntegra deveria validar: %+v", res)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `ALTER TABLE audit_logs DISABLE TRIGGER trg_audit_logs_immutable`); err != nil {
		t.Skipf("sem privilégio para simular adulteração: %v", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE audit_logs SET action = 'adulterado' WHERE chain_pos = (SELECT max(chain_pos) FROM audit_logs)`); err != nil {
		t.Fatal(err)
	}
	var valid bool
	if err := tx.QueryRow(ctx, `SELECT valid FROM audit_verify_chain(1, NULL)`).Scan(&valid); err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Fatal("adulteração não foi detectada pela verificação da cadeia")
	}
}
