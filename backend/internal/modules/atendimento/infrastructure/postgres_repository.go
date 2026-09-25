package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yurythx/projeto-aurora/internal/modules/atendimento/domain"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

var _ domain.Repository = (*PostgresRepository)(nil)

func (r *PostgresRepository) nextProtocolNumber(ctx context.Context) (string, error) {
	var seq int64
	err := r.pool.QueryRow(ctx, "SELECT nextval('atendimentos_protocol_seq')").Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("atendimento: generate protocol sequence: %w", err)
	}
	return fmt.Sprintf("ATD-%d-%05d", time.Now().Year(), seq), nil
}

func (r *PostgresRepository) Create(ctx context.Context, a *domain.Atendimento) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	if a.ProtocolNumber == "" {
		proto, err := r.nextProtocolNumber(ctx)
		if err != nil {
			return err
		}
		a.ProtocolNumber = proto
	}
	if a.Status == "" {
		a.Status = "aguardando"
	}
	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	vulnJSON, err := json.Marshal(a.Vulnerabilities)
	if err != nil {
		vulnJSON = []byte("{}")
	}
	refJSON, err := json.Marshal(a.Referrals)
	if err != nil {
		refJSON = []byte("{}")
	}

	cleanCPF := strings.TrimSpace(a.CitizenCPF)

	const q = `
		INSERT INTO atendimentos (
			id, protocol_number, service_slug, unit, citizen_name, citizen_cpf, citizen_rg,
			citizen_phone, citizen_neighborhood, citizen_address, access_form, demand_type,
			risk_level, vulnerabilities, technical_notes, referrals, status, created_by,
			technician_id, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
		)
	`
	_, err = r.pool.Exec(ctx, q,
		a.ID, a.ProtocolNumber, a.ServiceSlug, a.Unit, a.CitizenName, cleanCPF, a.CitizenRG,
		a.CitizenPhone, a.CitizenNeighborhood, a.CitizenAddress, a.AccessForm, a.DemandType,
		a.RiskLevel, vulnJSON, a.TechnicalNotes, refJSON, a.Status, a.CreatedBy, a.TechnicianID,
		a.CreatedAt, a.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("atendimento: insert: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Atendimento, error) {
	const q = `
		SELECT id, protocol_number, service_slug, unit, citizen_name, COALESCE(citizen_cpf, ''),
		       COALESCE(citizen_rg, ''), COALESCE(citizen_phone, ''), COALESCE(citizen_neighborhood, ''),
		       COALESCE(citizen_address, ''), access_form, demand_type, risk_level,
		       COALESCE(vulnerabilities, '{}'::jsonb), COALESCE(technical_notes, ''),
		       COALESCE(referrals, '{}'::jsonb), status, created_by, technician_id,
		       created_at, updated_at, finished_at
		FROM atendimentos
		WHERE id = $1
	`
	var a domain.Atendimento
	var vulnBytes, refBytes []byte
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&a.ID, &a.ProtocolNumber, &a.ServiceSlug, &a.Unit, &a.CitizenName, &a.CitizenCPF,
		&a.CitizenRG, &a.CitizenPhone, &a.CitizenNeighborhood, &a.CitizenAddress, &a.AccessForm,
		&a.DemandType, &a.RiskLevel, &vulnBytes, &a.TechnicalNotes, &refBytes, &a.Status,
		&a.CreatedBy, &a.TechnicianID, &a.CreatedAt, &a.UpdatedAt, &a.FinishedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("atendimento not found")
		}
		return nil, fmt.Errorf("atendimento: get by id: %w", err)
	}
	_ = json.Unmarshal(vulnBytes, &a.Vulnerabilities)
	_ = json.Unmarshal(refBytes, &a.Referrals)
	return &a, nil
}

