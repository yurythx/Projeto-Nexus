// Package transport expõe a API HTTP do módulo núcleo IAM & Usuários.
package transport

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/yurythx/projeto-nexus/internal/modules/iam/application"
	"github.com/yurythx/projeto-nexus/internal/modules/iam/domain"
	"github.com/yurythx/projeto-nexus/internal/modules/iam/infrastructure"
	"github.com/yurythx/projeto-nexus/internal/platform/auth"
	"github.com/yurythx/projeto-nexus/pkg/httputil"
)

// PermissionCatalog devolve as permissões declaradas por todos os plugins.
type PermissionCatalog func() []PermissionGroup

// PermissionGroup agrupa as permissões de um módulo.
type PermissionGroup struct {
	Module      string           `json:"module"`
	Name        string           `json:"name"`
	Permissions []PermissionItem `json:"permissions"`
}

// PermissionItem é uma permissão com descrição.
type PermissionItem struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

// Handlers da API do IAM.
type Handlers struct {
	svc         *application.Service
	catalog     PermissionCatalog
	logger      *slog.Logger
	maxPageSize int
}

// NewHandlers cria os handlers.
func NewHandlers(svc *application.Service, catalog PermissionCatalog, logger *slog.Logger, maxPageSize int) *Handlers {
	return &Handlers{svc: svc, catalog: catalog, logger: logger, maxPageSize: maxPageSize}
}

func (h *Handlers) fail(w http.ResponseWriter, r *http.Request, err error) {
	httputil.WriteError(w, r, h.logger, application.MapError(err))
}

// RegisterRoutes monta as rotas (grupo autenticado, relativo a /api/v1).
func (h *Handlers) RegisterRoutes(r chi.Router) {
	r.Get("/me", h.Me)

	// Leitura da estrutura organizacional e do catálogo: qualquer
	// autenticado (formulários de outros módulos escolhem unidades).
	r.Get("/iam/org-tree", h.OrgTree)
	r.Get("/iam/entidades", h.ListEntidades)
	r.Get("/iam/unidades", h.ListUnidades)
	r.Get("/iam/departamentos", h.ListDepartamentos)
	r.Get("/iam/perfis", h.ListPerfis)
	r.Get("/iam/permissions", h.Permissions)

	r.Group(func(adm chi.Router) {
		adm.Use(auth.RequirePermission(h.logger, auth.PermIAMManage))
		adm.Post("/iam/entidades", h.SaveEntidade)
		adm.Put("/iam/entidades/{id}", h.SaveEntidade)
		adm.Delete("/iam/entidades/{id}", h.DeleteEntidade)
		adm.Post("/iam/unidades", h.SaveUnidade)
		adm.Put("/iam/unidades/{id}", h.SaveUnidade)
		adm.Delete("/iam/unidades/{id}", h.DeleteUnidade)
		adm.Post("/iam/departamentos", h.SaveDepartamento)
		adm.Put("/iam/departamentos/{id}", h.SaveDepartamento)
		adm.Delete("/iam/departamentos/{id}", h.DeleteDepartamento)
		adm.Post("/iam/perfis", h.SavePerfil)
		adm.Put("/iam/perfis/{id}", h.SavePerfil)
		adm.Delete("/iam/perfis/{id}", h.DeletePerfil)
		adm.Get("/iam/ad-mappings", h.ListMappings)
		adm.Post("/iam/ad-mappings", h.CreateMapping)
		adm.Delete("/iam/ad-mappings/{id}", h.DeleteMapping)
		adm.Get("/users/{id}/lotacoes", h.ListLotacoes)
		adm.Post("/users/{id}/lotacoes", h.CreateLotacao)
		adm.Delete("/users/{id}/lotacoes/{lotacaoID}", h.DeleteLotacao)
	})

	r.Group(func(read chi.Router) {
		read.Use(auth.RequirePermission(h.logger, auth.PermUsersRead))
		read.Get("/users", h.ListUsers)
		read.Get("/users/{id}", h.GetUser)
	})
	r.Group(func(adm chi.Router) {
		adm.Use(auth.RequirePermission(h.logger, auth.PermUsersManage))
		adm.Post("/users", h.CreateLocalUser)
		adm.Patch("/users/{id}", h.UpdateUser)
		adm.Post("/users/{id}/password", h.ResetPassword)
		adm.Post("/users/{id}/unlock", h.Unlock)
	})
}

