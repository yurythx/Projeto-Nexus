package transport

import (
	"time"

	"github.com/google/uuid"
	"github.com/yurythx/projeto-aurora/internal/modules/atendimento/domain"
)

type CreateAtendimentoRequest struct {
	ServiceSlug         string         `json:"service_slug" validate:"required"`
	Unit                string         `json:"unit" validate:"required"`
	CitizenName         string         `json:"citizen_name" validate:"required"`
	CitizenCPF          string         `json:"citizen_cpf"`
	CitizenRG           string         `json:"citizen_rg"`
	CitizenPhone        string         `json:"citizen_phone"`
	CitizenNeighborhood string         `json:"citizen_neighborhood"`
	CitizenAddress      string         `json:"citizen_address"`
	AccessForm          string         `json:"access_form"`
	DemandType          string         `json:"demand_type"`
	RiskLevel           string         `json:"risk_level"`
	Vulnerabilities     map[string]any `json:"vulnerabilities"`
	TechnicalNotes      string         `json:"technical_notes"`
	Referrals           map[string]any `json:"referrals"`
	Status              string         `json:"status"`
}

type UpdateAtendimentoRequest struct {
	Status          string         `json:"status"`
	TechnicalNotes  string         `json:"technical_notes"`
	Vulnerabilities map[string]any `json:"vulnerabilities"`
	Referrals       map[string]any `json:"referrals"`
}

type CreateProntuarioRequest struct {
	LegacyID           *int           `json:"legacy_id"`
	Name               string         `json:"name" validate:"required"`
	PreferredName      string         `json:"preferred_name"`
	Nickname           string         `json:"nickname"`
	CPF                string         `json:"cpf"`
	RG                 string         `json:"rg"`
	DateBirth          *string        `json:"date_birth"`
	MotherName         string         `json:"mother_name"`
	FatherName         string         `json:"father_name"`
	Sex                string         `json:"sex"`
	Gender             string         `json:"gender"`
	Situation          string         `json:"situation"`
	TimeHomelessness   *float64       `json:"time_homelessness"`
	ReasonHomelessness string         `json:"reason_homelessness"`
	Address            string         `json:"address"`
	HealthNotes        map[string]any `json:"health_notes"`
	Photo              string         `json:"photo"`
}

func (req CreateAtendimentoRequest) toDomain() *domain.Atendimento {
	status := req.Status
	if status == "" {
		status = "aguardando"
	}
	access := req.AccessForm
	if access == "" {
		access = "Demanda Espontânea"
	}
	demand := req.DemandType
	if demand == "" {
		demand = "Acolhida e Atendimento Geral"
	}
	risk := req.RiskLevel
	if risk == "" {
		risk = "Baixo"
	}

	return &domain.Atendimento{
		ID:                  uuid.New(),
		ServiceSlug:         req.ServiceSlug,
		Unit:                req.Unit,
		CitizenName:         req.CitizenName,
		CitizenCPF:          req.CitizenCPF,
		CitizenRG:           req.CitizenRG,
		CitizenPhone:        req.CitizenPhone,
		CitizenNeighborhood: req.CitizenNeighborhood,
		CitizenAddress:      req.CitizenAddress,
		AccessForm:          access,
		DemandType:          demand,
		RiskLevel:           risk,
		Vulnerabilities:     req.Vulnerabilities,
		TechnicalNotes:      req.TechnicalNotes,
		Referrals:           req.Referrals,
		Status:              status,
	}
}

func (req CreateProntuarioRequest) toDomain() *domain.CentroPopProntuario {
	p := &domain.CentroPopProntuario{
		ID:                 uuid.New(),
		LegacyID:           req.LegacyID,
		Name:               req.Name,
		PreferredName:      req.PreferredName,
		Nickname:           req.Nickname,
		CPF:                req.CPF,
		RG:                 req.RG,
		MotherName:         req.MotherName,
		FatherName:         req.FatherName,
		Sex:                req.Sex,
		Gender:             req.Gender,
		Situation:          req.Situation,
		TimeHomelessness:   req.TimeHomelessness,
		ReasonHomelessness: req.ReasonHomelessness,
		Address:            req.Address,
		DateOpened:         time.Now(),
		HealthNotes:        req.HealthNotes,
		Photo:              req.Photo,
		CreatedAt:          time.Now(),
	}

	if req.DateBirth != nil && *req.DateBirth != "" {
		if t, err := time.Parse("2006-01-02", *req.DateBirth); err == nil {
			p.DateBirth = &t
		}
	}
	if p.Situation == "" {
		p.Situation = "rua"
	}

	return p
}