func (r *PostgresRepository) Update(ctx context.Context, a *domain.Atendimento) error {
	a.UpdatedAt = time.Now()
	if a.Status == "concluido" && a.FinishedAt == nil {
		now := time.Now()
		a.FinishedAt = &now
	}

	vulnJSON, err := json.Marshal(a.Vulnerabilities)
	if err != nil {
		vulnJSON = []byte("{}")
	}
	refJSON, err := json.Marshal(a.Referrals)
	if err != nil {
		refJSON = []byte("{}")
	}

	const q = `
		UPDATE atendimentos SET
			status = $1,
			technical_notes = $2,
			referrals = $3,
			vulnerabilities = $4,
			technician_id = COALESCE($5, technician_id),
			updated_at = $6,
			finished_at = $7
		WHERE id = $8
	`
	tag, err := r.pool.Exec(ctx, q,
		a.Status, a.TechnicalNotes, refJSON, vulnJSON, a.TechnicianID, a.UpdatedAt, a.FinishedAt, a.ID,
	)
	if err != nil {
		return fmt.Errorf("atendimento: update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("atendimento not found for update")
	}
	return nil
}

func normalizeSlugCondition(slug string, argIdx int) (string, []any, int) {
	switch slug {
	case "centro_pop", "centro-pop":
		return fmt.Sprintf("service_slug IN ($%d, $%d)", argIdx, argIdx+1), []any{"centro_pop", "centro-pop"}, argIdx + 2
	case "casa_mulher", "casa-da-mulher":
		return fmt.Sprintf("service_slug IN ($%d, $%d)", argIdx, argIdx+1), []any{"casa_mulher", "casa-da-mulher"}, argIdx + 2
	case "conselho_tutelar", "conselho-tutelar":
		return fmt.Sprintf("service_slug IN ($%d, $%d)", argIdx, argIdx+1), []any{"conselho_tutelar", "conselho-tutelar"}, argIdx + 2
	case "cadunico", "cadunico-bolsa-familia":
		return fmt.Sprintf("service_slug IN ($%d, $%d)", argIdx, argIdx+1), []any{"cadunico", "cadunico-bolsa-familia"}, argIdx + 2
	case "beneficios", "beneficios-eventuais":
		return fmt.Sprintf("service_slug IN ($%d, $%d)", argIdx, argIdx+1), []any{"beneficios", "beneficios-eventuais"}, argIdx + 2
	default:
		return fmt.Sprintf("service_slug = $%d", argIdx), []any{slug}, argIdx + 1
	}
}

func (r *PostgresRepository) List(ctx context.Context, filter domain.FilterParams) ([]*domain.Atendimento, int64, error) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.ServiceSlug != "" && filter.ServiceSlug != "all" {
		cond, slugArgs, nextIdx := normalizeSlugCondition(filter.ServiceSlug, argIdx)
		conditions = append(conditions, cond)
		args = append(args, slugArgs...)
		argIdx = nextIdx
	}

	if filter.Unit != "" && filter.Unit != "all" {
		conditions = append(conditions, fmt.Sprintf("unit ILIKE $%d", argIdx))
		args = append(args, "%"+filter.Unit+"%")
		argIdx++
	}

	if filter.Status != "" && filter.Status != "all" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}

	if filter.Search != "" {
		clean := strings.ReplaceAll(strings.ReplaceAll(filter.Search, ".", ""), "-", "")
		conditions = append(conditions, fmt.Sprintf("(citizen_name ILIKE $%d OR citizen_cpf ILIKE $%d OR protocol_number ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+clean+"%")
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM atendimentos %s", whereClause)
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("atendimento: count list: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT id, protocol_number, service_slug, unit, citizen_name, COALESCE(citizen_cpf, ''),
		       COALESCE(citizen_rg, ''), COALESCE(citizen_phone, ''), COALESCE(citizen_neighborhood, ''),
		       COALESCE(citizen_address, ''), access_form, demand_type, risk_level,
		       COALESCE(vulnerabilities, '{}'::jsonb), COALESCE(technical_notes, ''),
		       COALESCE(referrals, '{}'::jsonb), status, created_by, technician_id,
		       created_at, updated_at, finished_at
		FROM atendimentos
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("atendimento: query list: %w", err)
	}
	defer rows.Close()

	var list []*domain.Atendimento
	for rows.Next() {
		var a domain.Atendimento
		var vulnBytes, refBytes []byte
		err := rows.Scan(
			&a.ID, &a.ProtocolNumber, &a.ServiceSlug, &a.Unit, &a.CitizenName, &a.CitizenCPF,
			&a.CitizenRG, &a.CitizenPhone, &a.CitizenNeighborhood, &a.CitizenAddress, &a.AccessForm,
			&a.DemandType, &a.RiskLevel, &vulnBytes, &a.TechnicalNotes, &refBytes, &a.Status,
			&a.CreatedBy, &a.TechnicianID, &a.CreatedAt, &a.UpdatedAt, &a.FinishedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("atendimento: scan list row: %w", err)
		}
		_ = json.Unmarshal(vulnBytes, &a.Vulnerabilities)
		_ = json.Unmarshal(refBytes, &a.Referrals)
		list = append(list, &a)
	}

	return list, total, nil
}

func (r *PostgresRepository) SearchByCPF(ctx context.Context, cpf string) ([]*domain.Atendimento, error) {
	clean := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(cpf, ".", ""), "-", ""), " ", "")
	if clean == "" {
		return nil, nil
	}

	const q = `
		SELECT id, protocol_number, service_slug, unit, citizen_name, COALESCE(citizen_cpf, ''),
		       COALESCE(citizen_rg, ''), COALESCE(citizen_phone, ''), COALESCE(citizen_neighborhood, ''),
		       COALESCE(citizen_address, ''), access_form, demand_type, risk_level,
		       COALESCE(vulnerabilities, '{}'::jsonb), COALESCE(technical_notes, ''),
		       COALESCE(referrals, '{}'::jsonb), status, created_by, technician_id,
		       created_at, updated_at, finished_at
		FROM atendimentos
		WHERE REPLACE(REPLACE(REPLACE(citizen_cpf, '.', ''), '-', ''), ' ', '') = $1
		ORDER BY created_at DESC
		LIMIT 10
	`
	rows, err := r.pool.Query(ctx, q, clean)
	if err != nil {
		return nil, fmt.Errorf("atendimento: search by cpf: %w", err)
	}
	defer rows.Close()

	var list []*domain.Atendimento
	for rows.Next() {
		var a domain.Atendimento
		var vulnBytes, refBytes []byte
		err := rows.Scan(
			&a.ID, &a.ProtocolNumber, &a.ServiceSlug, &a.Unit, &a.CitizenName, &a.CitizenCPF,
			&a.CitizenRG, &a.CitizenPhone, &a.CitizenNeighborhood, &a.CitizenAddress, &a.AccessForm,
			&a.DemandType, &a.RiskLevel, &vulnBytes, &a.TechnicalNotes, &refBytes, &a.Status,
			&a.CreatedBy, &a.TechnicianID, &a.CreatedAt, &a.UpdatedAt, &a.FinishedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("atendimento: scan cpf match: %w", err)
		}
		_ = json.Unmarshal(vulnBytes, &a.Vulnerabilities)
		_ = json.Unmarshal(refBytes, &a.Referrals)
		list = append(list, &a)
	}
	return list, nil
}

