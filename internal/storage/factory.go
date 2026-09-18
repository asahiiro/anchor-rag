package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/asahiiro/anchor-rag/internal/database"
	"github.com/asahiiro/anchor-rag/internal/repository"
	"github.com/asahiiro/anchor-rag/internal/repository/memory"
	"github.com/asahiiro/anchor-rag/internal/repository/postgres"
)

var ErrUnsupportedProvider = errors.New(
	"unsupported storage provider",
)

type Config struct {
	Provider    string
	DatabaseURL string
}

type Store struct {
	DocumentRepository repository.DocumentRepository
	ChunkRepository    repository.ChunkRepository
	KnowledgeWriter    repository.KnowledgeWriter
	ChunkSearcher      repository.ChunkSearcher

	close func()
}

func (s *Store) Close() {
	if s.close != nil {
		s.close()
	}
}

func New(
	ctx context.Context,
	config Config,
) (*Store, error) {
	provider := strings.ToLower(
		strings.TrimSpace(config.Provider),
	)

	switch provider {
	case "", "memory":
		documentRepo := memory.NewDocumentRepository()
		chunkRepo := memory.NewChunkRepository()
		knowledgeWriter := memory.NewKnowledgeWriter(
			documentRepo,
			chunkRepo,
		)

		return &Store{
			DocumentRepository: documentRepo,
			ChunkRepository:    chunkRepo,
			KnowledgeWriter:    knowledgeWriter,
			ChunkSearcher:      chunkRepo,
			close:              func() {},
		}, nil

	case "postgres":
		pool, err := database.NewPostgresPool(
			ctx,
			config.DatabaseURL,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"open postgres storage: %w",
				err,
			)
		}

		documentRepo := postgres.NewDocumentRepository(pool)
		chunkRepo := postgres.NewChunkRepository(pool)

		return &Store{
			DocumentRepository: documentRepo,
			ChunkRepository:    chunkRepo,
			KnowledgeWriter:    documentRepo,
			ChunkSearcher:      chunkRepo,
			close:              pool.Close,
		}, nil
	default:
		return nil, fmt.Errorf(
			"%w: %s",
			ErrUnsupportedProvider,
			provider,
		)
	}
}
