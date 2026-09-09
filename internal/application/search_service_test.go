package application

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/repository/memory"
)

type staticEmbedder struct {
	value []float32
}

func (s staticEmbedder) Embed(
	ctx context.Context,
	texts []string,
) ([][]float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := make([][]float32, len(texts))
	for index := range texts {
		result[index] = append([]float32(nil), s.value...)
	}

	return result, nil
}

func TestSearchServiceReturnsTopK(t *testing.T) {
	repo := memory.NewChunkRepository()

	chunks := []domain.Chunk{
		{
			ID:         "chunk-orthogonal",
			DocumentID: "doc-1",
			Content:    "orthogonal",
			Embedding:  []float32{0, 1},
		},
		{
			ID:         "chunk-exact",
			DocumentID: "doc-2",
			Content:    "exact match",
			Embedding:  []float32{1, 0},
		},
		{
			ID:         "chunk-similar",
			DocumentID: "doc-3",
			Content:    "similar match",
			Embedding:  []float32{0.8, 0.2},
		},
	}

	if err := repo.SaveBatch(context.Background(), chunks); err != nil {
		t.Fatalf("SaveBatch() return error: %v", err)
	}

	service := NewSearchService(
		repo,
		staticEmbedder{value: []float32{1, 0}},
	)

	got, err := service.Search(
		context.Background(),
		"retrieval query",
		2,
	)
	if err != nil {
		t.Fatalf("Search() returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d results, want 2", len(got))
	}

	if got[0].Chunk.ID != "chunk-exact" {
		t.Fatalf(
			"first result = %q, want chunk-exact",
			got[0].Chunk.ID,
		)
	}
	if got[1].Chunk.ID != "chunk-similar" {
		t.Fatalf(
			"second result = %q, want chunk-similar",
			got[1].Chunk.ID,
		)
	}

	if math.Abs(got[0].Score-1) > 1e-9 {
		t.Fatalf("first score = %f, want 1", got[0].Score)
	}
}

func TestSearchServiceRejectsEmptyQuery(t *testing.T) {
	service := NewSearchService(
		memory.NewChunkRepository(),
		staticEmbedder{value: []float32{1, 0}},
	)

	_, err := service.Search(context.Background(), "     ", 3)
	if !errors.Is(err, ErrEmptySearchQuery) {
		t.Fatalf("expected ErrEmptySearchQuery, got %v", err)
	}
}

func TestSearchServiceRejectsInvalidLimit(t *testing.T) {
	service := NewSearchService(
		memory.NewChunkRepository(),
		staticEmbedder{value: []float32{1, 0}},
	)

	_, err := service.Search(context.Background(), "RAG", 0)
	if !errors.Is(err, ErrInvalidSearchLimit) {
		t.Fatalf("expected ErrInvalidSearchLimit, got %v", err)
	}
}
