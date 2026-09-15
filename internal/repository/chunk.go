package repository

import (
	"context"

	"github.com/asahiiro/anchor-rag/internal/domain"
)

type ChunkRepository interface {
	SaveBatch(ctx context.Context, chunks []domain.Chunk) error
	FindByDocumentID(
		ctx context.Context,
		documentID string,
	) ([]domain.Chunk, error)
	FindAll(ctx context.Context) ([]domain.Chunk, error)
}

type ChunkSearcher interface {
	SearchSimilar(
		ctx context.Context,
		queryEmbedding []float32,
		limit int,
	) ([]domain.SearchResult, error)
}
