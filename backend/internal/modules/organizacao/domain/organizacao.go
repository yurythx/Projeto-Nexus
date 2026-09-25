package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Localidade representa uma unidade física ou polo socioassistencial da SEMPRAS (CRAS, CREAS, Centro POP, etc.).
type Localidade struct {
	ID        uuid.UUID  `json:"id"`
	Nome      string     `json:"nome"`
	Slug      string     `json:"slug"`
	Tipo      string     `json:"tipo"`
	GrupoAD   string     `json:"grupo_ad"`
	Endereco  string     `json:"endereco"`
	Telefone  string     `json:"telefone"`
	Bairro    string     `json:"bairro"`
	Ativo     bool       `json:"ativo"`
	Setores   []*Setor   `json:"setores,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Setor representa um departamento ou setor operacional vinculado a uma Localidade específica
// (ex.: Atendimento Técnico, Recepção e Triagem, Gerência / Coordenação).
type Setor struct {
	ID           uuid.UUID `json:"id"`
	LocalidadeID uuid.UUID `json:"localidade_id"`
	Nome         string    `json:"nome"`
	Slug         string    `json:"slug"`
	Tipo         string    `json:"tipo"` // TECNICO, RECEPCAO, GERENCIA, ADMINISTRATIVO
	GrupoAD      string    `json:"grupo_ad"`
	Descricao    string    `json:"descricao"`
	Ativo        bool      `json:"ativo"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Perfil representa uma função ou cargo RBAC da Assistência Social amarrado ao Active Directory.
type Perfil struct {
	ID         uuid.UUID `json:"id"`
	Nome       string    `json:"nome"`
	Slug       string    `json:"slug"`
	Descricao  string    `json:"descricao"`
	GrupoAD    string    `json:"grupo_ad"`
	Nivel      string    `json:"nivel"` // admin, gestao, gerencia, tecnico, recepcao
	Permissoes []string  `json:"permissoes"`
	Ativo      bool      `json:"ativo"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UsuarioLotacao consolida a alocação completa de um servidor na rede socioassistencial:
// Seus perfis, as localidades às quais tem acesso (podendo ser mais de uma) e os setores específicos.
type UsuarioLotacao struct {
	UserID      uuid.UUID     `json:"user_id"`
	Username    string        `json:"username"`
	DisplayName string        `json:"display_name"`
	Perfis      []*Perfil     `json:"perfis"`
	Localidades []*Localidade `json:"localidades"`
	Setores     []*Setor      `json:"setores"`
}

// Repository define os métodos de persistência para a estrutura organizacional da SEMPRAS.
type Repository interface {
	// Localidades
	ListLocalidades(ctx context.Context, onlyActive bool) ([]*Localidade, error)
	GetLocalidadeByID(ctx context.Context, id uuid.UUID) (*Localidade, error)
	GetLocalidadeBySlug(ctx context.Context, slug string) (*Localidade, error)
	CreateLocalidade(ctx context.Context, loc *Localidade) error
	UpdateLocalidade(ctx context.Context, loc *Localidade) error
	DeleteLocalidade(ctx context.Context, id uuid.UUID) error

	// Setores
	ListSetoresByLocalidade(ctx context.Context, localidadeID uuid.UUID) ([]*Setor, error)
	GetSetorByID(ctx context.Context, id uuid.UUID) (*Setor, error)
	CreateSetor(ctx context.Context, s *Setor) error
	UpdateSetor(ctx context.Context, s *Setor) error
	DeleteSetor(ctx context.Context, id uuid.UUID) error

	// Perfis
	ListPerfis(ctx context.Context) ([]*Perfil, error)
	GetPerfilByID(ctx context.Context, id uuid.UUID) (*Perfil, error)
	CreatePerfil(ctx context.Context, p *Perfil) error
	UpdatePerfil(ctx context.Context, p *Perfil) error

	// Lotações do Usuário
	GetUsuarioLotacao(ctx context.Context, userID uuid.UUID) (*UsuarioLotacao, error)
	SetUsuarioLotacao(ctx context.Context, userID uuid.UUID, perfilIDs []uuid.UUID, localidadeIDs []uuid.UUID, setorIDs []uuid.UUID) error
}