type meResponse struct {
	ID          uuid.UUID    `json:"id"`
	Subject     string       `json:"subject"`
	Username    string       `json:"username"`
	Email       string       `json:"email"`
	Name        string       `json:"name"`
	Source      string       `json:"source"`
	Roles       []string     `json:"roles"`
	Groups      []string     `json:"groups"`
	Permissions []string     `json:"permissions"`
	Scopes      []auth.Scope `json:"scopes"`
}

// Me — GET /me: identidade efetiva (permissões e lotações resolvidas).
func (h *Handlers) Me(w http.ResponseWriter, r *http.Request) {
	identity, err := auth.Require(r.Context())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	nz := func(s []string) []string {
		if s == nil {
			return []string{}
		}
		return s
	}
	scopes := identity.Scopes
	if scopes == nil {
		scopes = []auth.Scope{}
	}
	httputil.WriteOK(w, meResponse{
		ID: identity.UserID, Subject: identity.Subject, Username: identity.Username, Email: identity.Email,
		Name: identity.Name, Source: string(identity.Source), Roles: nz(identity.Roles), Groups: nz(identity.Groups),
		Permissions: nz(identity.Permissions), Scopes: scopes,
	})
}

func (h *Handlers) Permissions(w http.ResponseWriter, _ *http.Request) {
	httputil.WriteOK(w, h.catalog())
}

func (h *Handlers) OrgTree(w http.ResponseWriter, r *http.Request) {
	tree, err := h.svc.OrgTree(r.Context())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, tree)
}

// ---------------------------------------------------------------- entidades

type entidadeRequest struct {
	Nome      string `json:"nome" validate:"required,max=200"`
	Sigla     string `json:"sigla" validate:"max=30"`
	Slug      string `json:"slug" validate:"omitempty,max=80"`
	Documento string `json:"documento" validate:"max=40"`
	Ativo     *bool  `json:"ativo"`
}

func optionalID(r *http.Request) (uuid.UUID, error) {
	if chi.URLParam(r, "id") == "" {
		return uuid.Nil, nil
	}
	return httputil.UUIDParam(r, "id")
}

func boolOr(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

func (h *Handlers) ListEntidades(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.ListEntidades(r.Context())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, out)
}

func (h *Handlers) SaveEntidade(w http.ResponseWriter, r *http.Request) {
	id, err := optionalID(r)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	var req entidadeRequest
	if err := httputil.Bind(w, r, &req); err != nil {
		h.fail(w, r, err)
		return
	}
	out, err := h.svc.SaveEntidade(r.Context(), domain.Entidade{
		ID: id, Nome: req.Nome, Sigla: req.Sigla, Slug: req.Slug, Documento: req.Documento, Ativo: boolOr(req.Ativo, true),
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, out)
}

func (h *Handlers) DeleteEntidade(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.UUIDParam(r, "id")
	if err == nil {
		err = h.svc.DeleteEntidade(r.Context(), id)
	}
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteNoContent(w)
}

// ----------------------------------------------------------------- unidades

type unidadeRequest struct {
	EntidadeID uuid.UUID  `json:"entidade_id" validate:"required"`
	ParentID   *uuid.UUID `json:"parent_id"`
	Nome       string     `json:"nome" validate:"required,max=200"`
	Sigla      string     `json:"sigla" validate:"max=30"`
	Slug       string     `json:"slug" validate:"omitempty,max=80"`
	ADGroup    string     `json:"ad_group" validate:"max=200"`
	Email      string     `json:"email" validate:"omitempty,email,max=200"`
	Telefone   string     `json:"telefone" validate:"max=40"`
	Endereco   string     `json:"endereco" validate:"max=300"`
	Ativo      *bool      `json:"ativo"`
}

func (h *Handlers) ListUnidades(w http.ResponseWriter, r *http.Request) {
	ent, err := httputil.OptionalUUIDQuery(r, "entidade_id")
	if err != nil {
		h.fail(w, r, err)
		return
	}
	out, err := h.svc.ListUnidades(r.Context(), ent)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, out)
}

