// Package application contém os casos de uso do Blog. Publicar grava o
// evento blog.post.published no Outbox na MESMA transação da mudança de
// estado (Transactional Outbox), junto com a auditoria.
package application

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/yurythx/projeto-nexus/internal/domain/errors"
	"github.com/yurythx/projeto-nexus/internal/domain/pagination"
	"github.com/yurythx/projeto-nexus/internal/modules/blog/domain"
	"github.com/yurythx/projeto-nexus/internal/platform/audit"
	"github.com/yurythx/projeto-nexus/internal/platform/auth"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
	"github.com/yurythx/projeto-nexus/internal/platform/modkit"
	"github.com/yurythx/projeto-nexus/internal/platform/outbox"
	"github.com/yurythx/projeto-nexus/internal/platform/storage"
)

// EventPostPublished é emitido ao publicar.
const EventPostPublished = "blog.post.published"

const (
	coverPrefix   = "blog/covers"
	coverMaxBytes = 5 * 1024 * 1024
)

// Service implementa os casos de uso.
type Service struct {
	pool   *pgxpool.Pool
	repo   domain.Repository
	outbox *outbox.Writer
	store  storage.Provider
	bucket string
	expiry time.Duration
	logger *slog.Logger
}

// NewService cria o serviço.
func NewService(pool *pgxpool.Pool, repo domain.Repository, ob *outbox.Writer, store storage.Provider, bucket string, expiry time.Duration, logger *slog.Logger) *Service {
	return &Service{pool: pool, repo: repo, outbox: ob, store: store, bucket: bucket, expiry: expiry, logger: logger}
}

// MapError traduz erros de domínio.
func MapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return apperrors.NotFound("publicação não encontrada")
	case errors.Is(err, domain.ErrSlugTaken):
		return apperrors.Conflict("já existe uma publicação com este endereço (slug)")
	case errors.Is(err, domain.ErrInvalidState):
		return apperrors.Conflict("transição de estado não permitida")
	}
	return err
}

func (s *Service) withCover(ctx context.Context, p domain.Post) domain.Post {
	if p.CoverObjectKey != "" && s.store != nil {
		if url, err := s.store.PresignedGetURL(ctx, s.bucket, p.CoverObjectKey, time.Hour); err == nil {
			p.CoverURL = url
		}
	}
	return p
}

// List lista publicações. Rascunhos/arquivados só para quem tem blog:manage.
func (s *Service) List(ctx context.Context, identity auth.Identity, f domain.Filter, p pagination.Params) ([]domain.Post, int64, error) {
	if f.Status != "" && f.Status != domain.StatusPublished && !auth.HasPermission(identity, auth.PermBlogManage) {
		return nil, 0, apperrors.Forbidden("apenas gestores veem rascunhos e arquivados")
	}
	posts, total, err := s.repo.List(ctx, s.pool, f, p)
	if err != nil {
		return nil, 0, err
	}
	for i := range posts {
		posts[i] = s.withCover(ctx, posts[i])
	}
	return posts, total, nil
}

// Get devolve uma publicação por id ou slug (não publicada só p/ gestores).
func (s *Service) Get(ctx context.Context, identity auth.Identity, idOrSlug string) (domain.Post, error) {
	var (
		p   domain.Post
		err error
	)
	if id, perr := uuid.Parse(idOrSlug); perr == nil {
		p, err = s.repo.Get(ctx, s.pool, id)
	} else {
		p, err = s.repo.GetBySlug(ctx, s.pool, idOrSlug)
	}
	if err != nil {
		return domain.Post{}, MapError(err)
	}
	if p.Status != domain.StatusPublished && !auth.HasPermission(identity, auth.PermBlogManage) {
		return domain.Post{}, MapError(domain.ErrNotFound)
	}
	return s.withCover(ctx, p), nil
}

// Input são os campos editáveis.
type Input struct {
	Title          string
	Slug           string
	Summary        string
	Body           string
	Kind           string
	Pinned         bool
	CoverObjectKey string
}

func (s *Service) prepare(ctx context.Context, in Input) (Input, error) {
	in.Title = strings.TrimSpace(in.Title)
	if in.Slug == "" {
		in.Slug = modkit.Slugify(in.Title)
	} else {
		in.Slug = modkit.Slugify(in.Slug)
	}
	if in.Slug == "" {
		return in, apperrors.Validation("não foi possível gerar um slug a partir do título")
	}
	if in.Kind == "" {
		in.Kind = domain.KindNoticia
	}
	if in.CoverObjectKey != "" {
		if _, err := modkit.ConfirmUpload(ctx, s.store, s.bucket, in.CoverObjectKey, coverPrefix, coverMaxBytes, modkit.ImageTypes); err != nil {
			return in, err
		}
	}
	return in, nil
}

