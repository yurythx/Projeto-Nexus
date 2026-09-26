// Package application contém os casos de uso do IAM. Toda mutação grava
// a auditoria (com diff antes/depois) na mesma transação e, ao final,
// invalida o cache de permissões de todas as réplicas.
package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/yurythx/projeto-nexus/internal/domain/errors"
	"github.com/yurythx/projeto-nexus/internal/domain/pagination"
	"github.com/yurythx/projeto-nexus/internal/modules/iam/domain"
	"github.com/yurythx/projeto-nexus/internal/modules/iam/infrastructure"
	"github.com/yurythx/projeto-nexus/internal/platform/audit"
	"github.com/yurythx/projeto-nexus/internal/platform/auth"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
	"github.com/yurythx/projeto-nexus/internal/platform/modkit"
	"github.com/yurythx/projeto-nexus/internal/platform/passwords"
)

// Service implementa os casos de uso do IAM.
type Service struct {
	pool         *pgxpool.Pool
	repo         *infrastructure.Repository
	invalidate   func(ctx context.Context)
	resetLockout func(ctx context.Context, username string) error
}

// NewService cria o serviço.
func NewService(pool *pgxpool.Pool, repo *infrastructure.Repository, invalidate func(context.Context)) *Service {
	if invalidate == nil {
		invalidate = func(context.Context) {}
	}
	return &Service{pool: pool, repo: repo, invalidate: invalidate}
}

// WithLoginLockoutReset conecta a limpeza do bloqueio progressivo de login
// (Redis) ao desbloqueio administrativo.
func (s *Service) WithLoginLockoutReset(fn func(ctx context.Context, username string) error) *Service {
	s.resetLockout = fn
	return s
}

// guardGrant aplica "ninguém concede o que não tem" (A01): cada permissão
// nova (ausente em before) precisa estar coberta pelas permissões efetivas
// de quem está concedendo — senão um detentor de iam:manage criaria um
// perfil "*" e se lotaria nele.
func guardGrant(ctx context.Context, before, after []string) error {
	identity, _ := auth.IdentityFromContext(ctx)
	for _, p := range after {
		if slices.Contains(before, p) {
			continue
		}
		if !auth.HasPermission(identity, auth.Permission(p)) {
			return apperrors.Forbidden("você não pode conceder a permissão " + p + ", que não possui")
		}
	}
	return nil
}

// guardAdminRole impede escalada de privilégio (A01): só quem já detém
// acesso total ("*") concede ou retira o papel de administrador da
// plataforma (nexus-admin equivale a "*").
func guardAdminRole(ctx context.Context, before, after []string) error {
	if slices.Contains(before, auth.RoleAdmin) == slices.Contains(after, auth.RoleAdmin) {
		return nil
	}
	identity, _ := auth.IdentityFromContext(ctx)
	if !auth.HasPermission(identity, "*") {
		return apperrors.Forbidden("apenas administradores com acesso total podem conceder ou retirar o papel " + auth.RoleAdmin)
	}
	return nil
}

// MapError traduz erros de domínio em erros HTTP.
func MapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return apperrors.NotFound("registro não encontrado")
	case errors.Is(err, domain.ErrConflict):
		return apperrors.Conflict("já existe um registro com esses dados (slug, grupo ou vínculo duplicado)")
	case errors.Is(err, domain.ErrInUse):
		return apperrors.Conflict("registro em uso: remova antes a estrutura abaixo, as lotações e os mapeamentos do AD que o referenciam")
	case errors.Is(err, domain.ErrSystemProfile):
		return apperrors.Conflict("perfis de sistema não podem ser removidos")
	case errors.Is(err, domain.ErrInvalidPermision):
		return apperrors.Validation(err.Error())
	case errors.Is(err, domain.ErrInvalidScope):
		return apperrors.Validation("escopo organizacional inconsistente (unidade/departamento não pertencem ao nível acima)")
	}
	return err
}

// tx roda fn numa transação e grava a entrada de auditoria no final.
func (s *Service) tx(ctx context.Context, invalidate bool, fn func(ctx context.Context, tx pgx.Tx) (audit.Entry, error)) error {
	err := database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		entry, err := fn(ctx, tx)
		if err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, entry)
	})
	if err != nil {
		return MapError(err)
	}
	if invalidate {
		s.invalidate(ctx)
	}
	return nil
}

// ------------------------------------------------------------ estrutura org.

