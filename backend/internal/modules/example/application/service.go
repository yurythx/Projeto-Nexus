package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yurythx/projeto-nexus/internal/modules/example/domain"
	"github.com/yurythx/projeto-nexus/internal/platform/audit"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
	"github.com/yurythx/projeto-nexus/internal/platform/outbox"
)

// Service gerencia as regras de negócio para o módulo de exemplo.
type Service struct {
	pool   *pgxpool.Pool
	repo   domain.Repository
	outbox *outbox.Writer
	logger *slog.Logger
}

func NewService(pool *pgxpool.Pool, repo domain.Repository, outbox *outbox.Writer, logger *slog.Logger) *Service {
	return &Service{
		pool:   pool,
		repo:   repo,
		outbox: outbox,
		logger: logger,
	}
}

// CreateItem é o blueprint de referência do Transactional Outbox (§16): o
// INSERT do item de negócio e o outbox.Write do evento
// "example.item.created" (já roteado em messaging/topology.go pras filas
// nexus.notification.websocket e nexus.example.worker — só faltava
// alguém de fato publicar) commitam na MESMA transação via
// database.WithTx — uma falha parcial (item gravado mas evento perdido,
// ou vice-versa) nunca acontece. Achado de auditoria: todo módulo
// anterior a este recebia um *outbox.Writer no construtor mas nenhum
// chegava a chamar outbox.Write — a infraestrutura existia, mas nenhum
// caso de uso de verdade a demonstrava.
func (s *Service) CreateItem(ctx context.Context, title, description string) (*domain.Item, error) {
	item, err := domain.NewItem(title, description)
	if err != nil {
		return nil, fmt.Errorf("example.CreateItem: %w", err)
	}

	err = database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.repo.CreateTx(ctx, tx, item); err != nil {
			return err
		}
		payload := map[string]string{
			"id":    item.ID.String(),
			"title": item.Title,
		}
		// uuid.Nil: sem correlation id de negócio próprio aqui — events.New
		// gera um novo quando recebe Nil (ver internal/domain/events).
		if err := s.outbox.Write(ctx, tx, "example.item.created", "example_item", item.ID.String(), uuid.Nil, payload); err != nil {
			return err
		}
		// Trilha de auditoria na MESMA transação do INSERT + outbox:
		// audit.NewWriter(tx) — não o Writer preso ao pool — para que
		// item, evento e auditoria commitem ou revertam juntos. audit.Meta
		// extrai do context o ator, suas roles e o escopo organizacional.
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "example.item.created", "example_item", item.ID.String(), nil, item))
	})
	if err != nil {
		return nil, fmt.Errorf("example.CreateItem: %w", err)
	}

	s.logger.Info("item de exemplo criado com sucesso", slog.String("id", item.ID.String()))
	return item, nil
}

func (s *Service) GetItem(ctx context.Context, id uuid.UUID) (*domain.Item, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListItems(ctx context.Context, page, pageSize int) ([]*domain.Item, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, pageSize, offset)
}
