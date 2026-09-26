package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
)

const userPassword = "Senha-Forte-123!"

type envelopeResp struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Signers []struct {
		UserID string `json:"user_id"`
		Status string `json:"status"`
	} `json:"signers"`
}

type challengeResp struct {
	ChallengeID string `json:"challenge_id"`
	Nonce       string `json:"nonce"`
}

func sha(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// sign executa a cerimônia completa (desafio + reautenticação).
func (h *apiHarness) sign(token, envID, docHash, password string) *httpRec {
	h.t.Helper()
	c := data[challengeResp](h.t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/signum/envelopes/"+envID+"/challenge", token, ""))
	rec := h.do(http.MethodPost, "/api/v1/signum/envelopes/"+envID+"/sign", token,
		`{"challenge_id":"`+c.ChallengeID+`","nonce":"`+c.Nonce+`","password":"`+password+`","confirm_document_sha256":"`+docHash+`"}`)
	return &httpRec{code: rec.Code, body: rec.Body.String()}
}

type httpRec struct {
	code int
	body string
}

func TestSignumHTTP(t *testing.T) {
	h := newHarness(t)
	_, author := h.user("nexus-user")
	anaID, ana := h.user("nexus-user")
	betoID, beto := h.user("nexus-user")
	_, stranger := h.user("nexus-user")
	doc := sha("contrato " + uuid.NewString())

	h.expect(http.StatusUnprocessableEntity, http.MethodPost, "/api/v1/signum/envelopes", author,
		`{"title":"x","document_sha256":"curto","signer_ids":["`+anaID.String()+`"]}`)
	env := data[envelopeResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/signum/envelopes", author,
		`{"title":"Contrato","document_sha256":"`+doc+`","signer_ids":["`+anaID.String()+`","`+betoID.String()+`"],"sequential":true}`))
	if env.Status != "pending" || len(env.Signers) != 2 {
		t.Fatalf("envelope: %+v", env)
	}

	// Estranhos não veem; signatários veem.
	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/signum/envelopes/"+env.ID, stranger, "")
	h.expect(http.StatusOK, http.MethodGet, "/api/v1/signum/envelopes/"+env.ID, beto, "")
	if list := h.expect(http.StatusOK, http.MethodGet, "/api/v1/signum/envelopes?role=to_sign", ana, "").Body.String(); !strings.Contains(list, env.ID) {
		t.Fatal("pendente deveria estar na caixa de quem assina")
	}

	// Sequencial: Beto não assina antes da Ana.
	if rec := h.do(http.MethodPost, "/api/v1/signum/envelopes/"+env.ID+"/challenge", beto, ""); rec.Code < 400 {
		t.Fatalf("fora de ordem deveria ser recusado: %d", rec.Code)
	}
	// Senha errada e documento divergente são recusados.
	if r := h.sign(ana, env.ID, doc, "senha-errada"); r.code < 400 {
		t.Fatalf("reautenticação com senha errada deveria falhar: %d", r.code)
	}
	if r := h.sign(ana, env.ID, sha("outro documento"), userPassword); r.code < 400 {
		t.Fatalf("documento divergente deveria falhar: %d", r.code)
	}
	// Nonce é de uso único.
	c := data[challengeResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/signum/envelopes/"+env.ID+"/challenge", ana, ""))
	body := `{"challenge_id":"` + c.ChallengeID + `","nonce":"` + c.Nonce + `","password":"` + userPassword + `","confirm_document_sha256":"` + doc + `"}`
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/signum/envelopes/"+env.ID+"/sign", ana, body)
	if rec := h.do(http.MethodPost, "/api/v1/signum/envelopes/"+env.ID+"/sign", ana, body); rec.Code < 400 {
		t.Fatalf("reuso do desafio deveria falhar: %d", rec.Code)
	}

	if r := h.sign(beto, env.ID, doc, userPassword); r.code != http.StatusOK || !strings.Contains(r.body, `"status":"completed"`) {
		t.Fatalf("segunda assinatura conclui o envelope: %d %s", r.code, r.body)
	}
	if outboxCount(t, h, "signum.envelope.completed", env.ID) != 1 {
		t.Fatal("conclusão deveria emitir signum.envelope.completed")
	}

	// Verificação pública (anônima): assinaturas válidas e conferência do arquivo.
	pub := h.expect(http.StatusOK, http.MethodGet, "/api/v1/signum/verify/"+env.ID+"?document_sha256="+doc, "", "").Body.String()
	if !strings.Contains(pub, `"document_match":true`) || strings.Contains(pub, `"valid":false`) {
		t.Fatalf("verificação pública: %s", pub)
	}
	if bad := h.expect(http.StatusOK, http.MethodGet, "/api/v1/signum/verify/"+env.ID+"?document_sha256="+sha("x"), "", "").Body.String(); !strings.Contains(bad, `"document_match":false`) {
		t.Fatalf("arquivo diferente não confere: %s", bad)
	}

	// Recusa com motivo encerra; cancelamento só pelo autor.
	env2 := data[envelopeResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/signum/envelopes", author,
		`{"title":"Aditivo","document_sha256":"`+doc+`","signer_ids":["`+anaID.String()+`"]}`))
	h.expect(http.StatusUnprocessableEntity, http.MethodPost, "/api/v1/signum/envelopes/"+env2.ID+"/refuse", ana, `{"reason":"x"}`)
	ref := data[envelopeResp](t, h.expect(http.StatusOK, http.MethodPost, "/api/v1/signum/envelopes/"+env2.ID+"/refuse", ana, `{"reason":"cláusula errada"}`))
	if ref.Status != "refused" {
		t.Fatalf("recusado: %+v", ref)
	}
	env3 := data[envelopeResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/signum/envelopes", author,
		`{"title":"Termo","document_sha256":"`+doc+`","signer_ids":["`+anaID.String()+`"]}`))
	if rec := h.do(http.MethodPost, "/api/v1/signum/envelopes/"+env3.ID+"/cancel", ana, ""); rec.Code < 400 {
		t.Fatalf("signatário não cancela: %d", rec.Code)
	}
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/signum/envelopes/"+env3.ID+"/cancel", author, "")
	if rec := h.do(http.MethodPost, "/api/v1/signum/envelopes/"+env3.ID+"/challenge", ana, ""); rec.Code < 400 {
		t.Fatalf("envelope cancelado não aceita assinatura: %d", rec.Code)
	}
}

