package transport

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	apperrors "github.com/yurythx/projeto-aurora/internal/domain/errors"
	"github.com/yurythx/projeto-aurora/internal/domain/pagination"
	"github.com/yurythx/projeto-aurora/internal/modules/atendimento/application"
	"github.com/yurythx/projeto-aurora/internal/modules/atendimento/domain"
	"github.com/yurythx/projeto-aurora/internal/platform/auth"
	"github.com/yurythx/projeto-aurora/pkg/httputil"
)

type Handlers struct {
	service     *application.Service
	logger      *slog.Logger
	maxPageSize int
}

func NewHandlers(service *application.Service, logger *slog.Logger, maxPageSize int) *Handlers {
	return &Handlers{
		service:     service,
		logger:      logger,
		maxPageSize: maxPageSize,
	}
}

func (h *Handlers) getIdentity(r *http.Request) (auth.Identity, error) {
	identity, ok := auth.IdentityFromContext(r.Context())
	if !ok {
		return auth.Identity{}, apperrors.Unauthorized("authentication required")
	}
	return identity, nil
}

func (h *Handlers) ResolveMyServices(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	services, defaultUnit := h.service.ResolveUserServices(r.Context(), identity)
	isRecepcao := h.service.IsReceptionist(identity)
	isTecnico := h.service.IsTechnician(identity)
	httputil.WriteOK(w, map[string]any{
		"services":        services,
		"default_unit":    defaultUnit,
		"is_admin":        identity.IsAdmin(),
		"is_technician":   isTecnico,
		"is_receptionist": isRecepcao,
		"groups":          identity.Groups,
		"roles":           identity.Roles,
	})
}

func (h *Handlers) CreateAtendimento(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	var req CreateAtendimentoRequest
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}
	if err := httputil.Validate(req); err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	atendimento, err := h.service.CreateAtendimento(r.Context(), identity, req.toDomain())
	if err != nil {
		if errors.Is(err, application.ErrForbiddenModule) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		if errors.Is(err, application.ErrModuleDisabled) {
			httputil.WriteError(w, r, h.logger, apperrors.BadRequest(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteCreated(w, atendimento)
}

func (h *Handlers) GetAtendimento(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, r, h.logger, apperrors.BadRequest("invalid atendimento id"))
		return
	}

	atendimento, err := h.service.GetAtendimento(r.Context(), identity, id)
	if err != nil {
		if errors.Is(err, application.ErrForbiddenModule) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteOK(w, atendimento)
}

func (h *Handlers) UpdateAtendimento(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, r, h.logger, apperrors.BadRequest("invalid atendimento id"))
		return
	}

	var req UpdateAtendimentoRequest
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	atendimento, err := h.service.UpdateAtendimento(r.Context(), identity, id, req.Status, req.TechnicalNotes, req.Vulnerabilities, req.Referrals)
	if err != nil {
		if errors.Is(err, application.ErrForbiddenModule) || errors.Is(err, application.ErrReceptionistRestricted) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteOK(w, atendimento)
}

func (h *Handlers) ListAtendimentos(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}

	serviceSlug := r.URL.Query().Get("service_slug")
	if serviceSlug == "" {
		serviceSlug = r.URL.Query().Get("service")
	}

	filter := domain.FilterParams{
		ServiceSlug: serviceSlug,
		Unit:        r.URL.Query().Get("unit"),
		Status:      r.URL.Query().Get("status"),
		Search:      r.URL.Query().Get("search"),
		CPF:         r.URL.Query().Get("cpf"),
		Page:        page,
		PageSize:    pageSize,
	}

	items, total, err := h.service.ListAtendimentos(r.Context(), identity, filter)
	if err != nil {
		if errors.Is(err, application.ErrForbiddenModule) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	meta := pagination.Meta{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalItems: total,
		TotalPages: int((total + int64(filter.PageSize) - 1) / int64(filter.PageSize)),
	}
	httputil.WriteOKWithMeta(w, items, meta)
}

func (h *Handlers) SearchByCPF(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	cpf := chi.URLParam(r, "cpf")
	items, err := h.service.SearchByCPF(r.Context(), identity, cpf)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteOK(w, map[string]any{
		"found":       len(items) > 0,
		"cpf":         cpf,
		"history":     items,
		"last_record": func() any { if len(items) > 0 { return items[0] }; return nil }(),
	})
}

func (h *Handlers) GetStats(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	serviceSlug := r.URL.Query().Get("service_slug")
	if serviceSlug == "" {
		serviceSlug = r.URL.Query().Get("service")
	}
	unit := r.URL.Query().Get("unit")

	stats, err := h.service.GetStats(r.Context(), identity, serviceSlug, unit)
	if err != nil {
		if errors.Is(err, application.ErrForbiddenModule) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteOK(w, stats)
}

// Centro POP
func (h *Handlers) CreateProntuario(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	var req CreateProntuarioRequest
	if err := httputil.DecodeJSON(w, r, &req); err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}
	if err := httputil.Validate(req); err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	prontuario, err := h.service.CreateProntuario(r.Context(), identity, req.toDomain())
	if err != nil {
		if errors.Is(err, application.ErrForbiddenModule) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteCreated(w, prontuario)
}

func (h *Handlers) GetProntuario(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputil.WriteError(w, r, h.logger, apperrors.BadRequest("invalid prontuario id"))
		return
	}

	p, err := h.service.GetProntuario(r.Context(), identity, id)
	if err != nil {
		if errors.Is(err, application.ErrForbiddenModule) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	httputil.WriteOK(w, p)
}

func (h *Handlers) ListProntuarios(w http.ResponseWriter, r *http.Request) {
	identity, err := h.getIdentity(r)
	if err != nil {
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	search := r.URL.Query().Get("search")

	items, total, err := h.service.ListProntuarios(r.Context(), identity, search, page, pageSize)
	if err != nil {
		if errors.Is(err, application.ErrForbiddenModule) {
			httputil.WriteError(w, r, h.logger, apperrors.Forbidden(err.Error()))
			return
		}
		httputil.WriteError(w, r, h.logger, err)
		return
	}

	if pageSize <= 0 {
		pageSize = 20
	}
	meta := pagination.Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: total,
		TotalPages: int((total + int64(pageSize) - 1) / int64(pageSize)),
	}
	httputil.WriteOKWithMeta(w, items, meta)
}
