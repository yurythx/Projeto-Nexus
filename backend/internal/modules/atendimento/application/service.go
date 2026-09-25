package application

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/yurythx/projeto-aurora/internal/modules/atendimento/domain"
	"github.com/yurythx/projeto-aurora/internal/platform/auth"
)

var (
	ErrForbiddenModule        = errors.New("acesso negado: seu perfil do AD não tem permissão para este serviço ou unidade")
	ErrModuleDisabled         = errors.New("este serviço socioassistencial está temporariamente desativado nas configurações do sistema")
	ErrReceptionistRestricted = errors.New("perfil recepcionista tem permissão apenas para recepção e encaminhamento à fila, não podendo evoluir parecer técnico ou prontuário")
)

// FeatureChecker verifica se uma flag de módulo está ativa
type FeatureChecker interface {
	Enabled(ctx context.Context, key string) bool
}

// AuthorizedService resume um serviço disponível para o usuário autenticado.
type AuthorizedService struct {
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	Category      string `json:"category"`
	UnitDefault   string `json:"unit_default"`
	CanChangeUnit bool   `json:"can_change_unit"`
}

type Service struct {
	repo  domain.Repository
	flags FeatureChecker
}

func NewService(repo domain.Repository, flags FeatureChecker) *Service {
	return &Service{
		repo:  repo,
		flags: flags,
	}
}

func (s *Service) isModuleEnabled(ctx context.Context, serviceSlug string) bool {
	if s.flags == nil {
		return true
	}
	var flagKey string
	switch serviceSlug {
	case "cras":
		flagKey = "module_cras_enabled"
	case "centro-pop", "centro_pop":
		flagKey = "module_centro_pop_enabled"
	case "creas":
		flagKey = "module_creas_enabled"
	case "casa-da-mulher", "casa_mulher":
		flagKey = "module_casa_mulher_enabled"
	case "conselho-tutelar", "conselho_tutelar":
		flagKey = "module_conselho_tutelar_enabled"
	case "cadunico-bolsa-familia", "cadunico":
		flagKey = "module_cadunico_enabled"
	case "beneficios-eventuais", "beneficios":
		flagKey = "module_beneficios_enabled"
	default:
		return true
	}

	return s.flags.Enabled(ctx, flagKey)
}

// CanAccessService valida se a identidade possui permissão pelo AD para operar o serviço e unidade.
func (s *Service) CanAccessService(identity auth.Identity, serviceSlug, unit string) bool {
	if identity.IsAdmin() {
		return true
	}

	switch serviceSlug {
	case "cras":
		if !identity.HasGroup("cras") {
			return false
		}
		// Se for técnico de um CRAS específico, restringe à sua unidade
		unitMatch := s.extractUserUnit(identity)
		if unitMatch != "" && unit != "" && !strings.EqualFold(unitMatch, unit) {
			// Permite se o usuário tiver papel de coordenação/gestão geral
			if !identity.HasGroup("gestao") && !identity.HasGroup("coord") {
				return false
			}
		}
		return true

	case "centro-pop", "centro_pop":
		return identity.HasGroup("pop") || identity.HasGroup("abordagem")

	case "creas":
		return identity.HasGroup("creas")

	case "casa-da-mulher", "casa_mulher":
		return identity.HasGroup("mulher") || identity.HasGroup("abrigo")

	case "conselho-tutelar", "conselho_tutelar":
		return identity.HasGroup("conselho") || identity.HasGroup("tutelar")

	case "cadunico-bolsa-familia", "cadunico":
		return identity.HasGroup("cadunico") || identity.HasGroup("bolsa") || identity.HasGroup("cras")

	case "beneficios-eventuais", "beneficios":
		return identity.HasGroup("beneficio") || identity.HasGroup("cras")

	default:
		return identity.HasGroup("assistencia") || identity.HasGroup("sempras")
	}
}

func (s *Service) extractUserUnit(identity auth.Identity) string {
	crasUnits := []string{
		"CRAS - Alfredo de Castro",
		"CRAS - Ana Carla",
		"CRAS - Central",
		"CRAS - Conjunto São José",
		"CRAS - Iguaçu",
		"CRAS - Luz Dyara",
		"CRAS - Padre Lothar",
		"CRAS - Rio Vermelho",
		"CRAS - Sagrada Família",
	}

	for _, g := range identity.Groups {
		for _, u := range crasUnits {
			cleanU := strings.ToLower(strings.ReplaceAll(u, "CRAS - ", ""))
			if strings.Contains(strings.ToLower(g), cleanU) {
				return u
			}
		}
		if strings.Contains(strings.ToLower(g), "centro pop") {
			return "Centro POP"
		}
		if strings.Contains(strings.ToLower(g), "casa da mulher") {
			return "Casa da Mulher"
		}
		if strings.Contains(strings.ToLower(g), "creas") {
			return "CREAS"
		}
	}
	return ""
}

