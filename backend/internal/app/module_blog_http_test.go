package app

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Capa: trocar ou excluir remove a imagem antiga do armazenamento; texto
// vazio não se publica; estados e entrada malformada.
func TestBlogCoverAndPublicationRules(t *testing.T) {
	h := newHarness(t)
	admin := h.admin()
	ctx := context.Background()
	bucket := h.d.Config.MinIO.Bucket
	cover := func() string {
		tk := data[struct {
			ObjectKey string `json:"object_key"`
		}](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/blog/uploads", admin, `{"filename":"capa.png","content_type":"image/png"}`))
		h.upload(tk.ObjectKey, "image/png", []byte("\x89PNG capa"))
		return tk.ObjectKey
	}
	c1, c2 := cover(), cover()
	title := "Aviso " + uuid.NewString()[:6]
	p := data[postResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/blog/posts", admin, `{"title":"`+title+`","cover_object_key":"`+c1+`"}`))
	if got := h.expect(http.StatusOK, http.MethodGet, "/api/v1/blog/posts/"+p.ID, admin, "").Body.String(); !strings.Contains(got, `"cover_url":"https://minio.test/`) {
		t.Fatalf("capa servida por URL temporária: %s", got)
	}
	h.expect(http.StatusUnprocessableEntity, http.MethodPost, "/api/v1/blog/posts/"+p.ID+"/publish", admin, "")
	h.expect(http.StatusOK, http.MethodPut, "/api/v1/blog/posts/"+p.ID, admin, `{"title":"`+title+`","body":"Texto","cover_object_key":"`+c2+`"}`)
	if _, err := h.d.Storage.Stat(ctx, bucket, c1); err == nil {
		t.Fatal("a capa substituída sai do armazenamento")
	}
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/blog/posts/"+p.ID+"/publish", admin, "")
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/blog/posts/"+p.ID+"/unpublish", admin, "")
	h.expect(http.StatusConflict, http.MethodPost, "/api/v1/blog/posts/"+p.ID+"/unpublish", admin, "")
	h.expect(http.StatusNoContent, http.MethodDelete, "/api/v1/blog/posts/"+p.ID, admin, "")
	if _, err := h.d.Storage.Stat(ctx, bucket, c2); err == nil {
		t.Fatal("excluir a publicação remove a capa")
	}
	for _, c := range []struct{ method, path, body string }{
		{http.MethodPut, "/api/v1/blog/posts/x", `{"title":"x"}`},
		{http.MethodPost, "/api/v1/blog/posts/x/publish", ""},
		{http.MethodDelete, "/api/v1/blog/posts/x", ""},
		{http.MethodPost, "/api/v1/blog/posts", `{`},
		{http.MethodPut, "/api/v1/blog/posts/" + p.ID, `{`},
		{http.MethodPost, "/api/v1/blog/uploads", `{`},
	} {
		h.expect(http.StatusBadRequest, c.method, c.path, admin, c.body)
	}
	for _, c := range []struct{ method, path, body string }{
		{http.MethodPut, "/api/v1/blog/posts/" + uuid.NewString(), `{"title":"Fantasma"}`},
		{http.MethodPost, "/api/v1/blog/posts/" + uuid.NewString() + "/publish", ""},
		{http.MethodDelete, "/api/v1/blog/posts/" + uuid.NewString(), ""},
		{http.MethodGet, "/api/v1/blog/posts/nao-existe-" + uuid.NewString()[:6], ""},
	} {
		h.expect(http.StatusNotFound, c.method, c.path, admin, c.body)
	}
	h.expect(http.StatusUnprocessableEntity, http.MethodPost, "/api/v1/blog/posts", admin, `{"title":"???"}`)
	h.expect(http.StatusOK, http.MethodGet, "/api/v1/blog/posts?status=archived", admin, "")
}
