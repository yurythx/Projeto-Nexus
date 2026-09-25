package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yurythx/projeto-aurora/internal/modules/organizacao/domain"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// ============================================================
// LOCALIDADES
// ============================================================

func (r *PostgresRepository) ListLocalidades(ctx context.Context, onlyActive bool) ([]*domain.Localidade, error) {
	query := `
		SELECT id, nome, slug, tipo, grupo_ad, endereco, telefone, bairro, ativo, created_at, updated_at
		FROM localidades
	`
	if onlyActive {
		query += " WHERE ativo = true"
	}
	query += " ORDER BY tipo ASC, nome ASC"

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list localidades: %w", err)
	}
	defer rows.Close()

	locs := make([]*domain.Localidade, 0)
	locMap := make(map[uuid.UUID]*domain.Localidade)

	for rows.Next() {
		var l domain.Localidade
		if err := rows.Scan(
			&l.ID, &l.Nome, &l.Slug, &l.Tipo, &l.GrupoAD,
			&l.Endereco, &l.Telefone, &l.Bairro, &l.Ativo,
			&l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan localidade: %w", err)
		}
		l.Setores = make([]*domain.Setor, 0)
		locs = append(locs, &l)
		locMap[l.ID] = &l
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Buscar todos os setores dessas localidades
	setoresQuery := `
		SELECT id, localidade_id, nome, slug, tipo, grupo_ad, descricao, ativo, created_at, updated_at
		FROM setores
		WHERE ativo = true
		ORDER BY tipo ASC, nome ASC
	`
	sRows, err := r.pool.Query(ctx, setoresQuery)
	if err == nil {
		defer sRows.Close()
		for sRows.Next() {
			var s domain.Setor
			if err := sRows.Scan(
				&s.ID, &s.LocalidadeID, &s.Nome, &s.Slug, &s.Tipo,
				&s.GrupoAD, &s.Descricao, &s.Ativo, &s.CreatedAt, &s.UpdatedAt,
			); err == nil {
				if loc, ok := locMap[s.LocalidadeID]; ok {
					loc.Setores = append(loc.Setores, &s)
				}
			}
		}
	}

	return locs, nil
}

func (r *PostgresRepository) GetLocalidadeByID(ctx context.Context, id uuid.UUID) (*domain.Localidade, error) {
	query := `
		SELECT id, nome, slug, tipo, grupo_ad, endereco, telefone, bairro, ativo, created_at, updated_at
		FROM localidades
		WHERE id = $1
	`
	var l domain.Localidade
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&l.ID, &l.Nome, &l.Slug, &l.Tipo, &l.GrupoAD,
		&l.Endereco, &l.Telefone, &l.Bairro, &l.Ativo,
		&l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	setores, _ := r.ListSetoresByLocalidade(ctx, l.ID)
	l.Setores = setores
	return &l, nil
}

func (r *PostgresRepository) GetLocalidadeBySlug(ctx context.Context, slug string) (*domain.Localidade, error) {
	query := `
		SELECT id, nome, slug, tipo, grupo_ad, endereco, telefone, bairro, ativo, created_at, updated_at
		FROM localidades
		WHERE slug = $1
	`
	var l domain.Localidade
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&l.ID, &l.Nome, &l.Slug, &l.Tipo, &l.GrupoAD,
		&l.Endereco, &l.Telefone, &l.Bairro, &l.Ativo,
		&l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	setores, _ := r.ListSetoresByLocalidade(ctx, l.ID)
	l.Setores = setores
	return &l, nil
}

