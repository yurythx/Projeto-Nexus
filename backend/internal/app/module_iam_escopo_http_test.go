package app

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// relogin entra de novo com a conta local (o token do login local leva os
// grupos da conta, como o do Keycloak leva os do AD).
func (h *apiHarness) relogin(id uuid.UUID) string {
	h.t.Helper()
	var username string
	if err := h.d.DB.QueryRow(context.Background(), `SELECT username FROM users WHERE id = $1`, id).Scan(&username); err != nil {
		h.t.Fatal(err)
	}
	rec := h.do(http.MethodPost, "/api/v1/auth/login", "", `{"username":"`+username+`","password":"Senha-Forte-123!"}`)
	if rec.Code != http.StatusOK {
		h.t.Fatalf("login %s: %d %s", username, rec.Code, rec.Body.String())
	}
	var out struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	return out.Data.AccessToken
}

// Estrutura desativada deixa de conceder (migration 000126,
// docs/REGRAS_DE_NEGOCIO.md §2): lotações e mapeamentos de grupo num
// escopo com entidade, unidade ou departamento desativado não valem — e a
// mudança é imediata (salvar a estrutura invalida o cache do IAM).
func TestIAMEscopoDesativadoNaoConcede(t *testing.T) {
	h := newHarness(t)
	// blog:manage vale com escopo (ADR 013): aparece em /me enquanto o
	// escopo da concessão estiver ativo.
	quer := func(tok string, want bool) {
		t.Helper()
		var me struct {
			Data struct {
				Permissions []string `json:"permissions"`
			} `json:"data"`
		}
		_ = json.Unmarshal(h.expect(http.StatusOK, http.MethodGet, "/api/v1/me", tok, "").Body.Bytes(), &me)
		if got := slices.Contains(me.Data.Permissions, "blog:manage"); got != want {
			t.Fatalf("blog:manage em /me = %v, esperado %v (%v)", got, want, me.Data.Permissions)
		}
	}
	admin := h.admin()
	sfx := uuid.NewString()[:6]
	ent := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/entidades", admin, `{"nome":"Órgão `+sfx+`"}`))
	un := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/unidades", admin,
		`{"entidade_id":"`+ent.ID+`","nome":"Unidade `+sfx+`"}`))
	dep := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/departamentos", admin,
		`{"unidade_id":"`+un.ID+`","nome":"Depto `+sfx+`"}`))
	perfil := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/perfis", admin,
		`{"nome":"Editor `+sfx+`","permissoes":["blog:manage"]}`))
	setAtivo := func(kind, id, body string) {
		t.Helper()
		h.expect(http.StatusOK, http.MethodPut, "/api/v1/iam/"+kind+"/"+id, admin, body)
	}

	// Lotação manual no departamento (o escopo é completado até a entidade).
	userID, tok := h.user("nexus-user")
	h.expect(http.StatusCreated, http.MethodPost, "/api/v1/users/"+userID.String()+"/lotacoes", admin,
		`{"perfil_id":"`+perfil.ID+`","departamento_id":"`+dep.ID+`"}`)
	quer(tok, true)

	setAtivo("departamentos", dep.ID, `{"unidade_id":"`+un.ID+`","nome":"Depto `+sfx+`","ativo":false}`)
	quer(tok, false)
	setAtivo("departamentos", dep.ID, `{"unidade_id":"`+un.ID+`","nome":"Depto `+sfx+`","ativo":true}`)
	quer(tok, true)

	// Entidade desativada derruba tudo o que está abaixo dela.
	setAtivo("entidades", ent.ID, `{"nome":"Órgão `+sfx+`","ativo":false}`)
	quer(tok, false)
	setAtivo("entidades", ent.ID, `{"nome":"Órgão `+sfx+`","ativo":true}`)
	quer(tok, true)

	// Mapeamento de grupo com escopo na unidade: mesma regra.
	grupo := "GRP_ESCOPO_" + sfx
	h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/ad-mappings", admin,
		`{"ad_group":"`+grupo+`","perfil_id":"`+perfil.ID+`","unidade_id":"`+un.ID+`"}`)
	membroID, _ := h.user("nexus-user")
	if _, err := h.d.DB.Exec(context.Background(), `UPDATE users SET groups = $2 WHERE id = $1`, membroID, []string{grupo}); err != nil {
		t.Fatal(err)
	}
	membro := h.relogin(membroID)
	quer(membro, true)
	setAtivo("unidades", un.ID, `{"entidade_id":"`+ent.ID+`","nome":"Unidade `+sfx+`","ativo":false}`)
	quer(membro, false)
	quer(tok, false)
	setAtivo("unidades", un.ID, `{"entidade_id":"`+ent.ID+`","nome":"Unidade `+sfx+`","ativo":true}`)
	quer(membro, true)
}

// Permissão de plataforma só vale com concessão global (ADR 013): lotar
// alguém como auditor NUMA unidade não abre a auditoria da plataforma.
func TestIAMPermissaoDePlataformaSoGlobal(t *testing.T) {
	h := newHarness(t)
	admin := h.admin()
	sfx := uuid.NewString()[:6]
	ent := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/entidades", admin, `{"nome":"Órgão `+sfx+`"}`))
	un := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/unidades", admin,
		`{"entidade_id":"`+ent.ID+`","nome":"Unidade `+sfx+`"}`))
	perfil := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/perfis", admin,
		`{"nome":"Auditor `+sfx+`","permissoes":["audit:read"]}`))

	userID, tok := h.user("nexus-user")
	lot := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/users/"+userID.String()+"/lotacoes", admin,
		`{"perfil_id":"`+perfil.ID+`","unidade_id":"`+un.ID+`"}`))
	h.expect(http.StatusForbidden, http.MethodGet, "/api/v1/audit/logs", tok, "")
	h.expect(http.StatusNoContent, http.MethodDelete, "/api/v1/users/"+userID.String()+"/lotacoes/"+lot.ID, admin, "")
	h.expect(http.StatusCreated, http.MethodPost, "/api/v1/users/"+userID.String()+"/lotacoes", admin, `{"perfil_id":"`+perfil.ID+`"}`)
	h.expect(http.StatusOK, http.MethodGet, "/api/v1/audit/logs", tok, "")
}

// Excluir a unidade-mãe com subunidades é recusado (antes as subunidades
// perdiam a mãe em silêncio).
func TestIAMUnidadeMaeComSubunidadesNaoEExcluida(t *testing.T) {
	h := newHarness(t)
	admin := h.admin()
	sfx := uuid.NewString()[:6]
	ent := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/entidades", admin, `{"nome":"Rede `+sfx+`"}`))
	mae := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/unidades", admin,
		`{"entidade_id":"`+ent.ID+`","nome":"Sede `+sfx+`"}`))
	filha := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/unidades", admin,
		`{"entidade_id":"`+ent.ID+`","parent_id":"`+mae.ID+`","nome":"Filial `+sfx+`"}`))

	h.expect(http.StatusConflict, http.MethodDelete, "/api/v1/iam/unidades/"+mae.ID, admin, "")
	h.expect(http.StatusNoContent, http.MethodDelete, "/api/v1/iam/unidades/"+filha.ID, admin, "")
	h.expect(http.StatusNoContent, http.MethodDelete, "/api/v1/iam/unidades/"+mae.ID, admin, "")
}
