// Package application contém os casos de uso do Mercúrio. A mensagem é
// gravada no Postgres (durável) e, após o commit, publicada no Hub —
// que a distribui a todas as réplicas da API pelo backplane Redis.
package application

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/yurythx/projeto-nexus/internal/domain/errors"
	"github.com/yurythx/projeto-nexus/internal/modules/mercurio/domain"
	"github.com/yurythx/projeto-nexus/internal/platform/audit"
	"github.com/yurythx/projeto-nexus/internal/platform/auth"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
	"github.com/yurythx/projeto-nexus/internal/platform/outbox"
	"github.com/yurythx/projeto-nexus/internal/platform/ws"
)

// EventMessageCreated é emitido (sem o corpo da mensagem) a cada envio.
const EventMessageCreated = "mercurio.message.created"

// RoomTopic é o tópico WebSocket de uma sala.
func RoomTopic(id uuid.UUID) string { return "mercurio:room:" + id.String() }

// Service implementa os casos de uso.
type Service struct {
	pool   *pgxpool.Pool
	repo   domain.Repository
	hub    *ws.Hub
	outbox *outbox.Writer
	logger *slog.Logger
}

// NewService cria o serviço.
func NewService(pool *pgxpool.Pool, repo domain.Repository, hub *ws.Hub, ob *outbox.Writer, logger *slog.Logger) *Service {
	return &Service{pool: pool, repo: repo, hub: hub, outbox: ob, logger: logger}
}

// MapError traduz erros de domínio.
func MapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return apperrors.NotFound("sala ou mensagem não encontrada")
	case errors.Is(err, domain.ErrForbidden):
		return apperrors.Forbidden("você não participa desta sala")
	case errors.Is(err, domain.ErrEditWindow):
		return apperrors.Conflict(err.Error())
	}
	return err
}

// Rooms lista as salas acessíveis a identity.
func (s *Service) Rooms(ctx context.Context, identity auth.Identity) ([]domain.Room, error) {
	manage := auth.HasPermission(identity, auth.PermMercurioManage)
	all, err := s.repo.Rooms(ctx, s.pool, identity.UserID, manage)
	if err != nil {
		return nil, err
	}
	out := []domain.Room{}
	for _, r := range all {
		if !domain.CanAccess(identity, r) {
			continue
		}
		if r.Kind == "direct" {
			// Conversa direta exibe o nome do outro participante.
			for _, m := range r.Members {
				if m != identity.UserID {
					if name, err := s.repo.UserName(ctx, s.pool, m); err == nil {
						r.Name = name
					}
				}
			}
		}
		out = append(out, r)
	}
	return out, nil
}

func (s *Service) room(ctx context.Context, identity auth.Identity, id uuid.UUID) (domain.Room, error) {
	r, err := s.repo.Room(ctx, s.pool, id)
	if err != nil {
		return r, err
	}
	if !domain.CanAccess(identity, r) {
		return domain.Room{}, domain.ErrForbidden
	}
	return r, nil
}

// RoomInput cria/edita uma sala (mercurio:manage).
type RoomInput struct {
	Kind           string
	Name           string
	Description    string
	ADGroup        string
	DepartamentoID *uuid.UUID
	Archived       bool
}