func (r *PostgresRepository) CreateLocalidade(ctx context.Context, loc *domain.Localidade) error {
	if loc.ID == uuid.Nil {
		loc.ID = uuid.New()
	}
	now := time.Now()
	loc.CreatedAt = now
	loc.UpdatedAt = now

	query := `
		INSERT INTO localidades (id, nome, slug, tipo, grupo_ad, endereco, telefone, bairro, ativo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		loc.ID, loc.Nome, loc.Slug, loc.Tipo, loc.GrupoAD,
		loc.Endereco, loc.Telefone, loc.Bairro, loc.Ativo,
		loc.CreatedAt, loc.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) UpdateLocalidade(ctx context.Context, loc *domain.Localidade) error {
	query := `
		UPDATE localidades
		SET nome = $1, slug = $2, tipo = $3, grupo_ad = $4, endereco = $5, telefone = $6, bairro = $7, ativo = $8, updated_at = now()
		WHERE id = $9
	`
	_, err := r.pool.Exec(ctx, query,
		loc.Nome, loc.Slug, loc.Tipo, loc.GrupoAD,
		loc.Endereco, loc.Telefone, loc.Bairro, loc.Ativo,
		loc.ID,
	)
	return err
}

func (r *PostgresRepository) DeleteLocalidade(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE localidades SET ativo = false, updated_at = now() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// ============================================================
// SETORES
// ============================================================

func (r *PostgresRepository) ListSetoresByLocalidade(ctx context.Context, localidadeID uuid.UUID) ([]*domain.Setor, error) {
	query := `
		SELECT id, localidade_id, nome, slug, tipo, grupo_ad, descricao, ativo, created_at, updated_at
		FROM setores
		WHERE localidade_id = $1 AND ativo = true
		ORDER BY tipo ASC, nome ASC
	`
	rows, err := r.pool.Query(ctx, query, localidadeID)
	if err != nil {
		return nil, fmt.Errorf("list setores: %w", err)
	}
	defer rows.Close()

	setores := make([]*domain.Setor, 0)
	for rows.Next() {
		var s domain.Setor
		if err := rows.Scan(
			&s.ID, &s.LocalidadeID, &s.Nome, &s.Slug, &s.Tipo,
			&s.GrupoAD, &s.Descricao, &s.Ativo, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		setores = append(setores, &s)
	}
	return setores, rows.Err()
}

func (r *PostgresRepository) GetSetorByID(ctx context.Context, id uuid.UUID) (*domain.Setor, error) {
	query := `
		SELECT id, localidade_id, nome, slug, tipo, grupo_ad, descricao, ativo, created_at, updated_at
		FROM setores
		WHERE id = $1
	`
	var s domain.Setor
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.LocalidadeID, &s.Nome, &s.Slug, &s.Tipo,
		&s.GrupoAD, &s.Descricao, &s.Ativo, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepository) CreateSetor(ctx context.Context, s *domain.Setor) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now

	query := `
		INSERT INTO setores (id, localidade_id, nome, slug, tipo, grupo_ad, descricao, ativo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		s.ID, s.LocalidadeID, s.Nome, s.Slug, s.Tipo,
		s.GrupoAD, s.Descricao, s.Ativo, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) UpdateSetor(ctx context.Context, s *domain.Setor) error {
	query := `
		UPDATE setores
		SET nome = $1, slug = $2, tipo = $3, grupo_ad = $4, descricao = $5, ativo = $6, updated_at = now()
		WHERE id = $7
	`
	_, err := r.pool.Exec(ctx, query,
		s.Nome, s.Slug, s.Tipo, s.GrupoAD, s.Descricao, s.Ativo, s.ID,
	)
	return err
}

func (r *PostgresRepository) DeleteSetor(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE setores SET ativo = false, updated_at = now() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// ============================================================
// PERFIS
// ============================================================

func (r *PostgresRepository) ListPerfis(ctx context.Context) ([]*domain.Perfil, error) {
	query := `
		SELECT id, nome, slug, descricao, grupo_ad, nivel, permissoes, ativo, created_at, updated_at
		FROM perfis
		ORDER BY nivel ASC, nome ASC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list perfis: %w", err)
	}
	defer rows.Close()

	perfis := make([]*domain.Perfil, 0)
	for rows.Next() {
		var p domain.Perfil
		if err := rows.Scan(
			&p.ID, &p.Nome, &p.Slug, &p.Descricao, &p.GrupoAD, &p.Nivel,
			&p.Permissoes, &p.Ativo, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		perfis = append(perfis, &p)
	}
	return perfis, rows.Err()
}

func (r *PostgresRepository) GetPerfilByID(ctx context.Context, id uuid.UUID) (*domain.Perfil, error) {
	query := `
		SELECT id, nome, slug, descricao, grupo_ad, nivel, permissoes, ativo, created_at, updated_at
		FROM perfis
		WHERE id = $1
	`
	var p domain.Perfil
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Nome, &p.Slug, &p.Descricao, &p.GrupoAD, &p.Nivel,
		&p.Permissoes, &p.Ativo, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepository) CreatePerfil(ctx context.Context, p *domain.Perfil) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now

	query := `
		INSERT INTO perfis (id, nome, slug, descricao, grupo_ad, nivel, permissoes, ativo, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		p.ID, p.Nome, p.Slug, p.Descricao, p.GrupoAD, p.Nivel,
		p.Permissoes, p.Ativo, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) UpdatePerfil(ctx context.Context, p *domain.Perfil) error {
	query := `
		UPDATE perfis
		SET nome = $1, slug = $2, descricao = $3, grupo_ad = $4, nivel = $5, permissoes = $6, ativo = $7, updated_at = now()
		WHERE id = $8
	`
	_, err := r.pool.Exec(ctx, query,
		p.Nome, p.Slug, p.Descricao, p.GrupoAD, p.Nivel,
		p.Permissoes, p.Ativo, p.ID,
	)
	return err
}

// ============================================================
// LOTAÇÃO DO USUÁRIO
// ============================================================

func (r *PostgresRepository) GetUsuarioLotacao(ctx context.Context, userID uuid.UUID) (*domain.UsuarioLotacao, error) {
	var lot domain.UsuarioLotacao
	lot.UserID = userID
	lot.Perfis = make([]*domain.Perfil, 0)
	lot.Localidades = make([]*domain.Localidade, 0)
	lot.Setores = make([]*domain.Setor, 0)

	// Dados do usuário
	userQuery := `SELECT username, display_name FROM users WHERE id = $1`
	err := r.pool.QueryRow(ctx, userQuery, userID).Scan(&lot.Username, &lot.DisplayName)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Perfis do usuário
	perfisQuery := `
		SELECT p.id, p.nome, p.slug, p.descricao, p.grupo_ad, p.nivel, p.permissoes, p.ativo, p.created_at, p.updated_at
		FROM perfis p
		JOIN usuario_perfis up ON up.perfil_id = p.id
		WHERE up.user_id = $1 AND p.ativo = true
	`
	pRows, err := r.pool.Query(ctx, perfisQuery, userID)
	if err == nil {
		defer pRows.Close()
		for pRows.Next() {
			var p domain.Perfil
			if err := pRows.Scan(
				&p.ID, &p.Nome, &p.Slug, &p.Descricao, &p.GrupoAD, &p.Nivel,
				&p.Permissoes, &p.Ativo, &p.CreatedAt, &p.UpdatedAt,
			); err == nil {
				lot.Perfis = append(lot.Perfis, &p)
			}
		}
	}

	// Localidades do usuário
	locsQuery := `
		SELECT l.id, l.nome, l.slug, l.tipo, l.grupo_ad, l.endereco, l.telefone, l.bairro, l.ativo, l.created_at, l.updated_at
		FROM localidades l
		JOIN usuario_localidades ul ON ul.localidade_id = l.id
		WHERE ul.user_id = $1 AND l.ativo = true
		ORDER BY ul.is_principal DESC, l.nome ASC
	`
	lRows, err := r.pool.Query(ctx, locsQuery, userID)
	if err == nil {
		defer lRows.Close()
		for lRows.Next() {
			var l domain.Localidade
			if err := lRows.Scan(
				&l.ID, &l.Nome, &l.Slug, &l.Tipo, &l.GrupoAD,
				&l.Endereco, &l.Telefone, &l.Bairro, &l.Ativo,
				&l.CreatedAt, &l.UpdatedAt,
			); err == nil {
				lot.Localidades = append(lot.Localidades, &l)
			}
		}
	}

	// Setores do usuário
	setoresQuery := `
		SELECT s.id, s.localidade_id, s.nome, s.slug, s.tipo, s.grupo_ad, s.descricao, s.ativo, s.created_at, s.updated_at
		FROM setores s
		JOIN usuario_setores us ON us.setor_id = s.id
		WHERE us.user_id = $1 AND s.ativo = true
	`
	sRows, err := r.pool.Query(ctx, setoresQuery, userID)
	if err == nil {
		defer sRows.Close()
		for sRows.Next() {
			var s domain.Setor
			if err := sRows.Scan(
				&s.ID, &s.LocalidadeID, &s.Nome, &s.Slug, &s.Tipo,
				&s.GrupoAD, &s.Descricao, &s.Ativo, &s.CreatedAt, &s.UpdatedAt,
			); err == nil {
				lot.Setores = append(lot.Setores, &s)
			}
		}
	}

	return &lot, nil
}

func (r *PostgresRepository) SetUsuarioLotacao(
	ctx context.Context,
	userID uuid.UUID,
	perfilIDs []uuid.UUID,
	localidadeIDs []uuid.UUID,
	setorIDs []uuid.UUID,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// 1. Atualizar Perfis
	if _, err := tx.Exec(ctx, `DELETE FROM usuario_perfis WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, pid := range perfilIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO usuario_perfis (user_id, perfil_id) VALUES ($1, $2)`, userID, pid); err != nil {
			return err
		}
	}

	// 2. Atualizar Localidades
	if _, err := tx.Exec(ctx, `DELETE FROM usuario_localidades WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for i, lid := range localidadeIDs {
		isPrincipal := (i == 0)
		if _, err := tx.Exec(ctx, `INSERT INTO usuario_localidades (user_id, localidade_id, is_principal) VALUES ($1, $2, $3)`, userID, lid, isPrincipal); err != nil {
			return err
		}
	}

	// 3. Atualizar Setores
	if _, err := tx.Exec(ctx, `DELETE FROM usuario_setores WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for i, sid := range setorIDs {
		isPrincipal := (i == 0)
		if _, err := tx.Exec(ctx, `INSERT INTO usuario_setores (user_id, setor_id, is_principal) VALUES ($1, $2, $3)`, userID, sid, isPrincipal); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

var _ domain.Repository = (*PostgresRepository)(nil)
