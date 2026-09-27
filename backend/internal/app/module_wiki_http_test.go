package app

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Árvore: mãe inexistente, ciclo com neta, slug vazio, histórico de
// página inexistente e parâmetros malformados.
func TestWikiTreeRules(t *testing.T) {
	h := newHarness(t)
	_, ana := h.user("nexus-user")
	sfx := uuid.NewString()[:6]
	raiz := data[wikiResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/wiki/pages", ana, `{"title":"Raiz `+sfx+`"}`))
	filha := data[wikiResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/wiki/pages", ana, `{"title":"Filha `+sfx+`","parent_id":"`+raiz.ID+`"}`))
	neta := data[wikiResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/wiki/pages", ana, `{"title":"Neta `+sfx+`","parent_id":"`+filha.ID+`"}`))

	h.expect(http.StatusNotFound, http.MethodPost, "/api/v1/wiki/pages", ana, `{"title":"Órfã","parent_id":"`+uuid.NewString()+`"}`)
	h.expect(http.StatusUnprocessableEntity, http.MethodPost, "/api/v1/wiki/pages", ana, `{"title":"###"}`)
	h.expect(http.StatusConflict, http.MethodPost, "/api/v1/wiki/pages", ana, `{"title":"Outra","slug":"`+raiz.Slug+`"}`)
	h.expect(http.StatusConflict, http.MethodPut, "/api/v1/wiki/pages/"+raiz.ID, ana, `{"title":"Raiz","parent_id":"`+neta.ID+`","version":1}`)
	h.expect(http.StatusNotFound, http.MethodPut, "/api/v1/wiki/pages/"+raiz.ID, ana, `{"title":"Raiz","parent_id":"`+uuid.NewString()+`","version":1}`)
	moved := data[wikiResp](t, h.expect(http.StatusOK, http.MethodPut, "/api/v1/wiki/pages/"+neta.ID, ana, `{"title":"Neta promovida","parent_id":"`+raiz.ID+`","version":1}`))
	if moved.Version != 2 {
		t.Fatalf("mover gera nova versão: %+v", moved)
	}
	if body := h.expect(http.StatusOK, http.MethodGet, "/api/v1/wiki/pages/"+neta.ID, ana, "").Body.String(); !strings.Contains(body, neta.Slug) {
		t.Fatalf("slug mantido quando não informado: %s", body)
	}
	if res := h.expect(http.StatusOK, http.MethodGet, "/api/v1/search?q=Raiz+"+sfx+"&module=wiki", ana, "").Body.String(); !strings.Contains(res, raiz.ID) {
		t.Fatalf("busca global na wiki: %s", res)
	}

	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/wiki/pages/"+uuid.NewString()+"/revisions", ana, "")
	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/wiki/pages/"+raiz.ID+"/revisions/99", ana, "")
	h.expect(http.StatusNotFound, http.MethodPost, "/api/v1/wiki/pages/"+raiz.ID+"/revisions/99/restore", ana, "")
	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/wiki/pages/nao-existe-"+sfx, ana, "")
	h.expect(http.StatusNotFound, http.MethodPut, "/api/v1/wiki/pages/"+uuid.NewString(), ana, `{"title":"x","version":1}`)
	for _, c := range []struct{ method, path, body string }{
		{http.MethodPut, "/api/v1/wiki/pages/x", `{"title":"x","version":1}`},
		{http.MethodGet, "/api/v1/wiki/pages/x/revisions", ""},
		{http.MethodGet, "/api/v1/wiki/pages/x/revisions/1", ""},
		{http.MethodGet, "/api/v1/wiki/pages/" + raiz.ID + "/revisions/um", ""},
		{http.MethodPost, "/api/v1/wiki/pages/x/revisions/1/restore", ""},
		{http.MethodPost, "/api/v1/wiki/pages/" + raiz.ID + "/revisions/um/restore", ""},
		{http.MethodPost, "/api/v1/wiki/pages", `{`},
		{http.MethodPut, "/api/v1/wiki/pages/" + raiz.ID, `{`},
	} {
		h.expect(http.StatusBadRequest, c.method, c.path, ana, c.body)
	}
	admin := h.admin()
	h.expect(http.StatusBadRequest, http.MethodDelete, "/api/v1/wiki/pages/x", admin, "")
	h.expect(http.StatusNotFound, http.MethodDelete, "/api/v1/wiki/pages/"+uuid.NewString(), admin, "")
}

// Dono organizacional da página (ADR 013): marca como dona quem está
// lotado na unidade (ou tem wiki:manage cobrindo-a); subpágina herda a dona
// da mãe; excluir exige wiki:manage cobrindo a dona.
func TestWikiGestaoComEscopo(t *testing.T) {
	h := newHarness(t)
	admin := h.admin()
	o := h.orgEscopo(t)
	gestorA := h.gestorEm(t, o.a, "wiki:manage")
	lotadoB := h.gestorEm(t, o.b, "contact:read") // lotado em B, sem wiki:manage
	_, estranho := h.user("nexus-user")
	pagina := func(unidade, parent string) string {
		body := `{"title":"Página ` + uuid.NewString()[:8] + `"`
		if unidade != "" {
			body += `,"unidade_id":"` + unidade + `"`
		}
		if parent != "" {
			body += `,"parent_id":"` + parent + `"`
		}
		return body + `}`
	}
	type paginaResp struct {
		ID        string  `json:"id"`
		UnidadeID *string `json:"unidade_id"`
	}

	h.expect(http.StatusForbidden, http.MethodPost, "/api/v1/wiki/pages", estranho, pagina(o.b, ""))
	h.expect(http.StatusForbidden, http.MethodPost, "/api/v1/wiki/pages", gestorA, pagina(o.b, ""))
	deB := data[paginaResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/wiki/pages", lotadoB, pagina(o.b, "")))
	deSub := data[paginaResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/wiki/pages", gestorA, pagina(o.sub, "")))
	filha := data[paginaResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/wiki/pages", estranho, pagina("", deSub.ID)))
	if filha.UnidadeID == nil || *filha.UnidadeID != o.sub {
		t.Fatalf("subpágina herda a dona da mãe: %+v", filha)
	}

	h.expect(http.StatusForbidden, http.MethodDelete, "/api/v1/wiki/pages/"+deB.ID, gestorA, "")
	h.expect(http.StatusNoContent, http.MethodDelete, "/api/v1/wiki/pages/"+filha.ID, gestorA, "") // subunidade de A
	h.expect(http.StatusNoContent, http.MethodDelete, "/api/v1/wiki/pages/"+deSub.ID, gestorA, "")
	institucional := data[paginaResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/wiki/pages", estranho, pagina("", "")))
	h.expect(http.StatusForbidden, http.MethodDelete, "/api/v1/wiki/pages/"+institucional.ID, gestorA, "")
	h.expect(http.StatusNoContent, http.MethodDelete, "/api/v1/wiki/pages/"+institucional.ID, admin, "")
	h.expect(http.StatusNoContent, http.MethodDelete, "/api/v1/wiki/pages/"+deB.ID, admin, "")
}
