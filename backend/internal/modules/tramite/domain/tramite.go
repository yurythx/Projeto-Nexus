// Package domain define o Trâmite: processos administrativos numerados,
// com controle de sigilo, documentos e tramitação entre unidades.
package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/yurythx/projeto-nexus/internal/domain/pagination"
	"github.com/yurythx/projeto-nexus/internal/platform/auth"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
)

var (
	ErrNotFound      = errors.New("tramite: registro não encontrado")
	ErrForbidden     = errors.New("tramite: sem acesso a este processo")
	ErrInvalidState  = errors.New("tramite: operação inválida no estado atual")
	ErrSignatureOff  = errors.New("tramite: assinatura eletrônica indisponível (Signum desativado)")
	ErrDocumentState = errors.New("tramite: o documento não está em rascunho")
)

// Níveis de sigilo.
const (
	SigiloPublico  = "publico"
	SigiloRestrito = "restrito"
	SigiloSigiloso = "sigiloso"
)

// Estados do processo.
const (
	StatusAberto       = "aberto"
	StatusEmTramitacao = "em_tramitacao"
	StatusConcluido    = "concluido"
	StatusArquivado    = "arquivado"
)

// FormatNumero monta "NNNNNN/AAAA".
func FormatNumero(seq, ano int) string { return fmt.Sprintf("%06d/%04d", seq, ano) }

// Tipo é um tipo de processo.
type Tipo struct {
	ID        uuid.UUID `json:"id"`
	Slug      string    `json:"slug"`
	Nome      string    `json:"nome"`
	Descricao string    `json:"descricao"`
	Ativo     bool      `json:"ativo"`
}