// OrgTree devolve entidades > unidades > departamentos.
func (s *Service) OrgTree(ctx context.Context) ([]domain.OrgTree, error) {
	ents, err := s.repo.ListEntidades(ctx, s.pool)
	if err != nil {
		return nil, err
	}
	unids, err := s.repo.ListUnidades(ctx, s.pool, nil)
	if err != nil {
		return nil, err
	}
	deps, err := s.repo.ListDepartamentos(ctx, s.pool, nil)
	if err != nil {
		return nil, err
	}
	byUnidade := map[uuid.UUID][]domain.Departamento{}
	for _, d := range deps {
		byUnidade[d.UnidadeID] = append(byUnidade[d.UnidadeID], d)
	}
	byEntidade := map[uuid.UUID][]domain.OrgUnidade{}
	for _, u := range unids {
		d := byUnidade[u.ID]
		if d == nil {
			d = []domain.Departamento{}
		}
		byEntidade[u.EntidadeID] = append(byEntidade[u.EntidadeID], domain.OrgUnidade{Unidade: u, Departamentos: d})
	}
	out := make([]domain.OrgTree, 0, len(ents))
	for _, e := range ents {
		u := byEntidade[e.ID]
		if u == nil {
			u = []domain.OrgUnidade{}
		}
		out = append(out, domain.OrgTree{Entidade: e, Unidades: u})
	}
	return out, nil
}

func (s *Service) ListEntidades(ctx context.Context) ([]domain.Entidade, error) {
	return s.repo.ListEntidades(ctx, s.pool)
}

// SaveEntidade cria (ID nil) ou atualiza uma entidade.
func (s *Service) SaveEntidade(ctx context.Context, in domain.Entidade) (domain.Entidade, error) {
	if in.Slug == "" {
		in.Slug = modkit.Slugify(in.Nome)
	}
	var out domain.Entidade
	err := s.tx(ctx, false, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		var before any
		if in.ID == uuid.Nil {
			in.ID = uuid.New()
		} else if prev, err := s.repo.GetEntidade(ctx, tx, in.ID); err == nil {
			before = prev
		} else {
			return audit.Entry{}, err
		}
		var err error
		out, err = s.repo.UpsertEntidade(ctx, tx, in)
		return audit.Meta(ctx, "iam.entidade.saved", "entidade", out.ID.String(), before, out), err
	})
	return out, err
}

func (s *Service) DeleteEntidade(ctx context.Context, id uuid.UUID) error {
	return s.tx(ctx, true, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		prev, err := s.repo.GetEntidade(ctx, tx, id)
		if err != nil {
			return audit.Entry{}, err
		}
		return audit.Meta(ctx, "iam.entidade.deleted", "entidade", id.String(), prev, nil), s.repo.DeleteEntidade(ctx, tx, id)
	})
}

func (s *Service) ListUnidades(ctx context.Context, entidadeID *uuid.UUID) ([]domain.Unidade, error) {
	return s.repo.ListUnidades(ctx, s.pool, entidadeID)
}

func (s *Service) SaveUnidade(ctx context.Context, in domain.Unidade) (domain.Unidade, error) {
	if in.Slug == "" {
		in.Slug = modkit.Slugify(in.Nome)
	}
	var out domain.Unidade
	err := s.tx(ctx, false, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		var before any
		if in.ID == uuid.Nil {
			in.ID = uuid.New()
		} else if prev, err := s.repo.GetUnidade(ctx, tx, in.ID); err == nil {
			before = prev
		} else {
			return audit.Entry{}, err
		}
		if in.ParentID != nil {
			parent, err := s.repo.GetUnidade(ctx, tx, *in.ParentID)
			if err != nil || parent.EntidadeID != in.EntidadeID {
				return audit.Entry{}, domain.ErrInvalidScope
			}
			if cycle, err := s.repo.UnidadeCreatesCycle(ctx, tx, in.ID, *in.ParentID); err != nil || cycle {
				return audit.Entry{}, fmt.Errorf("%w: hierarquia circular", domain.ErrInvalidScope)
			}
		}
		var err error
		out, err = s.repo.UpsertUnidade(ctx, tx, in)
		return audit.Meta(ctx, "iam.unidade.saved", "unidade", out.ID.String(), before, out), err
	})
	return out, err
}

