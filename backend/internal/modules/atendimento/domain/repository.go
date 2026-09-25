package domain

import (
	"context"

	"github.com/google/uuid"
)

// Repository define o contrato de persistência para atendimentos e prontuários socioassistenciais.
type Repository interface {
	Create(ctx context.Context, a *Atendimento) error
	GetByID(ctx context.Context, id uuid.UUID) (*Atendimento, error)
	Update(ctx context.Context, a *Atendimento) error
	List(ctx context.Context, filter FilterParams) ([]*Atendimento, int64, error)
	SearchByCPF(ctx context.Context, cpf string) ([]*Atendimento, error)
	GetStats(ctx context.Context, serviceSlug, unit string) (*AtendimentoStats, error)

	// Centro POP
	CreateProntuario(ctx context.Context, p *CentroPopProntuario) error
	GetProntuarioByID(ctx context.Context, id uuid.UUID) (*CentroPopProntuario, error)
	ListProntuarios(ctx context.Context, search string, page, pageSize int) ([]*CentroPopProntuario, int64, error)
}
