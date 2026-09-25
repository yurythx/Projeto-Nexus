package application

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/yurythx/projeto-aurora/internal/modules/organizacao/domain"
	"github.com/yurythx/projeto-aurora/internal/platform/auth"
)

var (
	ErrForbidden = errors.New("acesso negado: ação restrita a administradores")
	ErrNotFound  = errors.New("registro não encontrado")
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// Localidades
func (s *Service) ListLocalidades(ctx context.Context, onlyActive bool) ([]*domain.Localidade, error) {
	return s.repo.ListLocalidades(ctx, onlyActive)
}

func (s *Service) GetLocalidadeByID(ctx context.Context, id uuid.UUID) (*domain.Localidade, error) {
	return s.repo.GetLocalidadeByID(ctx, id)
}

func (s *Service) CreateLocalidade(ctx context.Context, identity auth.Identity, loc *domain.Localidade) error {
	if !identity.IsAdmin() {
		return ErrForbidden
	}
	return s.repo.CreateLocalidade(ctx, loc)
}

func (s *Service) UpdateLocalidade(ctx context.Context, identity auth.Identity, loc *domain.Localidade) error {
	if !identity.IsAdmin() {
		return ErrForbidden
	}
	return s.repo.UpdateLocalidade(ctx, loc)
}

func (s *Service) DeleteLocalidade(ctx context.Context, identity auth.Identity, id uuid.UUID) error {
	if !identity.IsAdmin() {
		return ErrForbidden
	}
	return s.repo.DeleteLocalidade(ctx, id)
}

// Setores
func (s *Service) ListSetoresByLocalidade(ctx context.Context, localidadeID uuid.UUID) ([]*domain.Setor, error) {
	return s.repo.ListSetoresByLocalidade(ctx, localidadeID)
}

func (s *Service) CreateSetor(ctx context.Context, identity auth.Identity, setor *domain.Setor) error {
	if !identity.IsAdmin() {
		return ErrForbidden
	}
	return s.repo.CreateSetor(ctx, setor)
}

func (s *Service) UpdateSetor(ctx context.Context, identity auth.Identity, setor *domain.Setor) error {
	if !identity.IsAdmin() {
		return ErrForbidden
	}
	return s.repo.UpdateSetor(ctx, setor)
}

func (s *Service) DeleteSetor(ctx context.Context, identity auth.Identity, id uuid.UUID) error {
	if !identity.IsAdmin() {
		return ErrForbidden
	}
	return s.repo.DeleteSetor(ctx, id)
}

// Perfis
func (s *Service) ListPerfis(ctx context.Context) ([]*domain.Perfil, error) {
	return s.repo.ListPerfis(ctx)
}

func (s *Service) CreatePerfil(ctx context.Context, identity auth.Identity, p *domain.Perfil) error {
	if !identity.IsAdmin() {
		return ErrForbidden
	}
	return s.repo.CreatePerfil(ctx, p)
}

func (s *Service) UpdatePerfil(ctx context.Context, identity auth.Identity, p *domain.Perfil) error {
	if !identity.IsAdmin() {
		return ErrForbidden
	}
	return s.repo.UpdatePerfil(ctx, p)
}

// Lotações do Usuário
func (s *Service) GetUsuarioLotacao(ctx context.Context, userID uuid.UUID) (*domain.UsuarioLotacao, error) {
	return s.repo.GetUsuarioLotacao(ctx, userID)
}

func (s *Service) SetUsuarioLotacao(
	ctx context.Context,
	identity auth.Identity,
	userID uuid.UUID,
	perfilIDs []uuid.UUID,
	localidadeIDs []uuid.UUID,
	setorIDs []uuid.UUID,
) error {
	if !identity.IsAdmin() {
		return ErrForbidden
	}
	return s.repo.SetUsuarioLotacao(ctx, userID, perfilIDs, localidadeIDs, setorIDs)
}

// UserAccessSummary resume o acesso do usuário logado (localidades permitidas, setores e perfil).
type UserAccessSummary struct {
	IsAdmin       bool                 `json:"is_admin"`
	Perfis        []*domain.Perfil     `json:"perfis"`
	Localidades   []*domain.Localidade `json:"localidades"`
	Setores       []*domain.Setor      `json:"setores"`
	AllowedSlugs  []string             `json:"allowed_slugs"`
	ActiveUnit    string               `json:"active_unit"`
	CanChangeUnit bool                 `json:"can_change_unit"`
}

func (s *Service) ResolveUserAccess(ctx context.Context, identity auth.Identity) (*UserAccessSummary, error) {
	allLocs, err := s.repo.ListLocalidades(ctx, true)
	if err != nil {
		return nil, err
	}

	summary := &UserAccessSummary{
		IsAdmin:       identity.IsAdmin(),
		Perfis:        make([]*domain.Perfil, 0),
		Localidades:   make([]*domain.Localidade, 0),
		Setores:       make([]*domain.Setor, 0),
		AllowedSlugs:  make([]string, 0),
		CanChangeUnit: false,
	}

	// 1. Administrador tem acesso total a TUDO
	if identity.IsAdmin() {
		summary.Localidades = allLocs
		summary.CanChangeUnit = true
		for _, l := range allLocs {
			summary.AllowedSlugs = append(summary.AllowedSlugs, l.Slug)
			summary.Setores = append(summary.Setores, l.Setores...)
		}
		summary.ActiveUnit = "Gestão Geral SEMPRAS"
		return summary, nil
	}

	// 2. Tentar buscar lotação do banco se subject for UUID
	userUUID, err := uuid.Parse(identity.Subject)
	if err == nil {
		lot, lotErr := s.repo.GetUsuarioLotacao(ctx, userUUID)
		if lotErr == nil && (len(lot.Localidades) > 0 || len(lot.Perfis) > 0) {
			summary.Perfis = lot.Perfis
			summary.Localidades = lot.Localidades
			summary.Setores = lot.Setores
			for _, l := range lot.Localidades {
				summary.AllowedSlugs = append(summary.AllowedSlugs, l.Slug)
			}
			if len(lot.Localidades) > 0 {
				summary.ActiveUnit = lot.Localidades[0].Nome
				summary.CanChangeUnit = len(lot.Localidades) > 1
			}
			return summary, nil
		}
	}

	// 3. Fallback: match dinâmico com grupos do Active Directory
	allowedLocMap := make(map[uuid.UUID]*domain.Localidade)
	for _, l := range allLocs {
		lNameLower := strings.ToLower(l.Nome)
		lSlugLower := strings.ToLower(l.Slug)
		lGroupLower := strings.ToLower(l.GrupoAD)

		hasMatch := false
		for _, g := range identity.Groups {
			gl := strings.ToLower(g)
			if lGroupLower != "" && strings.Contains(gl, lGroupLower) {
				hasMatch = true
				break
			}
			if strings.Contains(gl, lSlugLower) || strings.Contains(gl, lNameLower) {
				hasMatch = true
				break
			}
			if strings.HasPrefix(l.Slug, "cras-") {
				rawCras := strings.TrimPrefix(l.Slug, "cras-")
				if strings.Contains(gl, rawCras) {
					hasMatch = true
					break
				}
			}
			if l.Slug == "centro-pop" && (strings.Contains(gl, "pop") || strings.Contains(gl, "abordagem")) {
				hasMatch = true
				break
			}
			if l.Slug == "casa-da-mulher" && (strings.Contains(gl, "mulher") || strings.Contains(gl, "abrigo")) {
				hasMatch = true
				break
			}
			if l.Slug == "creas" && strings.Contains(gl, "creas") {
				hasMatch = true
				break
			}
		}

		if hasMatch {
			allowedLocMap[l.ID] = l
			summary.Localidades = append(summary.Localidades, l)
			summary.AllowedSlugs = append(summary.AllowedSlugs, l.Slug)
			summary.Setores = append(summary.Setores, l.Setores...)
		}
	}

	if len(summary.Localidades) > 0 {
		summary.ActiveUnit = summary.Localidades[0].Nome
		summary.CanChangeUnit = len(summary.Localidades) > 1
	} else {
		summary.ActiveUnit = "Unidade Socioassistencial"
	}

	return summary, nil
}