// SaveRoom cria (id nil) ou atualiza uma sala global/departamental.
func (s *Service) SaveRoom(ctx context.Context, identity auth.Identity, id uuid.UUID, in RoomInput) (domain.Room, error) {
	if in.Kind != "global" && in.Kind != "department" {
		return domain.Room{}, apperrors.Validation("tipo de sala deve ser global ou department")
	}
	if in.Kind == "department" && strings.TrimSpace(in.ADGroup) == "" && in.DepartamentoID == nil {
		return domain.Room{}, apperrors.Validation("sala departamental precisa de grupo do AD ou departamento")
	}
	var out domain.Room
	err := database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var before any
		room := domain.Room{ID: id, Kind: in.Kind, Name: strings.TrimSpace(in.Name), Description: in.Description,
			ADGroup: strings.TrimSpace(in.ADGroup), DepartamentoID: in.DepartamentoID, Archived: in.Archived}
		if id == uuid.Nil {
			room.ID = uuid.New()
		} else {
			prev, err := s.repo.Room(ctx, tx, id)
			if err != nil {
				return err
			}
			if prev.Kind == "direct" {
				return apperrors.Conflict("salas diretas não são editáveis")
			}
			room.Kind, before = prev.Kind, prev
		}
		var err error
		if out, err = s.repo.SaveRoom(ctx, tx, room, identity.UserID); err != nil {
			if database.IsForeignKeyViolation(err) {
				return apperrors.Validation("departamento inexistente")
			}
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "mercurio.room.saved", "mercurio_room", out.ID.String(), before, out))
	})
	if err == nil && out.Archived {
		s.hub.DropTopic(RoomTopic(out.ID))
	}
	return out, MapError(err)
}

// Direct obtém/cria a conversa direta com outro usuário.
func (s *Service) Direct(ctx context.Context, identity auth.Identity, other uuid.UUID) (domain.Room, error) {
	if other == identity.UserID {
		return domain.Room{}, apperrors.Validation("não é possível abrir conversa consigo mesmo")
	}
	name, err := s.repo.UserName(ctx, s.pool, other)
	if err != nil {
		return domain.Room{}, MapError(err)
	}
	room, err := s.repo.DirectRoom(ctx, s.pool, identity.UserID, other, "DM")
	if err != nil {
		return domain.Room{}, err
	}
	room.Name = name
	return room, nil
}

// Messages lista mensagens (paginação por cursor "before").
func (s *Service) Messages(ctx context.Context, identity auth.Identity, roomID uuid.UUID, before *time.Time, limit int) ([]domain.Message, error) {
	if _, err := s.room(ctx, identity, roomID); err != nil {
		return nil, MapError(err)
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return s.repo.Messages(ctx, s.pool, roomID, before, limit)
}

// Send grava a mensagem e a distribui em tempo real.
func (s *Service) Send(ctx context.Context, identity auth.Identity, roomID uuid.UUID, body string) (domain.Message, error) {
	body = strings.TrimSpace(body)
	if body == "" || len([]rune(body)) > 4000 {
		return domain.Message{}, apperrors.Validation("a mensagem deve ter de 1 a 4000 caracteres")
	}
	room, err := s.room(ctx, identity, roomID)
	if err != nil {
		return domain.Message{}, MapError(err)
	}
	var out domain.Message
	err = database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		var err error
		if out, err = s.repo.InsertMessage(ctx, tx, domain.Message{ID: uuid.New(), RoomID: roomID, AuthorID: identity.UserID, Body: body}); err != nil {
			return err
		}
		if err := s.repo.MarkRead(ctx, tx, roomID, identity.UserID); err != nil {
			return err
		}
		return s.outbox.Write(ctx, tx, EventMessageCreated, "mercurio_message", out.ID.String(), uuid.Nil, map[string]any{
			"id": out.ID.String(), "room_id": roomID.String(), "room_kind": room.Kind, "author_id": identity.UserID.String(),
		})
	})
	if err != nil {
		return domain.Message{}, MapError(err)
	}
	s.publish(ctx, RoomTopic(roomID), "mercurio.message", out)
	// Salas diretas: aviso no tópico pessoal (badge de não lidas mesmo
	// com a sala fechada).
	if room.Kind == "direct" {
		for _, m := range room.Members {
			if m != identity.UserID {
				s.publish(ctx, ws.UserTopic(m.String()), "mercurio.unread", map[string]string{"room_id": roomID.String()})
			}
		}
	}
	return out, nil
}

