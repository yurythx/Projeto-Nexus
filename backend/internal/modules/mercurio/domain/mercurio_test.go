package domain

import (
	"testing"

	"github.com/google/uuid"

	"github.com/yurythx/projeto-nexus/internal/platform/auth"
)

func TestCanAccess(t *testing.T) {
	me, other := uuid.New(), uuid.New()
	dep := uuid.New()
	identity := auth.Identity{UserID: me, Groups: []string{"/Nexus/Financeiro"}, Scopes: []auth.Scope{{DepartamentoID: &dep}}}

	cases := []struct {
		name string
		room Room
		want bool
	}{
		{"global", Room{Kind: "global"}, true},
		{"global arquivada", Room{Kind: "global", Archived: true}, false},
		{"departamental por grupo do AD", Room{Kind: "department", ADGroup: "financeiro"}, true},
		{"departamental por lotação", Room{Kind: "department", DepartamentoID: &dep}, true},
		{"departamental de outro grupo", Room{Kind: "department", ADGroup: "juridico"}, false},
		{"direta participante", Room{Kind: "direct", Members: []uuid.UUID{me, other}}, true},
		{"direta alheia", Room{Kind: "direct", Members: []uuid.UUID{other, uuid.New()}}, false},
	}
	for _, c := range cases {
		if got := CanAccess(identity, c.room); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
	mod := auth.Identity{Permissions: []string{"mercurio:manage"}}
	if CanAccess(mod, Room{Kind: "direct", Members: []uuid.UUID{me, other}}) {
		t.Error("moderação não lê conversas diretas alheias")
	}
	if !CanAccess(mod, Room{Kind: "department", ADGroup: "juridico"}) {
		t.Error("moderação acessa salas departamentais")
	}
}

func TestDMKeyIsSymmetric(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	if DMKey(a, b) != DMKey(b, a) {
		t.Fatal("chave de sala direta deve independer da ordem")
	}
}
