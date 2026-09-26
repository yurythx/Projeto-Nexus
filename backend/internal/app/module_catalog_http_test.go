package app

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Carta de Serviços: publicar exige resumo e canal (Lei 13.460/2017);
// mudar para o mesmo estado não regrava; publicado não se exclui; slug e
// unidade responsável validados; busca global só com publicados.
func TestCatalogPublicationRules(t *testing.T) {
	h := newHarness(t)
	admin := h.admin()
	sfx := uuid.NewString()[:6]
	create := func(want int, body string) postResp {
		return data[postResp](t, h.expect(want, http.MethodPost, "/api/v1/catalog/admin/services", admin, body))
	}
	incompleto := create(http.StatusCreated, `{"title":"Passaporte `+sfx+`"}`)
	h.expect(http.StatusUnprocessableEntity, http.MethodPost, "/api/v1/catalog/admin/services/"+incompleto.ID+"/publish", admin, "")
	full := `{"title":"Passaporte ` + sfx + `","summary":"Emissão do passaporte","channels":[{"type":"presencial","label":"Posto","value":"Rua A, 1"}]}`
	h.expect(http.StatusOK, http.MethodPut, "/api/v1/catalog/admin/services/"+incompleto.ID, admin, full)
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/catalog/admin/services/"+incompleto.ID+"/publish", admin, "")
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/catalog/admin/services/"+incompleto.ID+"/publish", admin, "") // já publicado
	if outboxCount(t, h, "catalog.service.published", incompleto.ID) != 1 {
		t.Fatal("republicar não emite o evento de novo")
	}
	// Publicado continua precisando do mínimo: tirar o canal é recusado.
	h.expect(http.StatusUnprocessableEntity, http.MethodPut, "/api/v1/catalog/admin/services/"+incompleto.ID, admin,
		`{"title":"Passaporte `+sfx+`","summary":"Emissão do passaporte","channels":[]}`)
	h.expect(http.StatusConflict, http.MethodDelete, "/api/v1/catalog/admin/services/"+incompleto.ID, admin, "")

	// Busca global (anônima) encontra o publicado.
	if res := h.expect(http.StatusOK, http.MethodGet, "/api/v1/search?q=Passaporte+"+sfx+"&module=catalog", admin, "").Body.String(); !strings.Contains(res, incompleto.ID) {
		t.Fatalf("serviço publicado na busca global: %s", res)
	}
	// Lista de gestão por estado e texto.
	if l := h.expect(http.StatusOK, http.MethodGet, "/api/v1/catalog/admin/services?status=published&q="+sfx, admin, "").Body.String(); !strings.Contains(l, incompleto.ID) {
		t.Fatalf("lista de gestão filtrada: %s", l)
	}
	h.expect(http.StatusOK, http.MethodGet, "/api/v1/catalog/admin/services/"+incompleto.ID, admin, "")
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/catalog/admin/services/"+incompleto.ID+"/unpublish", admin, "")
	h.expect(http.StatusNoContent, http.MethodDelete, "/api/v1/catalog/admin/services/"+incompleto.ID, admin, "")

	// Slug e unidade responsável.
	create(http.StatusUnprocessableEntity, `{"title":"!!!"}`)
	create(http.StatusUnprocessableEntity, `{"title":"Com unidade","responsible_unidade_id":"`+uuid.NewString()+`"}`)
	dup := create(http.StatusCreated, `{"title":"Duplicado `+sfx+`"}`)
	create(http.StatusConflict, `{"title":"Outro","slug":"`+dup.Slug+`"}`)
	for _, c := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/v1/catalog/admin/services/x", ""},
		{http.MethodPut, "/api/v1/catalog/admin/services/x", `{"title":"x"}`},
		{http.MethodPost, "/api/v1/catalog/admin/services/x/publish", ""},
		{http.MethodDelete, "/api/v1/catalog/admin/services/x", ""},
		{http.MethodPost, "/api/v1/catalog/admin/services", `{`},
	} {
		h.expect(http.StatusBadRequest, c.method, c.path, admin, c.body)
	}
	for _, c := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/v1/catalog/admin/services/" + uuid.NewString(), ""},
		{http.MethodPut, "/api/v1/catalog/admin/services/" + uuid.NewString(), `{"title":"Fantasma"}`},
		{http.MethodPost, "/api/v1/catalog/admin/services/" + uuid.NewString() + "/archive", ""},
		{http.MethodDelete, "/api/v1/catalog/admin/services/" + uuid.NewString(), ""},
		{http.MethodGet, "/api/v1/catalog/services/nao-existe-" + sfx, ""},
	} {
		h.expect(http.StatusNotFound, c.method, c.path, admin, c.body)
	}
}
