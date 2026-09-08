package memory

import (
	"context"
	"sync"

	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/repository"
)

type ChunkRepository struct {
	mu               sync.RWMutex
	chunksByDocument map[string][]domain.Chunk
}

var _ repository.ChunkRepository = (*ChunkRepository)(nil)

func NewChunkRepository() *ChunkRepository {
	return &ChunkRepository{
		chunksByDocument: make(map[string][]domain.Chunk),
	}
}

func (r *ChunkRepository) SaveBatch(
	ctx context.Context,
	chunks []domain.Chunk,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, chunk := range chunks {
		documentID := chunk.DocumentID

		r.chunksByDocument[documentID] = append(
			r.chunksByDocument[documentID],
			chunk,
		)
	}

	return nil
}

func (r *ChunkRepository) FindByDocumentID(
	ctx context.Context,
	documentID string,
) ([]domain.Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	stored := r.chunksByDocument[documentID]

	result := make([]domain.Chunk, len(stored))
	copy(result, stored)

	return result, nil
}

func (r *ChunkRepository) FindAll(
	ctx context.Context,
) ([]domain.Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	total := 0
	for _, chunks := range r.chunksByDocument {
		total += len(chunks)
	}

	result := make([]domain.Chunk, 0, total)
	for _, chunks := range r.chunksByDocument {
		result = append(result, chunks...)
	}

	return result, nil
}
