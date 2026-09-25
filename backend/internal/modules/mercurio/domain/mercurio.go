// Package domain define o chat em tempo real Mercúrio: canais globais,
// salas departamentais (mapeadas a grupos do AD ou lotação) e mensagens
// diretas.
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/yurythx/projeto-nexus/internal/platform/auth"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
)

var (
	ErrNotFound   = errors.New("mercurio: não encontrado")
	ErrForbidden  = errors.New("mercurio: sem acesso a esta sala")
	ErrEditWindow = errors.New("mercurio: a mensagem só pode ser editada nos primeiros 15 minutos")
)

// EditWindow é o prazo para editar uma mensagem.
const EditWindow = 15 * time.Minute

// Room é uma sala.
type Room struct {
	ID             uuid.UUID   `json:"id"`
	Kind           string      `json:"kind"`
	Name           string      `json:"name"`
	Description    string      `json:"description"`
	ADGroup        string      `json:"ad_group,omitempty"`
	DepartamentoID *uuid.UUID  `json:"departamento_id,omitempty"`
	Archived       bool        `json:"archived"`
	Members        []uuid.UUID `json:"members,omitempty"`
	Unread         int         `json:"unread"`
	LastMessageAt  *time.Time  `json:"last_message_at,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
}

// Message é uma mensagem.
type Message struct {
	ID         uuid.UUID  `json:"id"`
	RoomID     uuid.UUID  `json:"room_id"`
	AuthorID   uuid.UUID  `json:"author_id"`
	AuthorName string     `json:"author_name"`
	Body       string     `json:"body"`
	CreatedAt  time.Time  `json:"created_at"`
	EditedAt   *time.Time `json:"edited_at,omitempty"`
	Deleted    bool       `json:"deleted"`
}

// CanAccess aplica a regra de participação:
//   - global: todo autenticado;
//   - department: membro do grupo do AD da sala OU lotado no departamento;
//   - direct: só os dois participantes;
//   - mercurio:manage enxerga tudo (moderação).
func CanAccess(identity auth.Identity, r Room) bool {
	if r.Archived && !auth.HasPermission(identity, auth.PermMercurioManage) {
		return false
	}
	switch r.Kind {
	case "global":
		return true
	case "department":
		if auth.HasPermission(identity, auth.PermMercurioManage) {
			return true
		}
		if r.ADGroup != "" && identity.HasGroup(r.ADGroup) {
			return true
		}
		if r.DepartamentoID != nil {
			for _, s := range identity.Scopes {
				if s.DepartamentoID != nil && *s.DepartamentoID == *r.DepartamentoID {
					return true
				}
			}
		}
		return false
	case "direct":
		for _, m := range r.Members {
			if m == identity.UserID {
				return true
			}
		}
	}
	return false
}

// DMKey é a chave única de uma sala direta entre dois usuários.
func DMKey(a, b uuid.UUID) string {
	if a.String() > b.String() {
		a, b = b, a
	}
	return a.String() + ":" + b.String()
}

// Repository é a porta de persistência.
type Repository interface {
	Rooms(ctx context.Context, db database.DBTX, userID uuid.UUID, includeArchived bool) ([]Room, error)
	Room(ctx context.Context, db database.DBTX, id uuid.UUID) (Room, error)
	SaveRoom(ctx context.Context, db database.DBTX, r Room, createdBy uuid.UUID) (Room, error)
	DirectRoom(ctx context.Context, db database.DBTX, a, b uuid.UUID, name string) (Room, error)
	Messages(ctx context.Context, db database.DBTX, roomID uuid.UUID, before *time.Time, limit int) ([]Message, error)
	InsertMessage(ctx context.Context, db database.DBTX, m Message) (Message, error)
	Message(ctx context.Context, db database.DBTX, id uuid.UUID) (Message, error)
	UpdateMessage(ctx context.Context, db database.DBTX, id uuid.UUID, body string, deleted bool) (Message, error)
	MarkRead(ctx context.Context, db database.DBTX, roomID, userID uuid.UUID) error
	UserName(ctx context.Context, db database.DBTX, id uuid.UUID) (string, error)
}