func (r *PostgresRepository) GetStats(ctx context.Context, serviceSlug, unit string) (*domain.AtendimentoStats, error) {
	var conditions = []string{"created_at >= CURRENT_DATE"}
	var args []any
	idx := 1

	if serviceSlug != "" && serviceSlug != "all" {
		cond, slugArgs, nextIdx := normalizeSlugCondition(serviceSlug, idx)
		conditions = append(conditions, cond)
		args = append(args, slugArgs...)
		idx = nextIdx
	}
	if unit != "" && unit != "all" {
		conditions = append(conditions, fmt.Sprintf("unit ILIKE $%d", idx))
		args = append(args, "%"+unit+"%")
		idx++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'aguardando') as waiting,
			COUNT(*) FILTER (WHERE status = 'em_atendimento') as in_progress,
			COUNT(*) FILTER (WHERE status = 'concluido') as completed
		FROM atendimentos
		%s
	`, whereClause)

	var s domain.AtendimentoStats
	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&s.TodayTotal, &s.TodayWaiting, &s.TodayInProgress, &s.TodayCompleted,
	)
	if err != nil {
		return nil, fmt.Errorf("atendimento: get stats: %w", err)
	}
	return &s, nil
}

// Centro POP Prontuários
func (r *PostgresRepository) CreateProntuario(ctx context.Context, p *domain.CentroPopProntuario) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	if p.DateOpened.IsZero() {
		p.DateOpened = time.Now()
	}

	healthJSON, err := json.Marshal(p.HealthNotes)
	if err != nil {
		healthJSON = []byte("{}")
	}

	const q = `
		INSERT INTO centro_pop_prontuarios (
			id, legacy_id, name, preferred_name, nickname, cpf, rg, date_birth,
			mother_name, father_name, sex, gender, situation, time_homelessness,
			reason_homelessness, address, date_opened, health_notes, photo, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
	`
	_, err = r.pool.Exec(ctx, q,
		p.ID, p.LegacyID, p.Name, p.PreferredName, p.Nickname, p.CPF, p.RG, p.DateBirth,
		p.MotherName, p.FatherName, p.Sex, p.Gender, p.Situation, p.TimeHomelessness,
		p.ReasonHomelessness, p.Address, p.DateOpened, healthJSON, p.Photo, p.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("atendimento: insert centro pop prontuario: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetProntuarioByID(ctx context.Context, id uuid.UUID) (*domain.CentroPopProntuario, error) {
	const q = `
		SELECT id, legacy_id, name, COALESCE(preferred_name, ''), COALESCE(nickname, ''),
		       COALESCE(cpf, ''), COALESCE(rg, ''), date_birth, COALESCE(mother_name, ''),
		       COALESCE(father_name, ''), COALESCE(sex, ''), COALESCE(gender, ''),
		       COALESCE(situation, 'rua'), time_homelessness, COALESCE(reason_homelessness, ''),
		       COALESCE(address, ''), date_opened, COALESCE(health_notes, '{}'::jsonb),
		       COALESCE(photo, ''), created_at
		FROM centro_pop_prontuarios
		WHERE id = $1
	`
	var p domain.CentroPopProntuario
	var healthBytes []byte
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.LegacyID, &p.Name, &p.PreferredName, &p.Nickname, &p.CPF, &p.RG,
		&p.DateBirth, &p.MotherName, &p.FatherName, &p.Sex, &p.Gender, &p.Situation,
		&p.TimeHomelessness, &p.ReasonHomelessness, &p.Address, &p.DateOpened,
		&healthBytes, &p.Photo, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("prontuario not found")
		}
		return nil, fmt.Errorf("atendimento: get prontuario by id: %w", err)
	}
	_ = json.Unmarshal(healthBytes, &p.HealthNotes)
	return &p, nil
}

func (r *PostgresRepository) ListProntuarios(ctx context.Context, search string, page, pageSize int) ([]*domain.CentroPopProntuario, int64, error) {
	where := ""
	var args []any
	if search != "" {
		clean := strings.TrimSpace(search)
		where = "WHERE name ILIKE $1 OR nickname ILIKE $1 OR cpf ILIKE $1"
		args = append(args, "%"+clean+"%")
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int64
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM centro_pop_prontuarios "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("atendimento: count prontuarios: %w", err)
	}

	limitIdx := len(args) + 1
	offsetIdx := len(args) + 2
	args = append(args, pageSize, offset)

	q := fmt.Sprintf(`
		SELECT id, legacy_id, name, COALESCE(preferred_name, ''), COALESCE(nickname, ''),
		       COALESCE(cpf, ''), COALESCE(rg, ''), date_birth, COALESCE(mother_name, ''),
		       COALESCE(father_name, ''), COALESCE(sex, ''), COALESCE(gender, ''),
		       COALESCE(situation, 'rua'), time_homelessness, COALESCE(reason_homelessness, ''),
		       COALESCE(address, ''), date_opened, COALESCE(health_notes, '{}'::jsonb),
		       COALESCE(photo, ''), created_at
		FROM centro_pop_prontuarios
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, limitIdx, offsetIdx)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("atendimento: query prontuarios: %w", err)
	}
	defer rows.Close()

	var list []*domain.CentroPopProntuario
	for rows.Next() {
		var p domain.CentroPopProntuario
		var healthBytes []byte
		err := rows.Scan(
			&p.ID, &p.LegacyID, &p.Name, &p.PreferredName, &p.Nickname, &p.CPF, &p.RG,
			&p.DateBirth, &p.MotherName, &p.FatherName, &p.Sex, &p.Gender, &p.Situation,
			&p.TimeHomelessness, &p.ReasonHomelessness, &p.Address, &p.DateOpened,
			&healthBytes, &p.Photo, &p.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("atendimento: scan prontuario row: %w", err)
		}
		_ = json.Unmarshal(healthBytes, &p.HealthNotes)
		list = append(list, &p)
	}
	return list, total, nil
}
