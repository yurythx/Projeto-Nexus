package app

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// Autoatendimento nunca define a própria lotação (nem no primeiro
// salvamento); a gestão define, com departamento coerente com a unidade;
// perfil oculto; busca e filtros.
func TestDirectoryProfileRules(t *testing.T) {
	h := newHarness(t)
	admin := h.admin()
	sfx := uuid.NewString()[:6]
	ent := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/entidades", admin, `{"nome":"Órgão `+sfx+`"}`))
	un1 := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/unidades", admin, `{"entidade_id":"`+ent.ID+`","nome":"Obras `+sfx+`"}`))
	un2 := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/unidades", admin, `{"entidade_id":"`+ent.ID+`","nome":"Saúde `+sfx+`"}`))
	dep := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/departamentos", admin, `{"unidade_id":"`+un1.ID+`","nome":"Projetos `+sfx+`"}`))
	anaID, ana := h.user("nexus-user")
	_, beto := h.user("nexus-user")

	// Primeiro salvamento da própria ficha: a lotação enviada é ignorada.
	me := h.expect(http.StatusOK, http.MethodPut, "/api/v1/directory/me", ana,
		`{"job_title":"Engenheira `+sfx+`","phone":"61 3333-0000","unidade_id":"`+un2.ID+`","visible":true}`).Body.String()
	if strings.Contains(me, un2.ID) {
		t.Fatalf("autoatendimento não escolhe a própria lotação: %s", me)
	}
	// A gestão define — o departamento precisa ser da unidade.
	put := func(want int, body string) string {
		return h.expect(want, http.MethodPut, "/api/v1/directory/people/"+anaID.String(), admin, body).Body.String()
	}
	put(http.StatusUnprocessableEntity, `{"job_title":"Engenheira","unidade_id":"`+un2.ID+`","departamento_id":"`+dep.ID+`"}`)
	put(http.StatusUnprocessableEntity, `{"job_title":"Engenheira","departamento_id":"`+uuid.NewString()+`"}`)
	put(http.StatusUnprocessableEntity, `{"job_title":"Engenheira","unidade_id":"`+uuid.NewString()+`"}`)
	if got := put(http.StatusOK, `{"job_title":"Engenheira `+sfx+`","departamento_id":"`+dep.ID+`"}`); !strings.Contains(got, un1.ID) {
		t.Fatalf("o departamento completa a unidade: %s", got)
	}
	h.expect(http.StatusForbidden, http.MethodPut, "/api/v1/directory/people/"+anaID.String(), beto, `{"job_title":"x"}`)
	// Depois, a própria pessoa edita a ficha sem perder a lotação da gestão.
	if got := h.expect(http.StatusOK, http.MethodPut, "/api/v1/directory/me", ana, `{"job_title":"Engenheira sênior `+sfx+`","unidade_id":"`+un2.ID+`"}`).Body.String(); !strings.Contains(got, un1.ID) {
		t.Fatalf("autoatendimento preserva a lotação definida pela gestão: %s", got)
	}

	// Busca por cargo, filtros e curingas do LIKE tratados como texto.
	if l := h.expect(http.StatusOK, http.MethodGet, "/api/v1/directory/people?q=s%C3%AAnior+"+sfx, beto, "").Body.String(); !strings.Contains(l, anaID.String()) {
		t.Fatalf("busca por cargo: %s", l)
	}
	if l := h.expect(http.StatusOK, http.MethodGet, "/api/v1/directory/people?unidade_id="+un1.ID+"&departamento_id="+dep.ID, beto, "").Body.String(); !strings.Contains(l, anaID.String()) {
		t.Fatalf("filtros por unidade e departamento: %s", l)
	}
	if l := h.expect(http.StatusOK, http.MethodGet, "/api/v1/directory/people?q=%25%25%25", beto, "").Body.String(); strings.Contains(l, anaID.String()) {
		t.Fatal("% digitado na busca não vira curinga")
	}

	// Ficha oculta: some da busca e da consulta de terceiros.
	h.expect(http.StatusOK, http.MethodPut, "/api/v1/directory/me", ana, `{"job_title":"Engenheira","visible":false}`)
	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/directory/people/"+anaID.String(), beto, "")
	h.expect(http.StatusOK, http.MethodGet, "/api/v1/directory/people/"+anaID.String(), ana, "")
	h.expect(http.StatusOK, http.MethodGet, "/api/v1/directory/people/"+anaID.String(), admin, "")
	if l := h.expect(http.StatusOK, http.MethodGet, "/api/v1/directory/people?q="+sfx, beto, "").Body.String(); strings.Contains(l, anaID.String()) {
		t.Fatal("ficha oculta fora da busca")
	}
	h.expect(http.StatusOK, http.MethodGet, "/api/v1/directory/me", ana, "")
	if s := h.expect(http.StatusOK, http.MethodGet, "/api/v1/directory/public/sectors?q=Obras+"+sfx, "", "").Body.String(); !strings.Contains(s, un1.ID) {
		t.Fatalf("setores públicos: %s", s)
	}
	h.expect(http.StatusOK, http.MethodGet, "/api/v1/directory/sectors?q="+sfx, beto, "")
	for _, c := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/v1/directory/people/x", ""},
		{http.MethodPut, "/api/v1/directory/people/x", `{}`},
		{http.MethodGet, "/api/v1/directory/people?unidade_id=x", ""},
		{http.MethodGet, "/api/v1/directory/people?departamento_id=x", ""},
		{http.MethodPut, "/api/v1/directory/me", `{`},
		{http.MethodPut, "/api/v1/directory/people/" + anaID.String(), `{`},
	} {
		h.expect(http.StatusBadRequest, c.method, c.path, admin, c.body)
	}
	h.expect(http.StatusNotFound, http.MethodPut, "/api/v1/directory/people/"+uuid.NewString(), admin, `{"job_title":"Fantasma"}`)
	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/directory/people/"+uuid.NewString(), admin, "")
}