// Processo é um processo administrativo.
type Processo struct {
	ID              uuid.UUID  `json:"id"`
	Numero          string     `json:"numero"`
	TipoID          uuid.UUID  `json:"tipo_id"`
	Tipo            string     `json:"tipo"`
	Assunto         string     `json:"assunto"`
	Interessado     string     `json:"interessado"`
	Descricao       string     `json:"descricao"`
	Sigilo          string     `json:"sigilo"`
	Status          string     `json:"status"`
	UnidadeOrigemID uuid.UUID  `json:"unidade_origem_id"`
	UnidadeOrigem   string     `json:"unidade_origem"`
	UnidadeAtualID  uuid.UUID  `json:"unidade_atual_id"`
	UnidadeAtual    string     `json:"unidade_atual"`
	CreatedBy       uuid.UUID  `json:"created_by"`
	CreatedByName   string     `json:"created_by_name"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	ConcluidoAt     *time.Time `json:"concluido_at,omitempty"`
}

// Documento é uma peça do processo.
type Documento struct {
	ID          uuid.UUID  `json:"id"`
	ProcessoID  uuid.UUID  `json:"processo_id"`
	Tipo        string     `json:"tipo"`
	Titulo      string     `json:"titulo"`
	Origem      string     `json:"origem"`
	Conteudo    string     `json:"conteudo,omitempty"`
	ObjectKey   string     `json:"-"`
	ContentType string     `json:"content_type,omitempty"`
	SizeBytes   int64      `json:"size_bytes"`
	SHA256      *string    `json:"sha256,omitempty"`
	Status      string     `json:"status"`
	EnvelopeID  *uuid.UUID `json:"envelope_id,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// Movimento é um registro imutável do histórico.
type Movimento struct {
	ID            uuid.UUID  `json:"id"`
	ProcessoID    uuid.UUID  `json:"processo_id"`
	Acao          string     `json:"acao"`
	DeUnidadeID   *uuid.UUID `json:"de_unidade_id,omitempty"`
	DeUnidade     string     `json:"de_unidade,omitempty"`
	ParaUnidadeID *uuid.UUID `json:"para_unidade_id,omitempty"`
	ParaUnidade   string     `json:"para_unidade,omitempty"`
	Despacho      string     `json:"despacho"`
	ActorID       *uuid.UUID `json:"actor_id,omitempty"`
	ActorName     string     `json:"actor_name,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// AccessInfo é o que a regra de sigilo precisa saber sobre o processo.
type AccessInfo struct {
	Sigilo          string
	CreatedBy       uuid.UUID
	UnidadeOrigemID uuid.UUID
	UnidadeAtualID  uuid.UUID
	ExplicitGrant   bool
}

// CanRead aplica o controle de sigilo:
//   - público: qualquer autenticado;
//   - restrito: lotados na unidade atual ou de origem, credenciados, autor
//     e tramite:manage;
//   - sigiloso: SOMENTE credenciados explícitos e o autor (nem a unidade
//     atual vê sem credencial; tramite:manage também precisa de credencial).
func CanRead(identity auth.Identity, a AccessInfo) bool {
	if a.CreatedBy == identity.UserID || a.ExplicitGrant {
		return true
	}
	switch a.Sigilo {
	case SigiloPublico:
		return true
	case SigiloRestrito:
		if auth.HasPermission(identity, auth.PermTramiteManage) {
			return true
		}
		return InUnidade(identity, a.UnidadeAtualID) || InUnidade(identity, a.UnidadeOrigemID)
	default:
		return false
	}
}

// CanAct reporta se identity pode movimentar o processo (tramitar,
// juntar documento, concluir): precisa estar lotado na unidade ATUAL, ou
// ter tramite:manage (e acesso de leitura).
func CanAct(identity auth.Identity, a AccessInfo) bool {
	if !CanRead(identity, a) {
		return false
	}
	return InUnidade(identity, a.UnidadeAtualID) || auth.HasPermission(identity, auth.PermTramiteManage)
}

// InUnidade reporta se identity tem lotação na unidade (ou em departamento
// dela — o IAM completa unidade_id a partir do departamento).
func InUnidade(identity auth.Identity, unidadeID uuid.UUID) bool {
	for _, s := range identity.Scopes {
		if s.UnidadeID != nil && *s.UnidadeID == unidadeID {
			return true
		}
	}
	return false
}

// Filter restringe a listagem.
type Filter struct {
	Query     string
	Status    string
	UnidadeID *uuid.UUID
	Mine      bool
}

// Repository é a porta de persistência.
type Repository interface {
	Tipos(ctx context.Context, db database.DBTX) ([]Tipo, error)
	NextNumero(ctx context.Context, db database.DBTX, ano int) (int, error)
	Insert(ctx context.Context, db database.DBTX, p Processo) error
	Get(ctx context.Context, db database.DBTX, id uuid.UUID, forUpdate bool) (Processo, error)
	HasGrant(ctx context.Context, db database.DBTX, processoID, userID uuid.UUID) (bool, error)
	// ListVisible aplica o sigilo no próprio SQL (para paginar corretamente).
	ListVisible(ctx context.Context, db database.DBTX, identity auth.Identity, f Filter, p pagination.Params) ([]Processo, int64, error)
	Update(ctx context.Context, db database.DBTX, p Processo) error
	Grant(ctx context.Context, db database.DBTX, processoID, userID, grantedBy uuid.UUID) error
	Grants(ctx context.Context, db database.DBTX, processoID uuid.UUID) ([]Grant, error)

	Documentos(ctx context.Context, db database.DBTX, processoID uuid.UUID) ([]Documento, error)
	GetDocumento(ctx context.Context, db database.DBTX, id uuid.UUID) (Documento, error)
	InsertDocumento(ctx context.Context, db database.DBTX, d Documento) error
	UpdateDocumento(ctx context.Context, db database.DBTX, d Documento) error
	DocumentoByEnvelope(ctx context.Context, db database.DBTX, envelopeID uuid.UUID) (Documento, error)

	AddMovimento(ctx context.Context, db database.DBTX, m Movimento) error
	Movimentos(ctx context.Context, db database.DBTX, processoID uuid.UUID) ([]Movimento, error)
}

// Grant é uma credencial de acesso.
type Grant struct {
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	GrantedBy uuid.UUID `json:"granted_by"`
	GrantedAt time.Time `json:"granted_at"`
}

// SignaturePort é a porta para o motor de assinatura (implementada em
// internal/app com o Signum — o Trâmite nunca importa o Signum).
type SignaturePort interface {
	Available() bool
	OpenEnvelope(ctx context.Context, tx pgx.Tx, req SignatureRequest) (uuid.UUID, error)
}

// SignatureRequest pede a abertura de um envelope.
type SignatureRequest struct {
	Title, Description, DocumentSHA256, SourceRef string
	SignerIDs                                     []uuid.UUID
	Sequential                                    bool
}
