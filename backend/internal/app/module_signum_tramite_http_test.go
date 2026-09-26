package app

import (
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
