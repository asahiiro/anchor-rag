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
}
