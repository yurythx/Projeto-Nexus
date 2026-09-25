// Package domain define a Central de Serviços institucionais (Catálogo).
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/yurythx/projeto-nexus/internal/domain/pagination"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
)

var (
	ErrNotFound  = errors.New("catalog: serviço não encontrado")
	ErrSlugTaken = errors.New("catalog: slug já utilizado")
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

// Channel é um canal de atendimento do serviço.
type Channel struct {
	Type  string `json:"type" validate:"required,oneof=online presencial telefone email"`
	Label string `json:"label" validate:"required,max=120"`
	Value string `json:"value" validate:"required,max=300"`
}

// Service é um serviço institucional.
type Service struct {
	ID                   uuid.UUID  `json:"id"`
	Slug                 string     `json:"slug"`
	Title                string     `json:"title"`
	Summary              string     `json:"summary"`
	Description          string     `json:"description"`
	Category             string     `json:"category"`
	Audience             string     `json:"audience"`
	Requirements         []string   `json:"requirements"`
	Steps                []string   `json:"steps"`
	Channels             []Channel  `json:"channels"`
	SLA                  string     `json:"sla"`
	Cost                 string     `json:"cost"`
	Icon                 string     `json:"icon"`
	ResponsibleUnidadeID *uuid.UUID `json:"responsible_unidade_id,omitempty"`
	ResponsibleUnidade   string     `json:"responsible_unidade,omitempty"`
	Status               string     `json:"status"`
	Position             int        `json:"position"`
	PublishedAt          *time.Time `json:"published_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// Category agrega a contagem de serviços publicados.
type Category struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// Filter restringe a listagem.
type Filter struct {
	Status   string // "published" (público) ou "all" (gestão)
	Category string
	Query    string
}

// Repository é a porta de persistência.
type Repository interface {
	List(ctx context.Context, db database.DBTX, f Filter, p pagination.Params) ([]Service, int64, error)
	Get(ctx context.Context, db database.DBTX, id uuid.UUID) (Service, error)
	GetBySlug(ctx context.Context, db database.DBTX, slug string) (Service, error)
	Save(ctx context.Context, db database.DBTX, s Service) (Service, error)
	Delete(ctx context.Context, db database.DBTX, id uuid.UUID) error
	Categories(ctx context.Context, db database.DBTX) ([]Category, error)
	Search(ctx context.Context, db database.DBTX, query string, limit int) ([]Service, []float64, error)
}