// ResolveUserServices calcula quais serviços municipais o usuário pode ver na barra lateral e operar.
func (s *Service) ResolveUserServices(ctx context.Context, identity auth.Identity) ([]AuthorizedService, string) {
	isAdmin := identity.IsAdmin()
	userUnit := s.extractUserUnit(identity)
	if userUnit == "" {
		if isAdmin {
			userUnit = "Gestão Geral SEMPRAS"
		} else {
			userUnit = "Unidade Socioassistencial"
		}
	}

	allServices := []AuthorizedService{
		{Slug: "cras", Name: "Atendimento CRAS", Category: "Proteção Básica", UnitDefault: "CRAS - Central", CanChangeUnit: isAdmin},
		{Slug: "centro-pop", Name: "Centro POP & Abordagem", Category: "População em Situação de Rua", UnitDefault: "Centro POP", CanChangeUnit: isAdmin},
		{Slug: "creas", Name: "Atendimento CREAS", Category: "Média Complexidade", UnitDefault: "CREAS", CanChangeUnit: isAdmin},
		{Slug: "casa-da-mulher", Name: "Casa da Mulher", Category: "Proteção à Mulher", UnitDefault: "Casa da Mulher", CanChangeUnit: isAdmin},
		{Slug: "conselho-tutelar", Name: "Conselho Tutelar", Category: "Garantia de Direitos", UnitDefault: "Conselho Tutelar Central", CanChangeUnit: isAdmin},
		{Slug: "cadunico-bolsa-familia", Name: "Cadastro Único & Bolsa Família", Category: "Transferência de Renda", UnitDefault: "Central CadÚnico", CanChangeUnit: isAdmin},
		{Slug: "beneficios-eventuais", Name: "Benefícios Eventuais", Category: "Auxílios Emergenciais", UnitDefault: "Central de Benefícios", CanChangeUnit: isAdmin},
	}

	var allowed []AuthorizedService
	for _, srv := range allServices {
		// Checa se módulo está ativo
		if !s.isModuleEnabled(ctx, srv.Slug) {
			continue
		}

		if isAdmin || s.CanAccessService(identity, srv.Slug, "") {
			if strings.HasPrefix(userUnit, "CRAS") && srv.Slug == "cras" {
				srv.UnitDefault = userUnit
			}
			allowed = append(allowed, srv)
		}
	}

	return allowed, userUnit
}

func (s *Service) CreateAtendimento(ctx context.Context, identity auth.Identity, a *domain.Atendimento) (*domain.Atendimento, error) {
	if !s.isModuleEnabled(ctx, a.ServiceSlug) {
		return nil, ErrModuleDisabled
	}
	if !s.CanAccessService(identity, a.ServiceSlug, a.Unit) {
		return nil, ErrForbiddenModule
	}

	// Resolve usuário autor
	if actorID, _, ok := auth.ActorUUID(ctx); ok && actorID != nil {
		a.CreatedBy = actorID
	}

	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) IsReceptionist(identity auth.Identity) bool {
	if identity.IsAdmin() {
		return false
	}
	isRecep := false
	for _, g := range identity.Groups {
		gl := strings.ToLower(g)
		if strings.Contains(gl, "recepcao") || strings.Contains(gl, "recepção") || strings.Contains(gl, "atendente") {
			isRecep = true
			break
		}
	}
	if !isRecep {
		return false
	}
	for _, g := range identity.Groups {
		gl := strings.ToLower(g)
		if strings.Contains(gl, "tecnico") || strings.Contains(gl, "técnico") || strings.Contains(gl, "assistente_social") || strings.Contains(gl, "psicologo") {
			return false
		}
	}
	return true
}

func (s *Service) IsTechnician(identity auth.Identity) bool {
	if identity.IsAdmin() {
		return true
	}
	return !s.IsReceptionist(identity)
}

func (s *Service) GetAtendimento(ctx context.Context, identity auth.Identity, id uuid.UUID) (*domain.Atendimento, error) {
	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.isModuleEnabled(ctx, a.ServiceSlug) {
		return nil, ErrModuleDisabled
	}
	if !s.CanAccessService(identity, a.ServiceSlug, a.Unit) {
		return nil, ErrForbiddenModule
	}
	if s.IsReceptionist(identity) {
		a.TechnicalNotes = "🔒 Sigiloso: Conteúdo do parecer técnico restrito aos Técnicos / Assistentes Sociais credenciados."
		a.Referrals = nil
	}
	return a, nil
}

