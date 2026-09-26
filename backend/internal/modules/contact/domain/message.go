// Package domain define as mensagens do formulário público de Contato.
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/yurythx/projeto-nexus/internal/domain/pagination"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
)

// ErrNotFound indica mensagem inexistente.
var ErrNotFound = errors.New("contact: mensagem não encontrada")

// ErrAssignee indica responsável inexistente ou desativado.
var ErrAssignee = errors.New("contact: responsável inexistente ou desativado")

// Message é uma mensagem recebida (contém PII de visitante externo).
type Message struct {
	ID         uuid.UUID  `json:"id"`
	Protocol   string     `json:"protocol"`
	Name       string     `json:"name"`
	Email      string     `json:"email"`
	Phone      string     `json:"phone"`
	Subject    string     `json:"subject"`
	Category   string     `json:"category"`
	Message    string     `json:"message"`
	ConsentAt  time.Time  `json:"consent_at"`
	IPAddress  string     `json:"ip_address,omitempty"`
	UserAgent  string     `json:"-"`
	Status     string     `json:"status"`
	AssignedTo *uuid.UUID `json:"assigned_to,omitempty"`
	Notes      string     `json:"notes"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Summary é a forma de listagem (sem o corpo e sem telefone/IP).
type Summary struct {
	ID        uuid.UUID `json:"id"`
	Protocol  string    `json:"protocol"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Subject   string    `json:"subject"`
	Category  string    `json:"category"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Repository é a porta de persistência.
type Repository interface {
	Insert(ctx context.Context, db database.DBTX, m Message) (Message, error)
	List(ctx context.Context, db database.DBTX, status string, p pagination.Params) ([]Summary, int64, error)
	Get(ctx context.Context, db database.DBTX, id uuid.UUID) (Message, error)
	UpdateTriage(ctx context.Context, db database.DBTX, id uuid.UUID, status, notes string, assignedTo *uuid.UUID) error
	// ActiveUser reporta se a conta existe e está ativa (responsável).
	ActiveUser(ctx context.Context, db database.DBTX, id uuid.UUID) (bool, error)
}
