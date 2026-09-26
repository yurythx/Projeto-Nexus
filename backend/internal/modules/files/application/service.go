// Package application contém os casos de uso do plugin Arquivos. Toda
// operação avalia a ACL herdada ANTES de tocar dados (A01); o conteúdo
// dos arquivos trafega direto entre navegador e MinIO por URLs
// pré-assinadas de curta duração.
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
	"github.com/yurythx/projeto-nexus/internal/modules/files/domain"
	"github.com/yurythx/projeto-nexus/internal/platform/audit"
	"github.com/yurythx/projeto-nexus/internal/platform/auth"
	"github.com/yurythx/projeto-nexus/internal/platform/database"
	"github.com/yurythx/projeto-nexus/internal/platform/modkit"
	"github.com/yurythx/projeto-nexus/internal/platform/outbox"
	"github.com/yurythx/projeto-nexus/internal/platform/storage"
)

// EventObjectUploaded é emitido ao confirmar um upload.
const EventObjectUploaded = "files.object.uploaded"

const objectPrefix = "files"

// Service implementa os casos de uso.
type Service struct {
	pool     *pgxpool.Pool
	repo     domain.Repository
	outbox   *outbox.Writer
	store    storage.Provider
	bucket   string
	maxBytes int64
	expiry   time.Duration
	logger   *slog.Logger
}

// NewService cria o serviço.
func NewService(pool *pgxpool.Pool, repo domain.Repository, ob *outbox.Writer, store storage.Provider, bucket string, maxBytes int64, expiry time.Duration, logger *slog.Logger) *Service {
	return &Service{pool: pool, repo: repo, outbox: ob, store: store, bucket: bucket, maxBytes: maxBytes, expiry: expiry, logger: logger}
}

// MapError traduz erros de domínio.
func MapError(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return apperrors.NotFound("pasta ou arquivo não encontrado")
	case errors.Is(err, domain.ErrDuplicate):
		return apperrors.Conflict("já existe um item com este nome nesta pasta")
	case errors.Is(err, domain.ErrForbidden):
		return apperrors.Forbidden("você não tem acesso a esta pasta")
	case errors.Is(err, domain.ErrCycle):
		return apperrors.Conflict(err.Error())
	case errors.Is(err, domain.ErrNotEmpty):
		return apperrors.Conflict("a pasta tem subpastas ou arquivos — confirme a exclusão recursiva (recursive=true)").WithCode("FOLDER_NOT_EMPTY")
	}
	return err
}

