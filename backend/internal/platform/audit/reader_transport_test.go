package audit

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/yurythx/projeto-nexus/internal/domain/pagination"
	"github.com/yurythx/projeto-nexus/internal/platform/auth"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
	"github.com/yurythx/projeto-nexus/internal/platform/database/dbtest"
)

func TestReaderFiltersAndEscapesWildcards(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	actor := dbtest.User(t, pool)
	sfx := uuid.NewString()[:8]
	action := "esc." + sfx + ".a_b"
	if err := NewWriter(pool).Record(ctx, Entry{ActorID: &actor, Action: action, ResourceType: "rt", ResourceID: sfx}); err != nil {
		t.Fatal(err)
	}
	r := NewReader(pool)
	from, to := time.Now().Add(-time.Hour), time.Now().Add(time.Hour)
	p := pagination.New(1, 10, 10)
	recs, total, err := r.List(ctx, Filter{Action: "esc." + sfx + ".*", ActorID: &actor, ResourceType: "rt", ResourceID: sfx, From: &from, To: &to}, p)
	if err != nil || total != 1 || len(recs) != 1 || recs[0].Action != action {
		t.Fatalf("todos os filtros: %d %v %v", total, recs, err)
	}
	for _, f := range []string{"esc." + sfx + ".a%", "esc." + sfx + ".__b", "esc." + sfx + "_a_b"} {
		if _, n, err := r.List(ctx, Filter{Action: f}, p); err != nil || n != 0 {
			t.Errorf("%q: %% e _ digitados valem como texto, não como curinga (%d, %v)", f, n, err)
		}
	}
	got, err := r.Get(ctx, recs[0].ID)
	if err != nil || got.ResourceID != sfx || got.Hash == "" {
		t.Fatalf("Get: %+v %v", got, err)
	}
	if _, err := r.Get(ctx, uuid.New()); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("Get inexistente: %v", err)
	}
	if res, err := r.Verify(ctx, 1, 5); err != nil || res.Checked > 5 {
		t.Fatalf("Verify limitado: %+v %v", res, err)
	}
}

func TestReaderPropagatesDatabaseErrors(t *testing.T) {
	ctx := context.Background()
	p := pagination.New(1, 10, 10)
	count := scanRow(func(dest ...any) error { *(dest[0].(*int64)) = 1; return nil })
	for name, db := range map[string]database.DBTX{
		"contagem":             dbtest.Fail{},
		"página":               &dbtest.Seq{Row: count, Queries: []dbtest.QueryResult{{Err: dbtest.ErrInjected}}},
		"linha ilegível":       &dbtest.Seq{Row: count, Queries: []dbtest.QueryResult{{Rows: dbtest.BadRows()}}},
		"leitura interrompida": &dbtest.Seq{Row: count, Queries: []dbtest.QueryResult{{Rows: dbtest.ErrRows()}}},
	} {
		if _, _, err := NewReader(db).List(ctx, Filter{}, p); err == nil {
			t.Errorf("List (%s) deveria falhar", name)
		}
	}
	if _, err := NewReader(dbtest.Fail{}).Verify(ctx, 1, 0); !errors.Is(err, dbtest.ErrInjected) {
		t.Errorf("Verify com o banco fora: %v", err)
	}
	invalid := scanRow(func(dest ...any) error {
		*(dest[0].(*int64)) = 7
		pos := int64(3)
		*(dest[2].(**int64)) = &pos
		reason := "hash divergente"
		*(dest[3].(**string)) = &reason
		return nil
	})
	res, err := NewReader(&dbtest.Seq{Row: invalid}).Verify(ctx, 1, 0)
	if err != nil || res.Reason != "hash divergente" || *res.FirstInvalidPos != 3 {
		t.Fatalf("motivo da quebra da cadeia: %+v %v", res, err)
	}
}

// router monta as rotas de auditoria com um usuário de permissão total.
func router(reader *Reader, exp *Exporter) http.Handler {
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req.WithContext(auth.WithIdentity(req.Context(), auth.Identity{Permissions: []string{"*"}})))
		})
	})
	RegisterRoutes(r, NewHandlers(reader, exp, quietLogger(), 100))
	return r
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestHandlers(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()
	actor := dbtest.User(t, pool)
	sfx := uuid.NewString()[:8]
	if err := NewWriter(pool).Record(ctx, Entry{ActorID: &actor, Action: "handlers." + sfx, ResourceID: sfx}); err != nil {
		t.Fatal(err)
	}
	h := router(NewReader(pool), NewExporter(pool, quietLogger()))
	rec := get(h, "/audit/logs?actor_id="+actor.String()+"&from=2000-01-01&to="+time.Now().Add(time.Hour).UTC().Format(time.RFC3339)+"&action=handlers."+sfx)
	var list struct {
		Data []Record `json:"data"`
	}
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &list) != nil || len(list.Data) != 1 {
		t.Fatalf("lista filtrada: %d %s", rec.Code, rec.Body.String())
	}
	id := list.Data[0].ID.String()
	for path, want := range map[string]int{
		"/audit/logs/" + id:               http.StatusOK,
		"/audit/logs/" + uuid.NewString(): http.StatusNotFound,
		"/audit/logs/nao-uuid":            http.StatusBadRequest,
		"/audit/logs?to=amanha":           http.StatusBadRequest,
		"/audit/verify?from=0&limit=3":    http.StatusOK,
	} {
		if rec := get(h, path); rec.Code != want {
			t.Errorf("%s: %d, esperado %d (%s)", path, rec.Code, want, rec.Body.String())
		}
	}

	// Banco fora: 500 sem vazar o erro interno.
	down := router(NewReader(dbtest.Fail{}), NewExporter(dbtest.Fail{}, quietLogger()))
	for _, path := range []string{"/audit/logs", "/audit/logs/" + id, "/audit/verify"} {
		if rec := get(down, path); rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), dbtest.ErrInjected.Error()) {
			t.Errorf("%s com o banco fora: %d %s", path, rec.Code, rec.Body.String())
		}
	}

	// A verificação responde mesmo se a trilha do próprio ato falhar.
	ok := scanRow(func(dest ...any) error { *(dest[0].(*int64)) = 1; *(dest[1].(*bool)) = true; return nil })
	noAudit := router(NewReader(&dbtest.Seq{Row: ok}), NewExporter(dbtest.Fail{}, quietLogger()))
	if rec := get(noAudit, "/audit/verify"); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"valid":true`) {
		t.Errorf("verificação sem trilha: %d %s", rec.Code, rec.Body.String())
	}
}