func (s *Service) DeleteUnidade(ctx context.Context, id uuid.UUID) error {
	return s.tx(ctx, true, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		prev, err := s.repo.GetUnidade(ctx, tx, id)
		if err != nil {
			return audit.Entry{}, err
		}
		return audit.Meta(ctx, "iam.unidade.deleted", "unidade", id.String(), prev, nil), s.repo.DeleteUnidade(ctx, tx, id)
	})
}

func (s *Service) ListDepartamentos(ctx context.Context, unidadeID *uuid.UUID) ([]domain.Departamento, error) {
	return s.repo.ListDepartamentos(ctx, s.pool, unidadeID)
}

func (s *Service) SaveDepartamento(ctx context.Context, in domain.Departamento) (domain.Departamento, error) {
	if in.Slug == "" {
		in.Slug = modkit.Slugify(in.Nome)
	}
	var out domain.Departamento
	err := s.tx(ctx, false, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		var before any
		if in.ID == uuid.Nil {
			in.ID = uuid.New()
		} else if prev, err := s.repo.GetDepartamento(ctx, tx, in.ID); err == nil {
			before = prev
		} else {
			return audit.Entry{}, err
		}
		var err error
		out, err = s.repo.UpsertDepartamento(ctx, tx, in)
		return audit.Meta(ctx, "iam.departamento.saved", "departamento", out.ID.String(), before, out), err
	})
	return out, err
}

func (s *Service) DeleteDepartamento(ctx context.Context, id uuid.UUID) error {
	return s.tx(ctx, true, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		prev, err := s.repo.GetDepartamento(ctx, tx, id)
		if err != nil {
			return audit.Entry{}, err
		}
		return audit.Meta(ctx, "iam.departamento.deleted", "departamento", id.String(), prev, nil), s.repo.DeleteDepartamento(ctx, tx, id)
	})
}

// ------------------------------------------------------------------ perfis

func (s *Service) ListPerfis(ctx context.Context) ([]domain.Perfil, error) {
	return s.repo.ListPerfis(ctx, s.pool)
}

func normalizePermissions(perms []string) ([]string, error) {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(perms))
	for _, p := range perms {
		p = strings.TrimSpace(strings.ToLower(p))
		if p == "" {
			continue
		}
		if !domain.ValidPermission(p) {
			return nil, fmt.Errorf("%w: %q (use recurso:acao, recurso:* ou *)", domain.ErrInvalidPermision, p)
		}
		if _, dup := seen[p]; !dup {
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	return out, nil
}

func (s *Service) SavePerfil(ctx context.Context, in domain.Perfil) (domain.Perfil, error) {
	perms, err := normalizePermissions(in.Permissoes)
	if err != nil {
		return domain.Perfil{}, MapError(err)
	}
	in.Permissoes = perms
	if in.Slug == "" {
		in.Slug = modkit.Slugify(in.Nome)
	}
	var out domain.Perfil
	err = s.tx(ctx, true, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		var before any
		var prevPerms []string
		if in.ID == uuid.Nil {
			in.ID = uuid.New()
		} else {
			prev, err := s.repo.GetPerfil(ctx, tx, in.ID)
			if err != nil {
				return audit.Entry{}, err
			}
			// O administrador de sistema mantém "*" sempre: um erro de
			// edição não pode trancar toda a plataforma.
			if prev.Sistema && prev.Slug == "administrador" {
				in.Permissoes = []string{"*"}
				in.Slug = prev.Slug
				in.Ativo = true
			}
			if prev.Sistema {
				in.Slug = prev.Slug
			}
			before = prev
			prevPerms = prev.Permissoes
		}
		if err := guardGrant(ctx, prevPerms, in.Permissoes); err != nil {
			return audit.Entry{}, err
		}
		var err error
		out, err = s.repo.UpsertPerfil(ctx, tx, in)
		return audit.Meta(ctx, "iam.perfil.saved", "perfil", out.ID.String(), before, out), err
	})
	return out, err
}

func (s *Service) DeletePerfil(ctx context.Context, id uuid.UUID) error {
	return s.tx(ctx, true, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		prev, err := s.repo.GetPerfil(ctx, tx, id)
		if err != nil {
			return audit.Entry{}, err
		}
		return audit.Meta(ctx, "iam.perfil.deleted", "perfil", id.String(), prev, nil), s.repo.DeletePerfil(ctx, tx, id)
	})
}

// ------------------------------------------------------- mapeamentos do AD

func (s *Service) ListMappings(ctx context.Context) ([]domain.ADMapping, error) {
	return s.repo.ListMappings(ctx, s.pool)
}