func cleanName(n string) (string, error) {
	n = strings.TrimSpace(n)
	if n == "" || len(n) > 200 || strings.ContainsAny(n, `/\`) || n == "." || n == ".." {
		return "", apperrors.Validation("nome inválido (sem barras, até 200 caracteres)")
	}
	return n, nil
}

func (s *Service) access(ctx context.Context, db database.DBTX, identity auth.Identity, folderID uuid.UUID) (domain.Access, error) {
	chain, err := s.repo.Chain(ctx, db, folderID)
	if err != nil {
		return domain.Access{}, err
	}
	return domain.Evaluate(identity, chain), nil
}

// Listing é o conteúdo de uma pasta.
type Listing struct {
	Folder      *domain.Folder  `json:"folder,omitempty"`
	Breadcrumbs []domain.Folder `json:"breadcrumbs"`
	Folders     []domain.Folder `json:"folders"`
	Files       []domain.File   `json:"files"`
	Access      domain.Access   `json:"access"`
}

// Browse lista a raiz (parent nil) ou o conteúdo de uma pasta.
func (s *Service) Browse(ctx context.Context, identity auth.Identity, parent *uuid.UUID) (Listing, error) {
	if parent == nil {
		roots, acls, err := s.repo.RootFolders(ctx, s.pool)
		if err != nil {
			return Listing{}, err
		}
		visible := []domain.Folder{}
		for _, f := range roots {
			if domain.Evaluate(identity, []domain.ChainLink{{FolderID: f.ID, OwnerID: f.OwnerID, ACL: acls[f.ID]}}).Read {
				visible = append(visible, f)
			}
		}
		// Na raiz qualquer autenticado pode criar a própria pasta.
		return Listing{Breadcrumbs: []domain.Folder{}, Folders: visible, Files: []domain.File{},
			Access: domain.Access{Read: true, Write: true}}, nil
	}
	acc, err := s.access(ctx, s.pool, identity, *parent)
	if err != nil {
		return Listing{}, MapError(err)
	}
	if !acc.Read {
		return Listing{}, MapError(domain.ErrForbidden)
	}
	folder, err := s.repo.GetFolder(ctx, s.pool, *parent)
	if err != nil {
		return Listing{}, MapError(err)
	}
	crumbs, err := s.repo.Breadcrumbs(ctx, s.pool, *parent)
	if err != nil {
		return Listing{}, err
	}
	sub, err := s.repo.Children(ctx, s.pool, *parent)
	if err != nil {
		return Listing{}, err
	}
	files, err := s.repo.Files(ctx, s.pool, *parent)
	if err != nil {
		return Listing{}, err
	}
	return Listing{Folder: &folder, Breadcrumbs: crumbs, Folders: sub, Files: files, Access: acc}, nil
}

// CreateFolder cria uma pasta (raiz: qualquer autenticado; subpasta: escrita).
func (s *Service) CreateFolder(ctx context.Context, identity auth.Identity, parent *uuid.UUID, name string) (domain.Folder, error) {
	name, err := cleanName(name)
	if err != nil {
		return domain.Folder{}, err
	}
	var out domain.Folder
	err = database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		if parent != nil {
			acc, err := s.access(ctx, tx, identity, *parent)
			if err != nil {
				return err
			}
			if !acc.Write {
				return domain.ErrForbidden
			}
		}
		var err error
		out, err = s.repo.SaveFolder(ctx, tx, domain.Folder{ID: uuid.New(), ParentID: parent, Name: name, OwnerID: identity.UserID})
		if err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "files.folder.created", "files_folder", out.ID.String(), nil, out))
	})
	return out, MapError(err)
}

// UpdateFolder renomeia e/ou move uma pasta (manage na pasta; escrita no destino).
func (s *Service) UpdateFolder(ctx context.Context, identity auth.Identity, id uuid.UUID, name string, newParent *uuid.UUID, move bool) (domain.Folder, error) {
	name, err := cleanName(name)
	if err != nil {
		return domain.Folder{}, err
	}
	var out domain.Folder
	err = database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		acc, err := s.access(ctx, tx, identity, id)
		if err != nil {
			return err
		}
		if !acc.Manage {
			return domain.ErrForbidden
		}
		prev, err := s.repo.GetFolder(ctx, tx, id)
		if err != nil {
			return err
		}
		next := prev
		next.Name = name
		if move {
			if newParent != nil {
				if *newParent == id {
					return domain.ErrCycle
				}
				if desc, err := s.repo.IsDescendant(ctx, tx, *newParent, id); err != nil || desc {
					return domain.ErrCycle
				}
				dst, err := s.access(ctx, tx, identity, *newParent)
				if err != nil {
					return err
				}
				if !dst.Write {
					return domain.ErrForbidden
				}
			}
			next.ParentID = newParent
		}
		if out, err = s.repo.SaveFolder(ctx, tx, next); err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "files.folder.updated", "files_folder", id.String(), prev, out))
	})
	return out, MapError(err)
}

// DeleteFolder apaga a pasta. Com conteúdo, só com recursive=true
// (confirmação explícita): aí leva subpastas, arquivos e os objetos no
// MinIO — operação destrutiva que nunca acontece por engano.
func (s *Service) DeleteFolder(ctx context.Context, identity auth.Identity, id uuid.UUID, recursive bool) error {
	var keys []string
	err := database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		acc, err := s.access(ctx, tx, identity, id)
		if err != nil {
			return err
		}
		if !acc.Manage {
			return domain.ErrForbidden
		}
		prev, err := s.repo.GetFolder(ctx, tx, id)
		if err != nil {
			return err
		}
		if !recursive {
			empty, err := s.repo.FolderEmpty(ctx, tx, id)
			if err != nil {
				return err
			}
			if !empty {
				return domain.ErrNotEmpty
			}
		}
		if keys, err = s.repo.DeleteFolder(ctx, tx, id); err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "files.folder.deleted", "files_folder", id.String(), prev,
			map[string]int{"objects_removed": len(keys)}))
	})
	if err == nil {
		s.removeObjects(ctx, keys)
	}
	return MapError(err)
}

func (s *Service) removeObjects(ctx context.Context, keys []string) {
	for _, k := range keys {
		if err := s.store.Delete(ctx, s.bucket, k); err != nil {
			s.logger.Warn("files: objeto órfão não removido do storage", slog.String("key", k), slog.Any("error", err))
		}
	}
}

// ACL devolve a ACL própria da pasta (manage).
func (s *Service) ACL(ctx context.Context, identity auth.Identity, id uuid.UUID) ([]domain.ACLEntry, error) {
	acc, err := s.access(ctx, s.pool, identity, id)
	if err != nil {
		return nil, MapError(err)
	}
	if !acc.Manage {
		return nil, MapError(domain.ErrForbidden)
	}
	return s.repo.ACL(ctx, s.pool, id)
}

// SetACL substitui a ACL própria da pasta (manage).
func (s *Service) SetACL(ctx context.Context, identity auth.Identity, id uuid.UUID, entries []domain.ACLEntry) ([]domain.ACLEntry, error) {
	for i := range entries {
		entries[i].Subject = strings.TrimSpace(entries[i].Subject)
		if entries[i].SubjectType == "everyone" {
			entries[i].Subject = "*"
		} else if entries[i].Subject == "" {
			return nil, apperrors.Validation("toda entrada de ACL (exceto 'everyone') precisa de um sujeito")
		}
		if entries[i].SubjectType == "user" || entries[i].SubjectType == "unidade" || entries[i].SubjectType == "departamento" {
			if _, err := uuid.Parse(entries[i].Subject); err != nil {
				return nil, apperrors.Validation("sujeito de ACL do tipo " + entries[i].SubjectType + " precisa ser um id (UUID)")
			}
		}
	}
	err := database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		acc, err := s.access(ctx, tx, identity, id)
		if err != nil {
			return err
		}
		if !acc.Manage {
			return domain.ErrForbidden
		}
		prev, err := s.repo.ACL(ctx, tx, id)
		if err != nil {
			return err
		}
		if err := s.repo.ReplaceACL(ctx, tx, id, entries, identity.Username); err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "files.folder.acl_changed", "files_folder", id.String(), prev, entries))
	})
	if err != nil {
		return nil, MapError(err)
	}
	return entries, nil
}

// UploadResult é o ticket de upload + o registro pendente.
type UploadResult struct {
	File   domain.File         `json:"file"`
	Upload modkit.UploadTicket `json:"upload"`
}

// StartUpload cria o registro pendente e a URL de upload direto.
func (s *Service) StartUpload(ctx context.Context, identity auth.Identity, folderID uuid.UUID, filename, contentType string, size int64) (UploadResult, error) {
	name, err := cleanName(filename)
	if err != nil {
		return UploadResult{}, err
	}
	if size <= 0 || size > s.maxBytes {
		return UploadResult{}, apperrors.Validation("tamanho de arquivo acima do limite permitido")
	}
	acc, err := s.access(ctx, s.pool, identity, folderID)
	if err != nil {
		return UploadResult{}, MapError(err)
	}
	if !acc.Write {
		return UploadResult{}, MapError(domain.ErrForbidden)
	}
	ticket, err := modkit.NewUpload(ctx, s.store, s.bucket, objectPrefix, name, contentType, modkit.DocumentTypes, s.expiry)
	if err != nil {
		return UploadResult{}, err
	}
	f, err := s.repo.InsertFile(ctx, s.pool, domain.File{
		ID: uuid.New(), FolderID: folderID, Name: name, ObjectKey: ticket.ObjectKey, SizeBytes: size,
		ContentType: ticket.Headers["Content-Type"], Status: "pending", OwnerID: identity.UserID,
	})
	if err != nil {
		return UploadResult{}, MapError(err)
	}
	return UploadResult{File: f, Upload: ticket}, nil
}

// ConfirmUpload valida o objeto enviado e publica o arquivo.
func (s *Service) ConfirmUpload(ctx context.Context, identity auth.Identity, id uuid.UUID) (domain.File, error) {
	f, err := s.repo.GetFile(ctx, s.pool, id)
	if err != nil {
		return domain.File{}, MapError(err)
	}
	if f.OwnerID != identity.UserID {
		return domain.File{}, MapError(domain.ErrForbidden)
	}
	if f.Status == "ready" {
		return f, nil
	}
	info, err := modkit.ConfirmUpload(ctx, s.store, s.bucket, f.ObjectKey, objectPrefix, s.maxBytes, modkit.DocumentTypes)
	if err != nil {
		return domain.File{}, err
	}
	var out domain.File
	err = database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		f.SizeBytes, f.Status = info.Size, "ready"
		var err error
		if out, err = s.repo.UpdateFile(ctx, tx, f); err != nil {
			return err
		}
		if err := s.outbox.Write(ctx, tx, EventObjectUploaded, "files_object", id.String(), uuid.Nil, map[string]any{
			"id": id.String(), "folder_id": f.FolderID.String(), "name": f.Name, "size_bytes": f.SizeBytes, "content_type": f.ContentType,
		}); err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "files.object.uploaded", "files_object", id.String(), nil, out))
	})
	return out, MapError(err)
}

// DownloadURL devolve uma URL pré-assinada de leitura (curta).
func (s *Service) DownloadURL(ctx context.Context, identity auth.Identity, id uuid.UUID) (string, domain.File, error) {
	f, err := s.repo.GetFile(ctx, s.pool, id)
	if err != nil || f.Status != "ready" {
		return "", domain.File{}, MapError(domain.ErrNotFound)
	}
	acc, err := s.access(ctx, s.pool, identity, f.FolderID)
	if err != nil {
		return "", domain.File{}, MapError(err)
	}
	if !acc.Read {
		return "", domain.File{}, MapError(domain.ErrForbidden)
	}
	url, err := s.store.PresignedGetURL(ctx, s.bucket, f.ObjectKey, 5*time.Minute)
	if err != nil {
		return "", domain.File{}, apperrors.DependencyUnavailable("armazenamento indisponível").WithCause(err)
	}
	_ = audit.NewWriter(s.pool).Record(ctx, audit.Meta(ctx, "files.object.downloaded", "files_object", id.String(), nil, nil))
	return url, f, nil
}

// UpdateFile renomeia/move um arquivo (escrita na origem e no destino).
func (s *Service) UpdateFile(ctx context.Context, identity auth.Identity, id uuid.UUID, name string, folderID uuid.UUID) (domain.File, error) {
	name, err := cleanName(name)
	if err != nil {
		return domain.File{}, err
	}
	var out domain.File
	err = database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		prev, err := s.repo.GetFile(ctx, tx, id)
		if err != nil {
			return err
		}
		for _, folder := range []uuid.UUID{prev.FolderID, folderID} {
			acc, err := s.access(ctx, tx, identity, folder)
			if err != nil {
				return err
			}
			if !acc.Write && prev.OwnerID != identity.UserID {
				return domain.ErrForbidden
			}
		}
		next := prev
		next.Name, next.FolderID = name, folderID
		if out, err = s.repo.UpdateFile(ctx, tx, next); err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "files.object.updated", "files_object", id.String(), prev, out))
	})
	return out, MapError(err)
}

// DeleteFile apaga o arquivo (dono, escrita+files:manage na pasta).
func (s *Service) DeleteFile(ctx context.Context, identity auth.Identity, id uuid.UUID) error {
	var key string
	err := database.WithTx(ctx, s.pool, func(ctx context.Context, tx pgx.Tx) error {
		prev, err := s.repo.GetFile(ctx, tx, id)
		if err != nil {
			return err
		}
		acc, err := s.access(ctx, tx, identity, prev.FolderID)
		if err != nil {
			return err
		}
		if prev.OwnerID != identity.UserID && !acc.Manage {
			return domain.ErrForbidden
		}
		key = prev.ObjectKey
		if err := s.repo.DeleteFile(ctx, tx, id); err != nil {
			return err
		}
		return audit.NewWriter(tx).Record(ctx, audit.Meta(ctx, "files.object.deleted", "files_object", id.String(), prev, nil))
	})
	if err == nil {
		s.removeObjects(ctx, []string{key})
	}
	return MapError(err)
}

// Search busca arquivos por nome, filtrando pela ACL do usuário.
func (s *Service) Search(ctx context.Context, identity auth.Identity, q string, limit int) ([]domain.File, error) {
	candidates, err := s.repo.SearchFiles(ctx, s.pool, q, limit*4)
	if err != nil {
		return nil, err
	}
	out := []domain.File{}
	cache := map[uuid.UUID]bool{}
	for _, f := range candidates {
		ok, seen := cache[f.FolderID]
		if !seen {
			acc, err := s.access(ctx, s.pool, identity, f.FolderID)
			ok = err == nil && acc.Read
			cache[f.FolderID] = ok
		}
		if ok {
			out = append(out, f)
			if len(out) == limit {
				break
			}
		}
	}
	return out, nil
}

// CollectStaleUploads remove uploads pendentes abandonados (>24h) — worker.
func (s *Service) CollectStaleUploads(ctx context.Context) error {
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		stale, err := s.repo.StalePending(ctx, s.pool, time.Now().Add(-24*time.Hour), 200)
		if err != nil && ctx.Err() == nil {
			s.logger.Warn("files: varredura de uploads pendentes falhou", slog.Any("error", err))
		}
		for _, f := range stale {
			_ = s.store.Delete(ctx, s.bucket, f.ObjectKey)
			_ = s.repo.DeleteFile(ctx, s.pool, f.ID)
		}
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
}