// Trâmite: lotação por unidade, sigilo, despachos e integração com o Signum.
func TestTramiteHTTP(t *testing.T) {
	h := newHarness(t)
	admin := h.admin()
	ctx := context.Background()
	sfx := uuid.NewString()[:6]

	ent := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/entidades", admin, `{"nome":"Órgão `+sfx+`"}`))
	unA := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/unidades", admin, `{"entidade_id":"`+ent.ID+`","nome":"Protocolo `+sfx+`","sigla":"PA`+sfx+`"}`))
	unB := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/iam/unidades", admin, `{"entidade_id":"`+ent.ID+`","nome":"Jurídico `+sfx+`","sigla":"JU`+sfx+`"}`))
	var protocolo, tipo string
	if err := h.d.DB.QueryRow(ctx, `SELECT id FROM perfis WHERE slug = 'protocolo'`).Scan(&protocolo); err != nil {
		t.Fatal(err)
	}
	if err := h.d.DB.QueryRow(ctx, `SELECT id FROM tramite_tipos ORDER BY slug LIMIT 1`).Scan(&tipo); err != nil {
		t.Fatal(err)
	}
	lotar := func(unidade string) (uuid.UUID, string) {
		id, tok := h.user("nexus-user")
		h.expect(http.StatusCreated, http.MethodPost, "/api/v1/users/"+id.String()+"/lotacoes", admin, `{"perfil_id":"`+protocolo+`","unidade_id":"`+unidade+`"}`)
		return id, tok
	}
	_, anaA := lotar(unA.ID)
	betoID, betoB := lotar(unB.ID)
	carlosID, carlosA := lotar(unA.ID)
	_, semLotacao := h.user("nexus-user")

	h.expect(http.StatusForbidden, http.MethodPost, "/api/v1/tramite/processos", semLotacao,
		`{"tipo_id":"`+tipo+`","assunto":"x","sigilo":"publico","unidade_origem_id":"`+unA.ID+`"}`)
	h.expect(http.StatusForbidden, http.MethodPost, "/api/v1/tramite/processos", anaA,
		`{"tipo_id":"`+tipo+`","assunto":"Unidade alheia","sigilo":"publico","unidade_origem_id":"`+unB.ID+`"}`)

	proc := data[struct {
		ID     string `json:"id"`
		Numero string `json:"numero"`
		Status string `json:"status"`
	}](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/tramite/processos", anaA,
		`{"tipo_id":"`+tipo+`","assunto":"Pedido `+sfx+`","interessado":"Fulano","sigilo":"sigiloso","unidade_origem_id":"`+unA.ID+`"}`))
	if proc.Numero == "" || proc.Status != "aberto" {
		t.Fatalf("processo numerado e aberto: %+v", proc)
	}

	// Sigiloso: nem colega da mesma unidade vê sem credencial nominal — e a
	// resposta é 404, sem revelar que o processo existe.
	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/tramite/processos/"+proc.ID, carlosA, "")
	if list := h.expect(http.StatusOK, http.MethodGet, "/api/v1/tramite/processos?page_size=100", carlosA, "").Body.String(); strings.Contains(list, proc.ID) {
		t.Fatal("processo sigiloso não pode aparecer na listagem de quem não tem credencial")
	}
	h.expect(http.StatusNoContent, http.MethodPost, "/api/v1/tramite/processos/"+proc.ID+"/acessos", anaA, `{"user_id":"`+carlosID.String()+`"}`)
	h.expect(http.StatusOK, http.MethodGet, "/api/v1/tramite/processos/"+proc.ID, carlosA, "")

	// Documento redigido -> assinatura via Signum (porta entre plugins).
	doc := data[struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/tramite/processos/"+proc.ID+"/documentos", anaA,
		`{"tipo":"Despacho","titulo":"Parecer","conteudo":"# Parecer\nDeferido."}`))
	h.expect(http.StatusOK, http.MethodPut, "/api/v1/tramite/documentos/"+doc.ID, anaA, `{"titulo":"Parecer final","conteudo":"Deferido."}`)
	signed := h.expect(http.StatusOK, http.MethodPost, "/api/v1/tramite/documentos/"+doc.ID+"/assinatura", anaA,
		`{"signer_ids":["`+carlosID.String()+`"]}`).Body.String()
	if !strings.Contains(signed, `"aguardando_assinatura"`) || !strings.Contains(signed, `"envelope_id"`) {
		t.Fatalf("documento aguardando assinatura com envelope: %s", signed)
	}
	// Em assinatura, o conteúdo fica congelado.
	h.expect(http.StatusConflict, http.MethodPut, "/api/v1/tramite/documentos/"+doc.ID, anaA, `{"titulo":"Alterado","conteudo":"x"}`)

	// Tramitar exige tramite:route e unidade atual; após tramitar, a origem perde a ação.
	h.expect(http.StatusUnprocessableEntity, http.MethodPost, "/api/v1/tramite/processos/"+proc.ID+"/tramitar", anaA, `{"para_unidade_id":"`+unB.ID+`","despacho":"x"}`)
	// Sigiloso: estar lotado no destino não basta — a origem credencia o
	// destinatário ANTES de tramitar (depois ela não atua mais).
	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/tramite/processos/"+proc.ID, betoB, "")
	h.expect(http.StatusNoContent, http.MethodPost, "/api/v1/tramite/processos/"+proc.ID+"/acessos", anaA, `{"user_id":"`+betoID.String()+`"}`)
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/tramite/processos/"+proc.ID+"/tramitar", anaA, `{"para_unidade_id":"`+unB.ID+`","despacho":"Encaminho ao jurídico"}`)
	h.expect(http.StatusForbidden, http.MethodPost, "/api/v1/tramite/processos/"+proc.ID+"/acessos", anaA, `{"user_id":"`+betoID.String()+`"}`)
	view := h.expect(http.StatusOK, http.MethodGet, "/api/v1/tramite/processos/"+proc.ID, betoB, "").Body.String()
	if !strings.Contains(view, `"em_tramitacao"`) || !strings.Contains(view, `"can_act":true`) || !strings.Contains(view, "Encaminho ao jurídico") {
		t.Fatalf("unidade de destino atua no processo: %s", view)
	}
	if rec := h.do(http.MethodPost, "/api/v1/tramite/processos/"+proc.ID+"/concluir", anaA, `{"despacho":"Tentativa da origem"}`); rec.Code < 400 {
		t.Fatalf("unidade de origem não conclui após tramitar: %d", rec.Code)
	}
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/tramite/processos/"+proc.ID+"/concluir", betoB, `{"despacho":"Concluído"}`)
	// Sigiloso é estritamente nominal: nem tramite:manage (admin) lê.
	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/tramite/processos/"+proc.ID, admin, "")

	// Restrito: unidades envolvidas leem; reabrir exige tramite:manage.
	rest := data[idResp](t, h.expect(http.StatusCreated, http.MethodPost, "/api/v1/tramite/processos", anaA,
		`{"tipo_id":"`+tipo+`","assunto":"Restrito `+sfx+`","sigilo":"restrito","unidade_origem_id":"`+unA.ID+`"}`))
	h.expect(http.StatusOK, http.MethodGet, "/api/v1/tramite/processos/"+rest.ID, carlosA, "")
	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/tramite/processos/"+rest.ID, betoB, "")
	h.expect(http.StatusConflict, http.MethodPost, "/api/v1/tramite/processos/"+rest.ID+"/arquivar", anaA, `{"despacho":"Cedo demais"}`)
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/tramite/processos/"+rest.ID+"/concluir", anaA, `{"despacho":"Concluído"}`)
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/tramite/processos/"+rest.ID+"/arquivar", anaA, `{"despacho":"Arquivado"}`)
	h.expect(http.StatusForbidden, http.MethodPost, "/api/v1/tramite/processos/"+rest.ID+"/reabrir", anaA, `{"despacho":"Reabrir"}`)
	h.expect(http.StatusOK, http.MethodPost, "/api/v1/tramite/processos/"+rest.ID+"/reabrir", admin, `{"despacho":"Reaberto pela gestão"}`)
	h.expect(http.StatusConflict, http.MethodPost, "/api/v1/tramite/processos/"+rest.ID+"/reabrir", admin, `{"despacho":"De novo"}`)

	// Sem o Signum, pedir assinatura falha com mensagem clara (e o Trâmite
	// precisa ser desligado antes — dependência declarada).
	h.setModule(admin, "signum", false)
	h.expect(http.StatusNotFound, http.MethodGet, "/api/v1/tramite/processos/"+proc.ID, betoB, "")
}
