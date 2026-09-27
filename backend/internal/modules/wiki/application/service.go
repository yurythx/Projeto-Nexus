// Package application contém os casos de uso da Wiki.
package application

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/yurythx/projeto-nexus/internal/domain/errors"
	"github.com/yurythx/projeto-nexus/internal/modules/wiki/domain"
	"github.com/yurythx/projeto-nexus/internal/platform/audit"
	"github.com/yurythx/projeto-nexus/internal/platform/auth"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
	"github.com/yurythx/projeto-nexus/internal/platform/modkit"
)

// Service implementa os casos de uso.
type Service struct {
	pool *pgxpool.Pool
	repo domain.Repository
}

// NewService cria o serviço.
func NewService(pool *pgxpool.Pool, repo domain.Repository) *Service {
	return &Service{pool: pool, repo: repo}
}

// MapError traduz erros de domínio.
func MapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return apperrors.NotFound("página não encontrada")
	case errors.Is(err, domain.ErrSlugTaken):
		return apperrors.Conflict("já existe uma página com este endereço (slug)")
	case errors.Is(err, domain.ErrStale):
		return apperrors.Conflict("a página foi alterada por outra pessoa desde que você a abriu — recarregue e reaplique suas mudanças").WithCode("WIKI_STALE_VERSION")
	case errors.Is(err, domain.ErrHasChildren):
		return apperrors.Conflict("mova ou exclua as subpáginas antes de excluir esta página")
	case errors.Is(err, domain.ErrCycle):
		return apperrors.Conflict(err.Error())
	case errors.Is(err, domain.ErrOutOfScope):
		return apperrors.Forbidden(err.Error())
	}
	return err
}

// Tree lista todas as páginas (sem corpo).
func (s *Service) Tree(ctx context.Context) ([]domain.Page, error) { return s.repo.Tree(ctx, s.pool) }

// PageView é uma página com a trilha até a raiz.
type PageView struct {
	domain.Page
	Breadcrumbs []domain.Page `json:"breadcrumbs"`
}

// Get devolve uma página por id ou slug.
func (s *Service) Get(ctx context.Context, ref string) (PageView, error) {
	var (
		p   domain.Page
		err error
	)
	if id, perr := uuid.Parse(ref); perr == nil {
		p, err = s.repo.Get(ctx, s.pool, id)
	} else {
		p, err = s.repo.GetBySlug(ctx, s.pool, ref)
	}
	if err != nil {
		return PageView{}, MapError(err)
	}
	crumbs, err := s.repo.Breadcrumbs(ctx, s.pool, p.ID)
	if err != nil {
		return PageView{}, err
	}
	return PageView{Page: p, Breadcrumbs: crumbs}, nil
}

// Input são os campos editáveis.
type Input struct {
	ParentID *uuid.UUID
	Title    string
	Slug     string
	Body     string
	Position int
	Summary  string
	// UnidadeID é a unidade dona; na criação, vazio herda a da página-mãe.
	UnidadeID *uuid.UUID
}

// podeMarcar: só marca uma unidade como dona quem está lotado nela ou tem
// wiki:manage cobrindo-a (ninguém atribui página a unidade alheia).
func podeMarcar(identity auth.Identity, unidade *uuid.UUID) bool {
	return unidade == nil || auth.LotadoEm(identity, *unidade) || auth.Can(identity, auth.PermWikiManage, auth.InUnidade(*unidade))
}

// dona é a posição da página para wiki:manage (nil = institucional).
func dona(unidade *uuid.UUID) auth.Target {
	if unidade == nil {
		return auth.Target{}
	}
	return auth.InUnidade(*unidade)
}

// Create cria uma página (qualquer autenticado).
func (s *Service) Create(ctx context.Context, identity auth.Identity, in Input) (domain.Page, error) {
	in.Title = strings.TrimSpace(in.Title)
	slug := modkit.Slugify(in.Slug)
	if slug == "" {
		slug = modkit.Slugify(in.Title)
	}
	if slug == "" {
		return domain.Page{}, apperrors.Validation("o título (ou o slug) precisa ter letras ou números para formar o endereço")
	}
	var out domain.Page
	err := database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		if in.ParentID != nil {
			parent, err := s.repo.Get(ctx, tx, *in.ParentID)
			if err != nil {
				return err
			}
			if in.UnidadeID == nil {
				in.UnidadeID = parent.UnidadeID // herda a dona da página-mãe
			} else if !podeMarcar(identity, in.UnidadeID) {
				return domain.ErrOutOfScope
			}
		} else if !podeMarcar(identity, in.UnidadeID) {
			return domain.ErrOutOfScope
		}
		p := domain.Page{ID: uuid.New(), ParentID: in.ParentID, Slug: slug, Title: in.Title, Body: in.Body,
			Position: in.Position, CreatedBy: identity.UserID, UpdatedBy: identity.UserID, UnidadeID: in.UnidadeID}
		if err := s.repo.Insert(ctx, tx, p); err != nil {
			return err
		}
		if err := s.repo.AddRevision(ctx, tx, domain.Revision{PageID: p.ID, Version: 1, Title: p.Title, Body: p.Body,
			Summary: firstNonEmpty(in.Summary, "Criação da página"), EditedBy: identity.UserID}); err != nil {
			return err
		}
		var err error
		if out, err = s.repo.Get(ctx, tx, p.ID); err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "wiki.page.created", "wiki_page", p.ID.String(), nil,
			map[string]any{"slug": out.Slug, "title": out.Title}))
	})
	return out, MapError(err)
}