// Create cria um rascunho.
func (s *Service) Create(ctx context.Context, in Input) (domain.Post, error) {
	in, err := s.prepare(ctx, in)
	if err != nil {
		return domain.Post{}, err
	}
	author, _, _ := auth.ActorUUID(ctx)
	var out domain.Post
	err = database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		out, err = s.repo.Insert(ctx, tx, domain.Post{
			ID: uuid.New(), Slug: in.Slug, Title: in.Title, Summary: in.Summary, Body: in.Body,
			CoverObjectKey: in.CoverObjectKey, Kind: in.Kind, Status: domain.StatusDraft, Pinned: in.Pinned, AuthorID: author,
		})
		if err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "blog.post.created", "blog_post", out.ID.String(), nil, summary(out)))
	})
	return s.withCover(ctx, out), MapError(err)
}

// Update altera os campos editáveis (qualquer estado).
func (s *Service) Update(ctx context.Context, id uuid.UUID, in Input) (domain.Post, error) {
	in, err := s.prepare(ctx, in)
	if err != nil {
		return domain.Post{}, err
	}
	var out domain.Post
	err = database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		prev, err := s.repo.Get(ctx, tx, id)
		if err != nil {
			return err
		}
		next := prev
		next.Slug, next.Title, next.Summary, next.Body = in.Slug, in.Title, in.Summary, in.Body
		next.Kind, next.Pinned, next.CoverObjectKey = in.Kind, in.Pinned, in.CoverObjectKey
		if out, err = s.repo.Update(ctx, tx, next); err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "blog.post.updated", "blog_post", id.String(), summary(prev), summary(out)))
	})
	return s.withCover(ctx, out), MapError(err)
}

// Transition muda o estado; publicar emite o evento no Outbox.
func (s *Service) Transition(ctx context.Context, id uuid.UUID, to string) (domain.Post, error) {
	var out domain.Post
	err := database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		prev, err := s.repo.Get(ctx, tx, id)
		if err != nil {
			return err
		}
		if !domain.CanTransition(prev.Status, to) {
			return domain.ErrInvalidState
		}
		next := prev
		next.Status = to
		if to == domain.StatusPublished && next.PublishedAt == nil {
			now := time.Now().UTC()
			next.PublishedAt = &now
		}
		if out, err = s.repo.Update(ctx, tx, next); err != nil {
			return err
		}
		if to == domain.StatusPublished {
			if err := s.outbox.Write(ctx, tx, EventPostPublished, "blog_post", id.String(), uuid.Nil, map[string]any{
				"id": id.String(), "slug": out.Slug, "title": out.Title, "kind": out.Kind, "summary": out.Summary,
				"published_at": out.PublishedAt,
			}); err != nil {
				return err
			}
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "blog.post."+to, "blog_post", id.String(),
			map[string]string{"status": prev.Status}, map[string]string{"status": to}))
	})
	return s.withCover(ctx, out), MapError(err)
}

// Delete remove uma publicação.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	var coverKey string
	err := database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		prev, err := s.repo.Get(ctx, tx, id)
		if err != nil {
			return err
		}
		coverKey = prev.CoverObjectKey
		if err := s.repo.Delete(ctx, tx, id); err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "blog.post.deleted", "blog_post", id.String(), summary(prev), nil))
	})
	if err == nil && coverKey != "" {
		if derr := s.store.Delete(ctx, s.bucket, coverKey); derr != nil {
			s.logger.Warn("blog: capa órfã não removida", slog.Any("error", derr))
		}
	}
	return MapError(err)
}

// CoverUpload emite a URL de upload direto da capa.
func (s *Service) CoverUpload(ctx context.Context, filename, contentType string) (modkit.UploadTicket, error) {
	return modkit.NewUpload(ctx, s.store, s.bucket, coverPrefix, filename, contentType, modkit.ImageTypes, s.expiry)
}

// Search alimenta a Busca Global.
func (s *Service) Search(ctx context.Context, query string, limit int) ([]domain.Post, []float64, error) {
	return s.repo.Search(ctx, s.pool, query, limit)
}

// summary é a forma auditada de um post (sem o corpo inteiro).
func summary(p domain.Post) map[string]any {
	return map[string]any{
		"slug": p.Slug, "title": p.Title, "kind": p.Kind, "status": p.Status, "pinned": p.Pinned,
		"cover_object_key": p.CoverObjectKey, "body_len": len(p.Body),
	}
}