// Edit altera a própria mensagem dentro da janela de edição.
func (s *Service) Edit(ctx context.Context, identity auth.Identity, id uuid.UUID, body string) (domain.Message, error) {
	body = strings.TrimSpace(body)
	if body == "" || len([]rune(body)) > 4000 {
		return domain.Message{}, apperrors.Validation("a mensagem deve ter de 1 a 4000 caracteres")
	}
	m, err := s.repo.Message(ctx, s.pool, id)
	if err != nil {
		return m, MapError(err)
	}
	if m.AuthorID != identity.UserID {
		return domain.Message{}, MapError(domain.ErrForbidden)
	}
	if m.Deleted || time.Since(m.CreatedAt) > domain.EditWindow {
		return domain.Message{}, MapError(domain.ErrEditWindow)
	}
	out, err := s.repo.UpdateMessage(ctx, s.pool, id, body, false)
	if err != nil {
		return out, MapError(err)
	}
	s.publish(ctx, RoomTopic(out.RoomID), "mercurio.message.edited", out)
	return out, nil
}

// Delete remove (soft) a mensagem: autor ou mercurio:manage (auditado).
func (s *Service) Delete(ctx context.Context, identity auth.Identity, id uuid.UUID) error {
	m, err := s.repo.Message(ctx, s.pool, id)
	if err != nil {
		return MapError(err)
	}
	moderation := m.AuthorID != identity.UserID
	if moderation && !auth.HasPermission(identity, auth.PermMercurioManage) {
		return MapError(domain.ErrForbidden)
	}
	out, err := s.repo.UpdateMessage(ctx, s.pool, id, "", true)
	if err != nil {
		return MapError(err)
	}
	if moderation {
		_ = audit.NewWriter(s.pool).Record(ctx, audit.Meta(ctx, "mercurio.message.moderated", "mercurio_message", id.String(),
			map[string]string{"author_id": m.AuthorID.String(), "room_id": m.RoomID.String()}, nil))
	}
	s.publish(ctx, RoomTopic(out.RoomID), "mercurio.message.deleted", map[string]string{"id": id.String(), "room_id": out.RoomID.String()})
	return nil
}

// MarkRead zera as não lidas da sala.
func (s *Service) MarkRead(ctx context.Context, identity auth.Identity, roomID uuid.UUID) error {
	if _, err := s.room(ctx, identity, roomID); err != nil {
		return MapError(err)
	}
	return s.repo.MarkRead(ctx, s.pool, roomID, identity.UserID)
}

func (s *Service) publish(ctx context.Context, topic, frameType string, data any) {
	if err := s.hub.Publish(ctx, topic, frameType, data); err != nil {
		s.logger.Warn("mercurio: falha ao publicar em tempo real (mensagem já persistida)", slog.Any("error", err))
	}
}

// Authorize é o autorizador de tópicos WebSocket do Mercúrio.
func (s *Service) Authorize(ctx context.Context, client ws.ClientInfo, topic string) error {
	idStr, ok := strings.CutPrefix(topic, "mercurio:room:")
	if !ok {
		return ws.ErrTopicForbidden
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ws.ErrTopicForbidden
	}
	if _, err := s.room(ctx, client.Identity(), id); err != nil {
		return ws.ErrTopicForbidden
	}
	return nil
}

type typingFrame struct {
	RoomID uuid.UUID `json:"room_id"`
}

// Inbound trata frames "mercurio.*" vindos do navegador (indicador de
// digitação — efêmero, não persiste).
func (s *Service) Inbound(ctx context.Context, client ws.ClientInfo, f ws.Frame) error {
	if f.Type != "mercurio.typing" {
		return errors.New("frame do mercúrio desconhecido")
	}
	var t typingFrame
	if err := json.Unmarshal(f.Data, &t); err != nil {
		return errors.New("frame inválido")
	}
	identity := client.Identity()
	if _, err := s.room(ctx, identity, t.RoomID); err != nil {
		return errors.New("sem acesso à sala")
	}
	return s.hub.Publish(ctx, RoomTopic(t.RoomID), "mercurio.typing", map[string]string{
		"room_id": t.RoomID.String(), "user_id": client.UserID, "username": client.Username,
	})
}
