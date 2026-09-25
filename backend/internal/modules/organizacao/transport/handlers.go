package transport

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	apperrors "github.com/yurythx/projeto-aurora/internal/domain/errors"
	"github.com/yurythx/projeto-aurora/internal/modules/organizacao/application"
	"github.com/yurythx/projeto-aurora/internal/modules/organizacao/domain"
	"github.com/yurythx/projeto-aurora/internal/platform/auth"
	"github.com/yurythx/projeto-aurora/pkg/httputil"
)

type Handlers struct {
	service *application.Service
	logger  *slog.Logger
}

func NewHandlers(service *application.Service, logger *slog.Logger) *Handlers {
	return &Handlers{
		service: service,
		logger:  logger,
	}
}

func (h *Handlers) getIdentity(r *http.Request) (auth.Identity, error) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		return auth.Identity{}, apperrors.Unauthorized("authentication required")
	}
	return identity, nil
}

// Localidades Handlers

func (h *Handlers) ListLocalidades(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("all") != "true"
	locs, err := h.service.ListLocalidades(r.Context(), onlyActive)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}
	httputil.WriteOK(w, locs)
}

func (h *Handlers) CreateLocalidade(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	var req struct {
		Nome     string `json:"nome"`
		Slug     string `json:"slug"`
		Tipo     string `json:"tipo"`
		GrupoAD  string `json:"grupo_ad"`
		Endereco string `json:"endereco"`
		Telefone string `json:"telefone"`
		Bairro   string `json:"bairro"`
	}
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	loc := &domain.Localidade{
		Nome:     req.Nome,
		Slug:     req.Slug,
		Tipo:     req.Tipo,
		GrupoAD:  req.GrupoAD,
		Endereco: req.Endereco,
		Telefone: req.Telefone,
		Bairro:   req.Bairro,
		Ativo:    true,
	}

	if err := h.service.CreateLocalidade(r.Context(), identity, loc); err != nil {
		if errors.Is(err, application.ErrForbidden) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteCreated(w, loc)
}

func (h *Handlers) UpdateLocalidade(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, r, h.logger, apperrors.BadRequest("invalid localidade id"))
		return
	}

	var req struct {
		Nome     string `json:"nome"`
		Slug     string `json:"slug"`
		Tipo     string `json:"tipo"`
		GrupoAD  string `json:"grupo_ad"`
		Endereco string `json:"endereco"`
		Telefone string `json:"telefone"`
		Bairro   string `json:"bairro"`
		Ativo    bool   `json:"ativo"`
	}
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	loc := &domain.Localidade{
		ID:       id,
		Nome:     req.Nome,
		Slug:     req.Slug,
		Tipo:     req.Tipo,
		GrupoAD:  req.GrupoAD,
		Endereco: req.Endereco,
		Telefone: req.Telefone,
		Bairro:   req.Bairro,
		Ativo:    req.Ativo,
	}

	if err := h.service.UpdateLocalidade(r.Context(), identity, loc); err != nil {
		if errors.Is(err, application.ErrForbidden) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteOK(w, loc)
}

func (h *Handlers) DeleteLocalidade(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, r, h.logger, apperrors.BadRequest("invalid localidade id"))
		return
	}

	if err := h.service.DeleteLocalidade(r.Context(), identity, id); err != nil {
		if errors.Is(err, application.ErrForbidden) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Setores Handlers

func (h *Handlers) ListSetores(w http.ResponseWriter, r *http.Request) {
	localidadeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, r, h.logger, apperrors.BadRequest("invalid localidade id"))
		return
	}

	setores, err := h.service.ListSetoresByLocalidade(r.Context(), localidadeID)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}
	httputil.WriteOK(w, setores)
}

func (h *Handlers) CreateSetor(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	localidadeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, r, h.logger, apperrors.BadRequest("invalid localidade id"))
		return
	}

	var req struct {
		Nome      string `json:"nome"`
		Slug      string `json:"slug"`
		Tipo      string `json:"tipo"`
		GrupoAD   string `json:"grupo_ad"`
		Descricao string `json:"descricao"`
	}
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	setor := &domain.Setor{
		LocalidadeID: localidadeID,
		Nome:         req.Nome,
		Slug:         req.Slug,
		Tipo:         req.Tipo,
		GrupoAD:      req.GrupoAD,
		Descricao:    req.Descricao,
		Ativo:        true,
	}

	if err := h.service.CreateSetor(r.Context(), identity, setor); err != nil {
		if errors.Is(err, application.ErrForbidden) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteCreated(w, setor)
}

