// Package infrastructure implementa o repositório da Agenda.
package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yurythx/projeto-nexus/internal/modules/calendar/domain"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
)

// Repository implementa domain.Repository.
type Repository struct{}

// NewRepository cria o repositório.
func NewRepository() *Repository { return &Repository{} }

var _ domain.Repository = (*Repository)(nil)

func wrap(err error) error {
	switch {
	case err == nil:
		return nil
	case database.IsNoRows(err):
		return domain.ErrNotFound
	case database.IsExclusionViolation(err):
		return domain.ErrRoomConflict
	case database.IsUniqueViolation(err):
		return domain.ErrDuplicate
	case database.IsCheckViolation(err):
		return domain.ErrInvalidRange
	}
	return fmt.Errorf("calendar: %w", err)
}

const roomCols = `id, name, location, capacity, resources, active, created_at, updated_at`

func scanRoom(row interface{ Scan(...any) error }) (domain.Room, error) {
	var r domain.Room
	err := row.Scan(&r.ID, &r.Name, &r.Location, &r.Capacity, &r.Resources, &r.Active, &r.CreatedAt, &r.UpdatedAt)
	return r, err
}

func (r *Repository) ListRooms(ctx context.Context, db database.DBTX, onlyActive bool) ([]domain.Room, error) {
	rows, err := db.Query(ctx, `SELECT `+roomCols+` FROM calendar_rooms WHERE (NOT $1 OR active) ORDER BY name`, onlyActive)
	if err != nil {
		return nil, wrap(err)
	}
	defer rows.Close()
	out := []domain.Room{}
	for rows.Next() {
		room, err := scanRoom(rows)
		if err != nil {
			return nil, wrap(err)
		}
		out = append(out, room)
	}
	return out, rows.Err()
}

func (r *Repository) GetRoom(ctx context.Context, db database.DBTX, id uuid.UUID) (domain.Room, error) {
	room, err := scanRoom(db.QueryRow(ctx, `SELECT `+roomCols+` FROM calendar_rooms WHERE id = $1`, id))
	return room, wrap(err)
}

func (r *Repository) SaveRoom(ctx context.Context, db database.DBTX, room domain.Room) (domain.Room, error) {
	out, err := scanRoom(db.QueryRow(ctx, `
		INSERT INTO calendar_rooms (id, name, location, capacity, resources, active) VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, location=EXCLUDED.location, capacity=EXCLUDED.capacity,
		    resources=EXCLUDED.resources, active=EXCLUDED.active
		RETURNING `+roomCols, room.ID, room.Name, room.Location, room.Capacity, room.Resources, room.Active))
	return out, wrap(err)
}

func (r *Repository) DeleteRoom(ctx context.Context, db database.DBTX, id uuid.UUID) error {
	tag, err := db.Exec(ctx, `DELETE FROM calendar_rooms WHERE id = $1`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return wrap(err)
}

func (r *Repository) RoomBusy(ctx context.Context, db database.DBTX, roomID uuid.UUID, from, to time.Time) ([]domain.Busy, error) {
	rows, err := db.Query(ctx, `
		SELECT id, CASE WHEN visibility = 'private' THEN 'Reservado' ELSE title END, starts_at, ends_at
		FROM calendar_events
		WHERE room_id = $1 AND status = 'confirmed' AND tstzrange(starts_at, ends_at, '[)') && tstzrange($2, $3, '[)')
		ORDER BY starts_at`, roomID, from, to)
	if err != nil {
		return nil, wrap(err)
	}
	defer rows.Close()
	out := []domain.Busy{}
	for rows.Next() {
		var b domain.Busy
		if err := rows.Scan(&b.EventID, &b.Title, &b.StartsAt, &b.EndsAt); err != nil {
			return nil, wrap(err)
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

const eventCols = `e.id, e.title, e.description, e.location, e.room_id, COALESCE(r.name,''), e.starts_at, e.ends_at,
	e.all_day, e.visibility, e.status, e.organizer_id, COALESCE(NULLIF(u.display_name,''), u.username, ''),
	e.created_at, e.updated_at`

const eventFrom = ` FROM calendar_events e LEFT JOIN calendar_rooms r ON r.id = e.room_id LEFT JOIN users u ON u.id = e.organizer_id `

func scanEvent(row interface{ Scan(...any) error }) (domain.Event, error) {
	var e domain.Event
	err := row.Scan(&e.ID, &e.Title, &e.Description, &e.Location, &e.RoomID, &e.RoomName, &e.StartsAt, &e.EndsAt,
		&e.AllDay, &e.Visibility, &e.Status, &e.OrganizerID, &e.OrganizerName, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *Repository) ListEvents(ctx context.Context, db database.DBTX, f domain.EventFilter) ([]domain.Event, error) {
	rows, err := db.Query(ctx, `SELECT `+eventCols+eventFrom+`
		WHERE tstzrange(e.starts_at, e.ends_at, '[)') && tstzrange($1, $2, '[)')
		  AND ($3::uuid IS NULL OR e.room_id = $3)
		  AND ($4 OR e.status = 'confirmed')
		  AND CASE
		        WHEN $5 THEN e.visibility = 'public'
		        WHEN $6 THEN true
		        ELSE e.visibility <> 'private' OR e.organizer_id = $7
		      END
		ORDER BY e.starts_at LIMIT 2000`,
		f.From, f.To, f.RoomID, f.IncludeCancel, f.OnlyPublic, f.SeeAll, f.ViewerID)
	if err != nil {
		return nil, wrap(err)
	}
	defer rows.Close()
	out := []domain.Event{}
	for rows.Next() {
		e, err := scanEvent(rows)
		if err != nil {
			return nil, wrap(err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *Repository) GetEvent(ctx context.Context, db database.DBTX, id uuid.UUID) (domain.Event, error) {
	e, err := scanEvent(db.QueryRow(ctx, `SELECT `+eventCols+eventFrom+`WHERE e.id = $1`, id))
	return e, wrap(err)
}

func (r *Repository) SaveEvent(ctx context.Context, db database.DBTX, e domain.Event) (domain.Event, error) {
	_, err := db.Exec(ctx, `
		INSERT INTO calendar_events (id, title, description, location, room_id, starts_at, ends_at, all_day, visibility, status, organizer_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description, location=EXCLUDED.location,
		    room_id=EXCLUDED.room_id, starts_at=EXCLUDED.starts_at, ends_at=EXCLUDED.ends_at, all_day=EXCLUDED.all_day,
		    visibility=EXCLUDED.visibility, status=EXCLUDED.status`,
		e.ID, e.Title, e.Description, e.Location, e.RoomID, e.StartsAt, e.EndsAt, e.AllDay, e.Visibility, e.Status, e.OrganizerID)
	if err != nil {
		return domain.Event{}, wrap(err)
	}
	return r.GetEvent(ctx, db, e.ID)
}

func (r *Repository) DeleteEvent(ctx context.Context, db database.DBTX, id uuid.UUID) error {
	tag, err := db.Exec(ctx, `DELETE FROM calendar_events WHERE id = $1`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return wrap(err)
}