func (h *Handlers) SaveUnidade(w http.ResponseWriter, r *http.Request) {
	id, err := optionalID(r)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	var req unidadeRequest
	if err := httputil.Bind(w, r, &req); err != nil {
		h.fail(w, r, err)
		return
	}
	out, err := h.svc.SaveUnidade(r.Context(), domain.Unidade{
		ID: id, EntidadeID: req.EntidadeID, ParentID: req.ParentID, Nome: req.Nome, Sigla: req.Sigla, Slug: req.Slug,
		ADGroup: req.ADGroup, Email: req.Email, Telefone: req.Telefone, Endereco: req.Endereco, Ativo: boolOr(req.Ativo, true),
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, out)
}

func (h *Handlers) DeleteUnidade(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.UUIDParam(r, "id")
	if err == nil {
		err = h.svc.DeleteUnidade(r.Context(), id)
	}
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteNoContent(w)
}

// ------------------------------------------------------------ departamentos

type departamentoRequest struct {
	UnidadeID uuid.UUID `json:"unidade_id" validate:"required"`
	Nome      string    `json:"nome" validate:"required,max=200"`
	Sigla     string    `json:"sigla" validate:"max=30"`
	Slug      string    `json:"slug" validate:"omitempty,max=80"`
	ADGroup   string    `json:"ad_group" validate:"max=200"`
	Email     string    `json:"email" validate:"omitempty,email,max=200"`
	Telefone  string    `json:"telefone" validate:"max=40"`
	Ativo     *bool     `json:"ativo"`
}

func (h *Handlers) ListDepartamentos(w http.ResponseWriter, r *http.Request) {
	un, err := httputil.OptionalUUIDQuery(r, "unidade_id")
	if err != nil {
		h.fail(w, r, err)
		return
	}
	out, err := h.svc.ListDepartamentos(r.Context(), un)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, out)
}

func (h *Handlers) SaveDepartamento(w http.ResponseWriter, r *http.Request) {
	id, err := optionalID(r)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	var req departamentoRequest
	if err := httputil.Bind(w, r, &req); err != nil {
		h.fail(w, r, err)
		return
	}
	out, err := h.svc.SaveDepartamento(r.Context(), domain.Departamento{
		ID: id, UnidadeID: req.UnidadeID, Nome: req.Nome, Sigla: req.Sigla, Slug: req.Slug,
		ADGroup: req.ADGroup, Email: req.Email, Telefone: req.Telefone, Ativo: boolOr(req.Ativo, true),
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, out)
}

func (h *Handlers) DeleteDepartamento(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.UUIDParam(r, "id")
	if err == nil {
		err = h.svc.DeleteDepartamento(r.Context(), id)
	}
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteNoContent(w)
}

// ------------------------------------------------------------------ perfis

type perfilRequest struct {
	Nome       string   `json:"nome" validate:"required,max=120"`
	Slug       string   `json:"slug" validate:"omitempty,max=80"`
	Descricao  string   `json:"descricao" validate:"max=500"`
	Permissoes []string `json:"permissoes" validate:"max=200,dive,max=80"`
	Ativo      *bool    `json:"ativo"`
}

func (h *Handlers) ListPerfis(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.ListPerfis(r.Context())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, out)
}

func (h *Handlers) SavePerfil(w http.ResponseWriter, r *http.Request) {
	id, err := optionalID(r)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	var req perfilRequest
	if err := httputil.Bind(w, r, &req); err != nil {
		h.fail(w, r, err)
		return
	}
	out, err := h.svc.SavePerfil(r.Context(), domain.Perfil{
		ID: id, Nome: req.Nome, Slug: req.Slug, Descricao: req.Descricao, Permissoes: req.Permissoes, Ativo: boolOr(req.Ativo, true),
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, out)
}

func (h *Handlers) DeletePerfil(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.UUIDParam(r, "id")
	if err == nil {
		err = h.svc.DeletePerfil(r.Context(), id)
	}
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteNoContent(w)
}

// ----------------------------------------------------------- mapeamentos AD

type scopeRequest struct {
	EntidadeID     *uuid.UUID `json:"entidade_id"`
	UnidadeID      *uuid.UUID `json:"unidade_id"`
	DepartamentoID *uuid.UUID `json:"departamento_id"`
}

type mappingRequest struct {
	ADGroup  string    `json:"ad_group" validate:"required,max=200"`
	PerfilID uuid.UUID `json:"perfil_id" validate:"required"`
	scopeRequest
	Descricao string `json:"descricao" validate:"max=300"`
}

func (h *Handlers) ListMappings(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.ListMappings(r.Context())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, out)
}

func (h *Handlers) CreateMapping(w http.ResponseWriter, r *http.Request) {
	var req mappingRequest
	if err := httputil.Bind(w, r, &req); err != nil {
		h.fail(w, r, err)
		return
	}
	id, err := h.svc.CreateMapping(r.Context(), domain.ADMapping{
		ADGroup: req.ADGroup, PerfilID: req.PerfilID, Descricao: req.Descricao,
		Scope: domain.Scope{EntidadeID: req.EntidadeID, UnidadeID: req.UnidadeID, DepartamentoID: req.DepartamentoID},
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteCreated(w, map[string]string{"id": id.String()})
}

func (h *Handlers) DeleteMapping(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.UUIDParam(r, "id")
	if err == nil {
		err = h.svc.DeleteMapping(r.Context(), id)
	}
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteNoContent(w)
}

// ----------------------------------------------------------------- usuários

func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	p := httputil.Page(r, h.maxPageSize)
	f := infrastructure.UserFilter{Query: httputil.Query(r, "q", 100)}
	if v := r.URL.Query().Get("active"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			f.Active = &b
		}
	}
	users, total, err := h.svc.ListUsers(r.Context(), f, p)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WritePage(w, users, p, total)
}

