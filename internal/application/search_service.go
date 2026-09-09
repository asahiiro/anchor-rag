package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/embedding"
	"github.com/asahiiro/anchor-rag/internal/repository"
	"github.com/asahiiro/anchor-rag/internal/vector"
)

var (
	ErrEmptySearchQuery = errors.New(
		"search query must not be empty",
	)
	ErrInvalidSearchLimit = errors.New(
		"search limit must be greater than zero",
	)
	ErrUnexpectedQueryEmbeddingCount = errors.New(
		"expected exactly one query embedding",
	)
)

type SearchService struct {
	chunkRepo repository.ChunkRepository
	embedder  embedding.Embedder
}

func NewSearchService(
	chunkRepo repository.ChunkRepository,
	textEmbedder embedding.Embedder,
) *SearchService {
	return &SearchService{
		chunkRepo: chunkRepo,
		embedder:  textEmbedder,
	}
}

func (s *SearchService) Search(
	ctx context.Context,
	query string,
	limit int,
) ([]domain.SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, ErrEmptySearchQuery
	}

	if limit <= 0 {
		return nil, ErrInvalidSearchLimit
	}

	queryEmbeddings, err := s.embedder.Embed(
		ctx,
		[]string{query},
	)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	if len(queryEmbeddings) != 1 {
		return nil, fmt.Errorf(
			"%w: got %d",
			ErrUnexpectedQueryEmbeddingCount,
			len(queryEmbeddings),
		)
	}

	chunks, err := s.chunkRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("find chunks: %w", err)
	}

	results := make([]domain.SearchResult, 0, len(chunks))

	for _, chunk := range chunks {
		score, err := vector.CosineSimilarity(
			queryEmbeddings[0],
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

		results = append(results, domain.SearchResult{
			Chunk: chunk,
			Score: score,
		})
	}

	sort.Slice(results, func(left, right int) bool {
		if results[left].Score == results[right].Score {
			return results[left].Chunk.ID <
				results[right].Chunk.ID
		}

		return results[left].Score > results[right].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}
