// Package domain define a Agenda corporativa: eventos e salas.
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/yurythx/projeto-nexus/internal/platform/database"
)

var (
	ErrNotFound     = errors.New("calendar: registro não encontrado")
	ErrRoomConflict = errors.New("calendar: sala já reservada no horário")
	ErrInvalidRange = errors.New("calendar: intervalo inválido")
	ErrDuplicate    = errors.New("calendar: registro duplicado")
)

// MaxRange limita consultas de período (evita varreduras gigantes).
const MaxRange = 93 * 24 * time.Hour

// Room é uma sala reservável.
type Room struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Location  string    `json:"location"`
	Capacity  int       `json:"capacity"`
	Resources []string  `json:"resources"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Event é um evento/reserva.
type Event struct {
	ID            uuid.UUID  `json:"id"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Location      string     `json:"location"`
	RoomID        *uuid.UUID `json:"room_id,omitempty"`
	RoomName      string     `json:"room_name,omitempty"`
	StartsAt      time.Time  `json:"starts_at"`
	EndsAt        time.Time  `json:"ends_at"`
	AllDay        bool       `json:"all_day"`
	Visibility    string     `json:"visibility"`
	Status        string     `json:"status"`
	OrganizerID   uuid.UUID  `json:"organizer_id"`
	OrganizerName string     `json:"organizer_name"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Busy é um intervalo ocupado de uma sala.
type Busy struct {
	EventID  uuid.UUID `json:"event_id"`
	Title    string    `json:"title"`
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
}

// EventFilter restringe a listagem.
type EventFilter struct {
	From, To      time.Time
	RoomID        *uuid.UUID
	ViewerID      uuid.UUID
	SeeAll        bool // calendar:manage enxerga eventos privados de terceiros
	OnlyPublic    bool
	IncludeCancel bool
}

// ValidateRange confere o intervalo de um evento.
func ValidateRange(start, end time.Time) error {
	if !end.After(start) || end.Sub(start) > 31*24*time.Hour {
		return ErrInvalidRange
	}
	return nil
}

// Repository é a porta de persistência.
type Repository interface {
	ListRooms(ctx context.Context, db database.DBTX, onlyActive bool) ([]Room, error)
	GetRoom(ctx context.Context, db database.DBTX, id uuid.UUID) (Room, error)
	SaveRoom(ctx context.Context, db database.DBTX, r Room) (Room, error)
	DeleteRoom(ctx context.Context, db database.DBTX, id uuid.UUID) error
	RoomBusy(ctx context.Context, db database.DBTX, roomID uuid.UUID, from, to time.Time) ([]Busy, error)

	ListEvents(ctx context.Context, db database.DBTX, f EventFilter) ([]Event, error)
	GetEvent(ctx context.Context, db database.DBTX, id uuid.UUID) (Event, error)
	SaveEvent(ctx context.Context, db database.DBTX, e Event) (Event, error)
	DeleteEvent(ctx context.Context, db database.DBTX, id uuid.UUID) error
}
