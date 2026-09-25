// Package domain define o modelo do módulo núcleo IAM & Usuários: a
// estrutura organizacional multi-escopo, os Perfis de permissão, o
// mapeamento de grupos do AD e as lotações de cada usuário.
package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound         = errors.New("iam: registro não encontrado")
	ErrConflict         = errors.New("iam: registro duplicado")
	ErrSystemProfile    = errors.New("iam: perfil de sistema não pode ser removido")
	ErrInvalidPermision = errors.New("iam: permissão inválida")
	ErrInvalidScope     = errors.New("iam: escopo organizacional inconsistente")
)

// Entidade é o tenant (órgão/empresa).
type Entidade struct {
	ID        uuid.UUID `json:"id"`
	Nome      string    `json:"nome"`
	Sigla     string    `json:"sigla"`
	Slug      string    `json:"slug"`
	Documento string    `json:"documento"`
	Ativo     bool      `json:"ativo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Unidade pertence a uma Entidade e pode ter unidade-mãe (hierarquia).
type Unidade struct {
	ID         uuid.UUID  `json:"id"`
	EntidadeID uuid.UUID  `json:"entidade_id"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	Nome       string     `json:"nome"`
	Sigla      string     `json:"sigla"`
	Slug       string     `json:"slug"`
	ADGroup    string     `json:"ad_group"`
	Email      string     `json:"email"`
	Telefone   string     `json:"telefone"`
	Endereco   string     `json:"endereco"`
	Ativo      bool       `json:"ativo"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Departamento pertence a uma Unidade.
type Departamento struct {
	ID        uuid.UUID `json:"id"`
	UnidadeID uuid.UUID `json:"unidade_id"`
	Nome      string    `json:"nome"`
	Sigla     string    `json:"sigla"`
	Slug      string    `json:"slug"`
	ADGroup   string    `json:"ad_group"`
	Email     string    `json:"email"`
	Telefone  string    `json:"telefone"`
	Ativo     bool      `json:"ativo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Perfil agrega permissões "recurso:ação".
type Perfil struct {
	ID         uuid.UUID `json:"id"`
	Slug       string    `json:"slug"`
	Nome       string    `json:"nome"`
	Descricao  string    `json:"descricao"`
	Permissoes []string  `json:"permissoes"`
	Sistema    bool      `json:"sistema"`
	Ativo      bool      `json:"ativo"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Scope é o escopo organizacional de uma lotação ou mapeamento.
type Scope struct {
	EntidadeID     *uuid.UUID `json:"entidade_id,omitempty"`
	UnidadeID      *uuid.UUID `json:"unidade_id,omitempty"`
	DepartamentoID *uuid.UUID `json:"departamento_id,omitempty"`
}

// ADMapping correlaciona um grupo do AD a um Perfil num escopo.
type ADMapping struct {
	ID         uuid.UUID `json:"id"`
	ADGroup    string    `json:"ad_group"`
	PerfilID   uuid.UUID `json:"perfil_id"`
	PerfilNome string    `json:"perfil_nome"`
	Scope
	ScopeLabel string    `json:"scope_label"`
	Descricao  string    `json:"descricao"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  string    `json:"created_by"`
}

// Lotacao é um perfil atribuído manualmente a um usuário num escopo.
type Lotacao struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	PerfilID   uuid.UUID `json:"perfil_id"`
	PerfilNome string    `json:"perfil_nome"`
	Scope
	ScopeLabel string    `json:"scope_label"`
	Principal  bool      `json:"principal"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  string    `json:"created_by"`
}

// User é a visão administrativa de uma conta.
type User struct {
	ID          uuid.UUID  `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	DisplayName string     `json:"display_name"`
	Active      bool       `json:"active"`
	Federated   bool       `json:"federated"` // conta do Keycloak/AD
	LocalLogin  bool       `json:"local_login"`
	Roles       []string   `json:"roles"`
	Groups      []string   `json:"groups"`
	LockedUntil *time.Time `json:"locked_until,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	LastSeenAt  *time.Time `json:"last_seen_at,omitempty"`
	ADSyncedAt  *time.Time `json:"ad_synced_at,omitempty"`
}

// OrgTree é a estrutura organizacional aninhada.
type OrgTree struct {
	Entidade
	Unidades []OrgUnidade `json:"unidades"`
}

// OrgUnidade é uma unidade com seus departamentos.
type OrgUnidade struct {
	Unidade
	Departamentos []Departamento `json:"departamentos"`
}

var permissionPattern = regexp.MustCompile(`^(\*|[a-z_]+:(\*|[a-z_]+))$`)

// ValidPermission reporta se p segue o formato "recurso:ação".
func ValidPermission(p string) bool { return permissionPattern.MatchString(p) }
