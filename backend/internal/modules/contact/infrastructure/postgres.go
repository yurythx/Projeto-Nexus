// Package infrastructure implementa o repositório do Contato.
package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/yurythx/projeto-nexus/internal/domain/pagination"
	"github.com/yurythx/projeto-nexus/internal/modules/contact/domain"
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
	}
	return fmt.Errorf("contact: %w", err)
}

func (r *Repository) ActiveUser(ctx context.Context, db database.DBTX, id uuid.UUID) (bool, error) {
	var ok bool
	err := db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND active)`, id).Scan(&ok)
	return ok, wrap(err)
}

// Insert grava a mensagem gerando o protocolo "CT-AAAA-NNNNNN".
func (r *Repository) Insert(ctx context.Context, db database.DBTX, m domain.Message) (domain.Message, error) {
	err := db.QueryRow(ctx, `
		INSERT INTO contact_messages (id, protocol, name, email, phone, subject, category, message, consent_at, ip_address, user_agent)
		VALUES ($1, 'CT-' || to_char(now(), 'YYYY') || '-' || lpad(nextval('contact_protocol_seq')::text, 6, '0'),
		        $2, $3, $4, $5, $6, $7, $8, NULLIF($9,'')::inet, $10)
		RETURNING protocol, status, created_at, updated_at`,
		m.ID, m.Name, m.Email, m.Phone, m.Subject, m.Category, m.Message, m.ConsentAt, m.IPAddress, m.UserAgent).
		Scan(&m.Protocol, &m.Status, &m.CreatedAt, &m.UpdatedAt)
	return m, wrap(err)
}

func (r *Repository) List(ctx context.Context, db database.DBTX, status string, p pagination.Params) ([]domain.Summary, int64, error) {
	var total int64
	if err := db.QueryRow(ctx, `SELECT count(*) FROM contact_messages WHERE ($1 = '' OR status = $1)`, status).Scan(&total); err != nil {
		return nil, 0, wrap(err)
	}
	rows, err := db.Query(ctx, `
		SELECT id, protocol, name, email, subject, category, status, created_at
		FROM contact_messages WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, status, p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, wrap(err)
	}
	defer rows.Close()
	out := []domain.Summary{}
	for rows.Next() {
		var s domain.Summary
		if err := rows.Scan(&s.ID, &s.Protocol, &s.Name, &s.Email, &s.Subject, &s.Category, &s.Status, &s.CreatedAt); err != nil {
			return nil, 0, wrap(err)
		}
		out = append(out, s)
	}
	return out, total, rows.Err()
}

func (r *Repository) Get(ctx context.Context, db database.DBTX, id uuid.UUID) (domain.Message, error) {
	var m domain.Message
	var consent time.Time
	err := db.QueryRow(ctx, `
		SELECT id, protocol, name, email, phone, subject, category, message, consent_at, COALESCE(host(ip_address),''),
		       user_agent, status, assigned_to, notes, created_at, updated_at
		FROM contact_messages WHERE id = $1`, id).
		Scan(&m.ID, &m.Protocol, &m.Name, &m.Email, &m.Phone, &m.Subject, &m.Category, &m.Message, &consent,
			&m.IPAddress, &m.UserAgent, &m.Status, &m.AssignedTo, &m.Notes, &m.CreatedAt, &m.UpdatedAt)
	m.ConsentAt = consent
	return m, wrap(err)
}

func (r *Repository) UpdateTriage(ctx context.Context, db database.DBTX, id uuid.UUID, status, notes string, assignedTo *uuid.UUID) error {
	tag, err := db.Exec(ctx, `UPDATE contact_messages SET status = $2, notes = $3, assigned_to = $4 WHERE id = $1`,
		id, status, notes, assignedTo)
	if err == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return wrap(err)
}
