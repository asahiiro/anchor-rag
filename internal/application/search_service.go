package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/embedding"
	"github.com/asahiiro/anchor-rag/internal/repository"
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
	chunkSearcher repository.ChunkSearcher
	embedder      embedding.Embedder
}

func NewSearchService(
	chunkSearcher repository.ChunkSearcher,
	textEmbedder embedding.Embedder,
) *SearchService {
	return &SearchService{
		chunkSearcher: chunkSearcher,
		embedder:      textEmbedder,
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

	results, err := s.chunkSearcher.SearchSimilar(
		ctx,
		queryEmbeddings[0],
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"search similar chunks: %w",
			err,
		)
	}

	return results, nil
}