func (s *Service) CreateMapping(ctx context.Context, in domain.ADMapping) (uuid.UUID, error) {
	in.ADGroup = strings.TrimSpace(in.ADGroup)
	if in.ADGroup == "" {
		return uuid.Nil, apperrors.Validation("ad_group é obrigatório")
	}
	identity, _ := auth.IdentityFromContext(ctx)
	in.CreatedBy = identity.Username
	var id uuid.UUID
	err := s.tx(ctx, true, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		if err := s.guardPerfilGrant(ctx, tx, in.PerfilID); err != nil {
			return audit.Entry{}, err
		}
		scope, err := s.repo.ScopeConsistent(ctx, tx, in.Scope)
		if err != nil {
			return audit.Entry{}, err
		}
		in.Scope = scope
		if id, err = s.repo.CreateMapping(ctx, tx, in); err != nil {
			return audit.Entry{}, err
		}
		in.ID = id
		return audit.Meta(ctx, "iam.ad_mapping.created", "ad_group_mapping", id.String(), nil, in), nil
	})
	return id, err
}

func (s *Service) DeleteMapping(ctx context.Context, id uuid.UUID) error {
	return s.tx(ctx, true, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		prev, err := s.repo.DeleteMapping(ctx, tx, id)
		return audit.Meta(ctx, "iam.ad_mapping.deleted", "ad_group_mapping", id.String(), prev, nil), err
	})
}

// ---------------------------------------------------------------- usuários

func (s *Service) ListUsers(ctx context.Context, f infrastructure.UserFilter, p pagination.Params) ([]domain.User, int64, error) {
	return s.repo.ListUsers(ctx, s.pool, f, p)
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (domain.User, error) {
	u, err := s.repo.GetUser(ctx, s.pool, id)
	return u, MapError(err)
}

// UpdateUserInput são os campos administráveis de uma conta.
type UpdateUserInput struct {
	DisplayName string
	Active      bool
	Roles       []string
}

func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, in UpdateUserInput) (domain.User, error) {
	identity, _ := auth.IdentityFromContext(ctx)
	if identity.UserID == id && !in.Active {
		return domain.User{}, apperrors.Conflict("você não pode desativar a própria conta")
	}
	if in.Roles == nil {
		in.Roles = []string{}
	}
	var out domain.User
	err := s.tx(ctx, true, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		prev, err := s.repo.GetUser(ctx, tx, id)
		if err != nil {
			return audit.Entry{}, err
		}
		if err := guardAdminRole(ctx, prev.Roles, in.Roles); err != nil {
			return audit.Entry{}, err
		}
		if err := s.repo.UpdateUser(ctx, tx, id, in.DisplayName, in.Active, in.Roles); err != nil {
			return audit.Entry{}, err
		}
		if out, err = s.repo.GetUser(ctx, tx, id); err != nil {
			return audit.Entry{}, err
		}
		return audit.Meta(ctx, audit.ActionUserUpdated, "user", id.String(), prev, out), nil
	})
	return out, err
}

// CreateLocalUserInput cria uma conta local (contingência/ambiente isolado).
type CreateLocalUserInput struct {
	Username    string
	Email       string
	DisplayName string
	Password    string
	Roles       []string
}

func (s *Service) CreateLocalUser(ctx context.Context, in CreateLocalUserInput) (domain.User, error) {
	if !usernamePattern.MatchString(in.Username) {
		return domain.User{}, apperrors.Validation("username deve ter 3 a 80 caracteres: letras, dígitos, ponto, hífen ou sublinhado")
	}
	if err := ValidatePasswordStrength(in.Password); err != nil {
		return domain.User{}, err
	}
	if err := guardAdminRole(ctx, nil, in.Roles); err != nil {
		return domain.User{}, err
	}
	hash, err := passwords.Hash(in.Password)
	if err != nil {
		return domain.User{}, apperrors.Internal(err)
	}
	if in.Roles == nil {
		in.Roles = []string{}
	}
	var out domain.User
	err = s.tx(ctx, true, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		id, err := s.repo.CreateLocalUser(ctx, tx, strings.TrimSpace(in.Username), strings.TrimSpace(strings.ToLower(in.Email)), in.DisplayName, hash, in.Roles)
		if err != nil {
			return audit.Entry{}, err
		}
		if out, err = s.repo.GetUser(ctx, tx, id); err != nil {
			return audit.Entry{}, err
		}
		return audit.Meta(ctx, audit.ActionUserCreated, "user", id.String(), nil, out), nil
	})
	return out, err
}