type userDetail struct {
	domain.User
	Lotacoes []domain.Lotacao `json:"lotacoes"`
}

func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.UUIDParam(r, "id")
	if err != nil {
		h.fail(w, r, err)
		return
	}
	u, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	lot, err := h.svc.ListLotacoes(r.Context(), id)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, userDetail{User: u, Lotacoes: lot})
}

type updateUserRequest struct {
	DisplayName string   `json:"display_name" validate:"max=200"`
	Active      *bool    `json:"active" validate:"required"`
	Roles       []string `json:"roles" validate:"max=20,dive,max=80"`
}

func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.UUIDParam(r, "id")
	if err != nil {
		h.fail(w, r, err)
		return
	}
	var req updateUserRequest
	if err := httputil.Bind(w, r, &req); err != nil {
		h.fail(w, r, err)
		return
	}
	out, err := h.svc.UpdateUser(r.Context(), id, application.UpdateUserInput{
		DisplayName: req.DisplayName, Active: *req.Active, Roles: req.Roles,
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, out)
}

type createUserRequest struct {
	Username    string   `json:"username" validate:"required,min=3,max=80"`
	Email       string   `json:"email" validate:"required,email,max=200"`
	DisplayName string   `json:"display_name" validate:"max=200"`
	Password    string   `json:"password" validate:"required,max=256"`
	Roles       []string `json:"roles" validate:"max=20,dive,max=80"`
}

func (h *Handlers) CreateLocalUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := httputil.Bind(w, r, &req); err != nil {
		h.fail(w, r, err)
		return
	}
	out, err := h.svc.CreateLocalUser(r.Context(), application.CreateLocalUserInput{
		Username: req.Username, Email: req.Email, DisplayName: req.DisplayName, Password: req.Password, Roles: req.Roles,
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteCreated(w, out)
}

type passwordRequest struct {
	Password string `json:"password" validate:"required,max=256"`
}

func (h *Handlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.UUIDParam(r, "id")
	if err != nil {
		h.fail(w, r, err)
		return
	}
	var req passwordRequest
	if err := httputil.Bind(w, r, &req); err != nil {
		h.fail(w, r, err)
		return
	}
	if err := h.svc.ResetPassword(r.Context(), id, req.Password); err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteNoContent(w)
}

func (h *Handlers) Unlock(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.UUIDParam(r, "id")
	if err == nil {
		err = h.svc.Unlock(r.Context(), id)
	}
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteNoContent(w)
}

// ---------------------------------------------------------------- lotações

type lotacaoRequest struct {
	PerfilID uuid.UUID `json:"perfil_id" validate:"required"`
	scopeRequest
	Principal bool `json:"principal"`
}

func (h *Handlers) ListLotacoes(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.UUIDParam(r, "id")
	if err != nil {
		h.fail(w, r, err)
		return
	}
	out, err := h.svc.ListLotacoes(r.Context(), id)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteOK(w, out)
}

func (h *Handlers) CreateLotacao(w http.ResponseWriter, r *http.Request) {
	userID, err := httputil.UUIDParam(r, "id")
	if err != nil {
		h.fail(w, r, err)
		return
	}
	var req lotacaoRequest
	if err := httputil.Bind(w, r, &req); err != nil {
		h.fail(w, r, err)
		return
	}
	id, err := h.svc.CreateLotacao(r.Context(), domain.Lotacao{
		UserID: userID, PerfilID: req.PerfilID, Principal: req.Principal,
		Scope: domain.Scope{EntidadeID: req.EntidadeID, UnidadeID: req.UnidadeID, DepartamentoID: req.DepartamentoID},
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteCreated(w, map[string]string{"id": id.String()})
}

func (h *Handlers) DeleteLotacao(w http.ResponseWriter, r *http.Request) {
	userID, err := httputil.UUIDParam(r, "id")
	if err != nil {
		h.fail(w, r, err)
		return
	}
	id, err := httputil.UUIDParam(r, "lotacaoID")
	if err == nil {
		err = h.svc.DeleteLotacao(r.Context(), userID, id)
	}
	if err != nil {
		h.fail(w, r, err)
		return
	}
	httputil.WriteNoContent(w)
}
