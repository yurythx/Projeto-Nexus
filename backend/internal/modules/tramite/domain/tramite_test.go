package domain

import (
	"testing"

	"github.com/google/uuid"

	"github.com/yurythx/projeto-nexus/internal/platform/auth"
)

func TestFormatNumero(t *testing.T) {
	if got := FormatNumero(42, 2026); got != "000042/2026" {
		t.Fatalf("got %q", got)
	}
}

func TestSigiloRules(t *testing.T) {
	unidadeA, unidadeB := uuid.New(), uuid.New()
	servidorA := auth.Identity{UserID: uuid.New(), Scopes: []auth.Scope{{Perfil: "servidor", UnidadeID: &unidadeA}}}
	gestor := auth.Identity{UserID: uuid.New(), Permissions: []string{"tramite:manage"}}
	estranho := auth.Identity{UserID: uuid.New()}

	base := AccessInfo{CreatedBy: uuid.New(), UnidadeOrigemID: unidadeB, UnidadeAtualID: unidadeA}

	pub := base
	pub.Sigilo = SigiloPublico
	if !CanRead(estranho, pub) {
		t.Error("processo público é legível por qualquer autenticado")
	}
	if CanAct(estranho, pub) {
		t.Error("ler não implica poder movimentar: só a unidade atual age")
	}
	if !CanAct(servidorA, pub) {
		t.Error("lotado na unidade atual deve poder movimentar")
	}

	res := base
	res.Sigilo = SigiloRestrito
	if CanRead(estranho, res) {
		t.Error("restrito não é legível por quem não está nas unidades")
	}
	if !CanRead(servidorA, res) || !CanRead(gestor, res) {
		t.Error("restrito é legível pela unidade atual e por tramite:manage")
	}

	sig := base
	sig.Sigilo = SigiloSigiloso
	if CanRead(servidorA, sig) || CanRead(gestor, sig) {
		t.Error("sigiloso exige credencial explícita, mesmo para a unidade atual e tramite:manage")
	}
	sig.ExplicitGrant = true
	if !CanRead(estranho, sig) {
		t.Error("credencial explícita dá acesso ao sigiloso")
	}
	autor := base
	autor.Sigilo = SigiloSigiloso
	if !CanRead(auth.Identity{UserID: base.CreatedBy}, autor) {
		t.Error("o autor sempre acessa o próprio processo")
	}
}