func (s *Service) ResetPassword(ctx context.Context, id uuid.UUID, password string) error {
	if err := ValidatePasswordStrength(password); err != nil {
		return err
	}
	hash, err := passwords.Hash(password)
	if err != nil {
		return apperrors.Internal(err)
	}
	return s.tx(ctx, false, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		// Nunca a senha nem o hash vão para a auditoria.
		return audit.Meta(ctx, "user.password.reset", "user", id.String(), nil, nil), s.repo.SetPassword(ctx, tx, id, hash)
	})
}

// Unlock zera o bloqueio da conta (banco) E o bloqueio progressivo de
// login distribuído (Redis) — sem o segundo, o login continuaria
// recusado até a janela expirar.
func (s *Service) Unlock(ctx context.Context, id uuid.UUID) error {
	var username string
	err := s.tx(ctx, false, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		u, err := s.repo.GetUser(ctx, tx, id)
		if err != nil {
			return audit.Entry{}, err
		}
		username = u.Username
		return audit.Meta(ctx, "user.unlocked", "user", id.String(), nil, nil), s.repo.Unlock(ctx, tx, id)
	})
	if err != nil || s.resetLockout == nil {
		return err
	}
	if err := s.resetLockout(ctx, username); err != nil {
		return apperrors.DependencyUnavailable("conta desbloqueada, mas o bloqueio de login distribuído não pôde ser limpo — tente novamente").WithCause(err)
	}
	return nil
}

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]{3,80}$`)

// ValidatePasswordStrength aplica a política mínima de senha local.
func ValidatePasswordStrength(p string) error {
	if len(p) < 12 {
		return apperrors.Validation("a senha precisa ter pelo menos 12 caracteres")
	}
	var lower, upper, digit, other bool
	for _, r := range p {
		switch {
		case r >= 'a' && r <= 'z':
			lower = true
		case r >= 'A' && r <= 'Z':
			upper = true
		case r >= '0' && r <= '9':
			digit = true
		default:
			other = true
		}
	}
	classes := 0
	for _, ok := range []bool{lower, upper, digit, other} {
		if ok {
			classes++
		}
	}
	if classes < 3 {
		return apperrors.Validation("a senha precisa combinar ao menos 3 tipos: minúsculas, maiúsculas, dígitos e símbolos")
	}
	return nil
}

// ---------------------------------------------------------------- lotações

// guardPerfilGrant: atribuir um perfil (lotação ou mapeamento do AD)
// concede todas as permissões dele.
func (s *Service) guardPerfilGrant(ctx context.Context, tx pgx.Tx, perfilID uuid.UUID) error {
	p, err := s.repo.GetPerfil(ctx, tx, perfilID)
	if err != nil {
		return err
	}
	return guardGrant(ctx, nil, p.Permissoes)
}

func (s *Service) ListLotacoes(ctx context.Context, userID uuid.UUID) ([]domain.Lotacao, error) {
	return s.repo.ListLotacoes(ctx, s.pool, userID)
}

func (s *Service) CreateLotacao(ctx context.Context, in domain.Lotacao) (uuid.UUID, error) {
	identity, _ := auth.IdentityFromContext(ctx)
	in.CreatedBy = identity.Username
	var id uuid.UUID
	err := s.tx(ctx, true, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		if err := s.guardPerfilGrant(ctx, tx, in.PerfilID); err != nil {
			return audit.Entry{}, err
		}
		scope, err := s.repo.ScopeConsistent(ctx, tx, in.Scope)
		if err != nil {
			return audit.Entry{}, err
		}
		in.Scope = scope
		if id, err = s.repo.CreateLotacao(ctx, tx, in); err != nil {
			return audit.Entry{}, err
		}
		in.ID = id
		return audit.Meta(ctx, "iam.lotacao.created", "user", in.UserID.String(), nil, in), nil
	})
	return id, err
}

func (s *Service) DeleteLotacao(ctx context.Context, userID, id uuid.UUID) error {
	return s.tx(ctx, true, func(ctx context.Context, tx pgx.Tx) (audit.Entry, error) {
		return audit.Meta(ctx, "iam.lotacao.deleted", "user", userID.String(), map[string]string{"lotacao_id": id.String()}, nil),
			s.repo.DeleteLotacao(ctx, tx, userID, id)
	})
}
