package memory

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/repository"
	"github.com/asahiiro/anchor-rag/internal/vector"
)

type ChunkRepository struct {
	mu               sync.RWMutex
	chunksByDocument map[string][]domain.Chunk
}

var _ repository.ChunkRepository = (*ChunkRepository)(nil)
var _ repository.ChunkSearcher = (*ChunkRepository)(nil)

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

func (r *ChunkRepository) SearchSimilar(
	ctx context.Context,
	queryEmbedding []float32,
	limit int,
) ([]domain.SearchResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if limit <= 0 {
		return nil, errors.New(
			"search limit must be greater than zero",
		)
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(
		[]domain.SearchResult,
		0,
	)

	for _, chunks := range r.chunksByDocument {
		for _, chunk := range chunks {
			score, err := vector.CosineSimilarity(
				queryEmbedding,
				chunk.Embedding,
			)
			if errors.Is(err, vector.ErrZeroVector) {
				continue
			}
			if err != nil {
				return nil, fmt.Errorf(
					"compare chunk %s: %w",
					chunk.ID,
					err,
				)
			}

			resultChunk := chunk
			resultChunk.Embedding = append(
				[]float32(nil),
				chunk.Embedding...,
			)

			results = append(
				results,
				domain.SearchResult{
					Chunk: resultChunk,
					Score: score,
				},
			)
		}
	}

	sort.Slice(
		results,
		func(left, right int) bool {
			if results[left].Score ==
				results[right].Score {
				return results[left].Chunk.ID <
					results[right].Chunk.ID
			}

			return results[left].Score >
				results[right].Score
		},
	)

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}
