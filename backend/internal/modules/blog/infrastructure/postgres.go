// Package infrastructure implementa o repositório do Blog em PostgreSQL.
package infrastructure

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/yurythx/projeto-nexus/internal/domain/pagination"
	"github.com/yurythx/projeto-nexus/internal/modules/blog/domain"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
)

// Repository implementa domain.Repository.
type Repository struct{}

// NewRepository cria o repositório.
func NewRepository() *Repository { return &Repository{} }

var _ domain.Repository = (*Repository)(nil)

const cols = `p.id, p.slug, p.title, p.summary, p.body, p.cover_object_key, p.kind, p.status, p.pinned,
	p.author_id, COALESCE(NULLIF(u.display_name,''), u.username, ''), p.published_at, p.created_at, p.updated_at, p.unidade_id`

const from = ` FROM blog_posts p LEFT JOIN users u ON u.id = p.author_id `

// scan lê as colunas de cols (na mesma ordem) e, depois delas, extra —
// fonte única da ordem das colunas (a busca lê o rank como extra).
func scan(row interface{ Scan(...any) error }, extra ...any) (domain.Post, error) {
	var p domain.Post
	err := row.Scan(append([]any{&p.ID, &p.Slug, &p.Title, &p.Summary, &p.Body, &p.CoverObjectKey, &p.Kind, &p.Status, &p.Pinned,
		&p.AuthorID, &p.AuthorName, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt, &p.UnidadeID}, extra...)...)
	return p, err
}

func wrap(err error) error {
	switch {
	case err == nil:
		return nil
	case database.IsNoRows(err):
		return domain.ErrNotFound
	case database.IsUniqueViolation(err):
		return domain.ErrSlugTaken
	}
	return fmt.Errorf("blog: %w", err)
}

func (r *Repository) List(ctx context.Context, db database.DBTX, f domain.Filter, p pagination.Params) ([]domain.Post, int64, error) {
	status := f.Status
	if status == "" {
		status = domain.StatusPublished
	}
	const where = `WHERE ($1 = 'all' OR p.status = $1) AND ($2 = '' OR p.kind = $2)
		AND ($3 = '' OR p.search @@ nexus_search_tsquery('portuguese', $3))
		AND (p.status = 'published' OR NOT $4 OR p.unidade_id = ANY($5::uuid[]))`
	unidades := make([]string, 0, len(f.Unidades))
	for _, u := range f.Unidades {
		unidades = append(unidades, u.String())
	}
	var total int64
	if err := db.QueryRow(ctx, `SELECT count(*)`+from+where, status, f.Kind, f.Query, f.Restrito, unidades).Scan(&total); err != nil {
		return nil, 0, wrap(err)
	}
	rows, err := db.Query(ctx, `SELECT `+cols+from+where+`
		ORDER BY p.pinned DESC, COALESCE(p.published_at, p.updated_at) DESC LIMIT $6 OFFSET $7`,
		status, f.Kind, f.Query, f.Restrito, unidades, p.Limit(), p.Offset())
	if err != nil {
		return nil, 0, wrap(err)
	}
	defer rows.Close()
	out := []domain.Post{}
	for rows.Next() {
		post, err := scan(rows)
		if err != nil {
			return nil, 0, wrap(err)
		}
		post.Body = "" // listagem não carrega o corpo inteiro
		out = append(out, post)
	}
	return out, total, rows.Err()
}

func (r *Repository) Get(ctx context.Context, db database.DBTX, id uuid.UUID) (domain.Post, error) {
	p, err := scan(db.QueryRow(ctx, `SELECT `+cols+from+`WHERE p.id = $1`, id))
	return p, wrap(err)
}

func (r *Repository) GetBySlug(ctx context.Context, db database.DBTX, slug string) (domain.Post, error) {
	p, err := scan(db.QueryRow(ctx, `SELECT `+cols+from+`WHERE p.slug = $1`, slug))
	return p, wrap(err)
}

func (r *Repository) Insert(ctx context.Context, db database.DBTX, p domain.Post) (domain.Post, error) {
	_, err := db.Exec(ctx, `
		INSERT INTO blog_posts (id, slug, title, summary, body, cover_object_key, kind, status, pinned, author_id, published_at, unidade_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		p.ID, p.Slug, p.Title, p.Summary, p.Body, p.CoverObjectKey, p.Kind, p.Status, p.Pinned, p.AuthorID, p.PublishedAt, p.UnidadeID)
	if err != nil {
		return domain.Post{}, wrap(err)
	}
	return r.Get(ctx, db, p.ID)
}

func (r *Repository) Update(ctx context.Context, db database.DBTX, p domain.Post) (domain.Post, error) {
	tag, err := db.Exec(ctx, `
		UPDATE blog_posts SET slug=$2, title=$3, summary=$4, body=$5, cover_object_key=$6, kind=$7,
		       status=$8, pinned=$9, published_at=$10, unidade_id=$11
		WHERE id = $1`,
		p.ID, p.Slug, p.Title, p.Summary, p.Body, p.CoverObjectKey, p.Kind, p.Status, p.Pinned, p.PublishedAt, p.UnidadeID)
	if err != nil {
		return domain.Post{}, wrap(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.Post{}, domain.ErrNotFound
	}
	return r.Get(ctx, db, p.ID)
}

func (r *Repository) Delete(ctx context.Context, db database.DBTX, id uuid.UUID) error {
	tag, err := db.Exec(ctx, `DELETE FROM blog_posts WHERE id = $1`, id)
	if err == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return wrap(err)
}

func (r *Repository) Search(ctx context.Context, db database.DBTX, query string, limit int) ([]domain.Post, []float64, error) {
	rows, err := db.Query(ctx, `
		SELECT `+cols+`, ts_rank(p.search, q) AS rank`+from+`,
		       nexus_search_tsquery('portuguese', $1) q
		WHERE p.status = 'published' AND p.search @@ q
		ORDER BY rank DESC LIMIT $2`, query, limit)
	if err != nil {
		return nil, nil, wrap(err)
	}
	defer rows.Close()
	var posts []domain.Post
	var ranks []float64
	for rows.Next() {
		var rank float32
		p, err := scan(rows, &rank)
		if err != nil {
			return nil, nil, wrap(err)
		}
		posts = append(posts, p)
		ranks = append(ranks, float64(rank))
	}
	return posts, ranks, rows.Err()
}
