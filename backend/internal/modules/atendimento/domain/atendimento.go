package domain

import (
	"time"

	"github.com/google/uuid"
)

// Atendimento representa um registro oficial de atendimento socioassistencial
// realizado em qualquer uma das unidades da rede municipal da SEMPRAS.
type Atendimento struct {
	ID                  uuid.UUID      `json:"id"`
	ProtocolNumber      string         `json:"protocol_number"`
	ServiceSlug         string         `json:"service_slug"`
	Unit                string         `json:"unit"`
	CitizenName         string         `json:"citizen_name"`
	CitizenCPF          string         `json:"citizen_cpf,omitempty"`
	CitizenRG           string         `json:"citizen_rg,omitempty"`
	CitizenPhone        string         `json:"citizen_phone,omitempty"`
	CitizenNeighborhood string         `json:"citizen_neighborhood,omitempty"`
	CitizenAddress      string         `json:"citizen_address,omitempty"`
	AccessForm          string         `json:"access_form"`
	DemandType          string         `json:"demand_type"`
	RiskLevel           string         `json:"risk_level"`
	Vulnerabilities     map[string]any `json:"vulnerabilities"`
	TechnicalNotes      string         `json:"technical_notes,omitempty"`
	Referrals           map[string]any `json:"referrals"`
	Status              string         `json:"status"` // aguardando, em_atendimento, concluido, cancelado
	CreatedBy           *uuid.UUID     `json:"created_by,omitempty"`
	TechnicianID        *uuid.UUID     `json:"technician_id,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	FinishedAt          *time.Time     `json:"finished_at,omitempty"`
}

// CentroPopProntuario representa a ficha detalhada de acolhimento de cidadão
// em situação de rua ou itinerância atendido pelo Centro POP / Abordagem Social.
type CentroPopProntuario struct {
	ID                 uuid.UUID      `json:"id"`
	LegacyID           *int           `json:"legacy_id,omitempty"`
	Name               string         `json:"name"`
	PreferredName      string         `json:"preferred_name,omitempty"`
	Nickname           string         `json:"nickname,omitempty"`
	CPF                string         `json:"cpf,omitempty"`
	RG                 string         `json:"rg,omitempty"`
	DateBirth          *time.Time     `json:"date_birth,omitempty"`
	MotherName         string         `json:"mother_name,omitempty"`
	FatherName         string         `json:"father_name,omitempty"`
	Sex                string         `json:"sex,omitempty"`
	Gender             string         `json:"gender,omitempty"`
	Situation          string         `json:"situation"`
	TimeHomelessness   *float64       `json:"time_homelessness,omitempty"`
	ReasonHomelessness string         `json:"reason_homelessness,omitempty"`
	Address            string         `json:"address,omitempty"`
	DateOpened         time.Time      `json:"date_opened"`
	HealthNotes        map[string]any `json:"health_notes"`
	Photo              string         `json:"photo,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
}

// AtendimentoStats reúne as métricas operacionais do dia para a unidade ativa.
type AtendimentoStats struct {
	TodayTotal      int `json:"today_total"`
	TodayWaiting    int `json:"today_waiting"`
	TodayInProgress int `json:"today_in_progress"`
	TodayCompleted  int `json:"today_completed"`
}

// FilterParams define os critérios de busca e paginação de atendimentos.
type FilterParams struct {
	ServiceSlug string
	Unit        string
	Status      string
	Search      string
	CPF         string
	Page        int
	PageSize    int
}