func (h *Handlers) UpdateSetor(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, r, h.logger, apperrors.BadRequest("invalid setor id"))
		return
	}

	var req struct {
		Nome      string `json:"nome"`
		Slug      string `json:"slug"`
		Tipo      string `json:"tipo"`
		GrupoAD   string `json:"grupo_ad"`
		Descricao string `json:"descricao"`
		Ativo     bool   `json:"ativo"`
	}
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	setor := &domain.Setor{
		ID:        id,
		Nome:      req.Nome,
		Slug:      req.Slug,
		Tipo:      req.Tipo,
		GrupoAD:   req.GrupoAD,
		Descricao: req.Descricao,
		Ativo:     req.Ativo,
	}

	if err := h.service.UpdateSetor(r.Context(), identity, setor); err != nil {
		if errors.Is(err, application.ErrForbidden) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteOK(w, setor)
}

func (h *Handlers) DeleteSetor(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, r, h.logger, apperrors.BadRequest("invalid setor id"))
		return
	}

	if err := h.service.DeleteSetor(r.Context(), identity, id); err != nil {
		if errors.Is(err, application.ErrForbidden) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Perfis Handlers

func (h *Handlers) ListPerfis(w http.ResponseWriter, r *http.Request) {
	perfis, err := h.service.ListPerfis(r.Context())
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}
	httputil.WriteOK(w, perfis)
}

func (h *Handlers) CreatePerfil(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	var req struct {
		Nome       string   `json:"nome"`
		Slug       string   `json:"slug"`
		Descricao  string   `json:"descricao"`
		GrupoAD    string   `json:"grupo_ad"`
		Nivel      string   `json:"nivel"`
		Permissoes []string `json:"permissoes"`
	}
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	perfil := &domain.Perfil{
		Nome:       req.Nome,
		Slug:       req.Slug,
		Descricao:  req.Descricao,
		GrupoAD:    req.GrupoAD,
		Nivel:      req.Nivel,
		Permissoes: req.Permissoes,
		Ativo:      true,
	}

	if err := h.service.CreatePerfil(r.Context(), identity, perfil); err != nil {
		if errors.Is(err, application.ErrForbidden) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteCreated(w, perfil)
}

func (h *Handlers) UpdatePerfil(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, r, h.logger, apperrors.BadRequest("invalid perfil id"))
		return
	}

	var req struct {
		Nome       string   `json:"nome"`
		Slug       string   `json:"slug"`
		Descricao  string   `json:"descricao"`
		GrupoAD    string   `json:"grupo_ad"`
		Nivel      string   `json:"nivel"`
		Permissoes []string `json:"permissoes"`
		Ativo      bool     `json:"ativo"`
	}
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	perfil := &domain.Perfil{
		ID:         id,
		Nome:       req.Nome,
		Slug:       req.Slug,
		Descricao:  req.Descricao,
		GrupoAD:    req.GrupoAD,
		Nivel:      req.Nivel,
		Permissoes: req.Permissoes,
		Ativo:      req.Ativo,
	}

	if err := h.service.UpdatePerfil(r.Context(), identity, perfil); err != nil {
		if errors.Is(err, application.ErrForbidden) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteOK(w, perfil)
}

// Lotação Handlers

func (h *Handlers) GetUsuarioLotacao(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, r, h.logger, apperrors.BadRequest("invalid user id"))
		return
	}

	lot, err := h.service.GetUsuarioLotacao(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}
	httputil.WriteOK(w, lot)
}

func (h *Handlers) SetUsuarioLotacao(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, r, h.logger, apperrors.BadRequest("invalid user id"))
		return
	}

	var req struct {
		PerfilIDs     []uuid.UUID `json:"perfil_ids"`
		LocalidadeIDs []uuid.UUID `json:"localidade_ids"`
		SetorIDs      []uuid.UUID `json:"setor_ids"`
	}
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	if err := h.service.SetUsuarioLotacao(r.Context(), identity, userID, req.PerfilIDs, req.LocalidadeIDs, req.SetorIDs); err != nil {
		if errors.Is(err, application.ErrForbidden) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteOK(w, map[string]string{"message": "lotação atualizada com sucesso"})
}

func (h *Handlers) ResolveMeuAcesso(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	summary, err := h.service.ResolveUserAccess(r.Context(), identity)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteOK(w, summary)
}