func (s *Service) UpdateAtendimento(ctx context.Context, identity auth.Identity, id uuid.UUID, status, notes string, vuln, ref map[string]any) (*domain.Atendimento, error) {
	if s.IsReceptionist(identity) {
		return nil, ErrReceptionistRestricted
	}

	a, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !s.isModuleEnabled(ctx, a.ServiceSlug) {
		return nil, ErrModuleDisabled
	}
	if !s.CanAccessService(identity, a.ServiceSlug, a.Unit) {
		return nil, ErrForbiddenModule
	}

	if status != "" {
		a.Status = status
	}
	if notes != "" {
		a.TechnicalNotes = notes
	}
	if vuln != nil {
		a.Vulnerabilities = vuln
	}
	if ref != nil {
		a.Referrals = ref
	}

	if actorID, _, ok := auth.ActorUUID(ctx); ok && actorID != nil {
		a.TechnicianID = actorID
	}

	if err := s.repo.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) ListAtendimentos(ctx context.Context, identity auth.Identity, filter domain.FilterParams) ([]*domain.Atendimento, int64, error) {
	if filter.ServiceSlug != "" && filter.ServiceSlug != "all" {
		if !s.isModuleEnabled(ctx, filter.ServiceSlug) {
			return nil, 0, ErrModuleDisabled
		}
		if !s.CanAccessService(identity, filter.ServiceSlug, filter.Unit) {
			return nil, 0, ErrForbiddenModule
		}
	} else if !identity.IsAdmin() {
		// Se não for admin e não especificou service_slug, força o primeiro serviço permitido
		services, unit := s.ResolveUserServices(ctx, identity)
		if len(services) == 0 {
			return nil, 0, nil
		}
		filter.ServiceSlug = services[0].Slug
		if filter.Unit == "" {
			filter.Unit = unit
		}
	}

	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	if s.IsReceptionist(identity) {
		for _, item := range items {
			item.TechnicalNotes = "🔒 Sigiloso: Restrito ao Técnico"
			item.Referrals = nil
		}
	}

	return items, total, nil
}

func (s *Service) SearchByCPF(ctx context.Context, identity auth.Identity, cpf string) ([]*domain.Atendimento, error) {
	items, err := s.repo.SearchByCPF(ctx, cpf)
	if err != nil {
		return nil, err
	}
	if s.IsReceptionist(identity) {
		for _, item := range items {
			item.TechnicalNotes = "🔒 Sigiloso: Restrito ao Técnico"
			item.Referrals = nil
		}
	}
	return items, nil
}

func (s *Service) GetStats(ctx context.Context, identity auth.Identity, serviceSlug, unit string) (*domain.AtendimentoStats, error) {
	if serviceSlug != "" && !identity.IsAdmin() {
		if !s.CanAccessService(identity, serviceSlug, unit) {
			return nil, ErrForbiddenModule
		}
	}
	return s.repo.GetStats(ctx, serviceSlug, unit)
}

// Centro POP
func (s *Service) CreateProntuario(ctx context.Context, identity auth.Identity, p *domain.CentroPopProntuario) (*domain.CentroPopProntuario, error) {
	if s.IsReceptionist(identity) {
		return nil, ErrForbiddenModule
	}
	if !s.isModuleEnabled(ctx, "centro-pop") {
		return nil, ErrModuleDisabled
	}
	if !s.CanAccessService(identity, "centro-pop", "Centro POP") {
		return nil, ErrForbiddenModule
	}
	if err := s.repo.CreateProntuario(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) GetProntuario(ctx context.Context, identity auth.Identity, id uuid.UUID) (*domain.CentroPopProntuario, error) {
	if s.IsReceptionist(identity) {
		return nil, ErrForbiddenModule
	}
	if !s.isModuleEnabled(ctx, "centro-pop") {
		return nil, ErrModuleDisabled
	}
	if !s.CanAccessService(identity, "centro-pop", "Centro POP") {
		return nil, ErrForbiddenModule
	}
	return s.repo.GetProntuarioByID(ctx, id)
}

func (s *Service) ListProntuarios(ctx context.Context, identity auth.Identity, search string, page, pageSize int) ([]*domain.CentroPopProntuario, int64, error) {
	if s.IsReceptionist(identity) {
		return nil, 0, ErrForbiddenModule
	}
	if !s.isModuleEnabled(ctx, "centro-pop") {
		return nil, 0, ErrModuleDisabled
	}
	if !s.CanAccessService(identity, "centro-pop", "Centro POP") {
		return nil, 0, ErrForbiddenModule
	}
	return s.repo.ListProntuarios(ctx, search, page, pageSize)
}