// Update edita a página com concorrência otimista (expectedVersion).
func (s *Service) Update(ctx context.Context, identity auth.Identity, id uuid.UUID, expectedVersion int, in Input) (domain.Page, error) {
	in.Title = strings.TrimSpace(in.Title)
	var out domain.Page
	err := database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		prev, err := s.repo.Get(ctx, tx, id)
		if err != nil {
			return err
		}
		if in.ParentID != nil {
			if *in.ParentID == id {
				return domain.ErrCycle
			}
			if _, err := s.repo.Get(ctx, tx, *in.ParentID); err != nil {
				return err
			}
			desc, err := s.repo.IsDescendant(ctx, tx, *in.ParentID, id)
			if err != nil {
				return err
			}
			if desc {
				return domain.ErrCycle
			}
		}
		// Trocar a dona: quem troca precisa poder marcar a antiga e a nova.
		if !sameUnidade(prev.UnidadeID, in.UnidadeID) && (!podeMarcar(identity, prev.UnidadeID) || !podeMarcar(identity, in.UnidadeID)) {
			return domain.ErrOutOfScope
		}
		slug := modkit.Slugify(in.Slug)
		if slug == "" {
			slug = prev.Slug
		}
		next := prev
		next.UnidadeID = in.UnidadeID
		next.ParentID, next.Slug, next.Title, next.Body, next.Position, next.UpdatedBy = in.ParentID, slug, in.Title, in.Body, in.Position, identity.UserID
		if err := s.repo.UpdateIfVersion(ctx, tx, next, expectedVersion); err != nil {
			return err
		}
		if out, err = s.repo.Get(ctx, tx, id); err != nil {
			return err
		}
		if err := s.repo.AddRevision(ctx, tx, domain.Revision{PageID: id, Version: out.Version, Title: out.Title, Body: out.Body,
			Summary: firstNonEmpty(in.Summary, "Edição"), EditedBy: identity.UserID}); err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "wiki.page.updated", "wiki_page", id.String(),
			map[string]any{"version": prev.Version, "title": prev.Title, "slug": prev.Slug},
			map[string]any{"version": out.Version, "title": out.Title, "slug": out.Slug}))
	})
	return out, MapError(err)
}

// Restore recria o conteúdo de uma revisão como nova versão.
func (s *Service) Restore(ctx context.Context, identity auth.Identity, id uuid.UUID, version int) (domain.Page, error) {
	rev, err := s.repo.Revision(ctx, s.pool, id, version)
	if err != nil {
		return domain.Page{}, MapError(err)
	}
	cur, err := s.repo.Get(ctx, s.pool, id)
	if err != nil {
		return domain.Page{}, MapError(err)
	}
	return s.Update(ctx, identity, id, cur.Version, Input{
		ParentID: cur.ParentID, Title: rev.Title, Slug: cur.Slug, Body: rev.Body, Position: cur.Position, UnidadeID: cur.UnidadeID,
		Summary: "Restauração da versão " + strconv.Itoa(version),
	})
}

// Delete exclui a página: wiki:manage cobrindo a unidade dona (ADR 013).
func (s *Service) Delete(ctx context.Context, identity auth.Identity, id uuid.UUID) error {
	return MapError(database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		prev, err := s.repo.Get(ctx, tx, id)
		if err != nil {
			return err
		}
		if !auth.Can(identity, auth.PermWikiManage, dona(prev.UnidadeID)) {
			return domain.ErrOutOfScope
		}
		if err := s.repo.Delete(ctx, tx, id); err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "wiki.page.deleted", "wiki_page", id.String(),
			map[string]any{"slug": prev.Slug, "title": prev.Title, "version": prev.Version}, nil))
	}))
}

func sameUnidade(a, b *uuid.UUID) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// Revisions lista o histórico (página inexistente: 404).
func (s *Service) Revisions(ctx context.Context, id uuid.UUID) ([]domain.Revision, error) {
	if _, err := s.repo.Get(ctx, s.pool, id); err != nil {
		return nil, MapError(err)
	}
	return s.repo.Revisions(ctx, s.pool, id)
}

// Revision devolve uma revisão.
func (s *Service) Revision(ctx context.Context, id uuid.UUID, version int) (domain.Revision, error) {
	rev, err := s.repo.Revision(ctx, s.pool, id, version)
	return rev, MapError(err)
}

// Search alimenta a Busca Global.
func (s *Service) Search(ctx context.Context, q string, limit int) ([]domain.Page, []float64, error) {
	return s.repo.Search(ctx, s.pool, q, limit)
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
