package application

import (
	"context"
	"strings"
	"testing"

	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/generation"
	"github.com/asahiiro/anchor-rag/internal/repository/memory"
)

type capturingGenerator struct {
	messages []generation.Message
	answer   string
}

func (g *capturingGenerator) Generate(
	ctx context.Context,
	messages []generation.Message,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	g.messages = append([]generation.Message(nil), messages...)
	return g.answer, nil
}

func TestAnswerService(t *testing.T) {
	chunkRepo := memory.NewChunkRepository()

	chunks := []domain.Chunk{
		{
			ID:         "chunk-go",
			DocumentID: "doc-go",
			Content:    "Go uses goroutines and channels for concurrency.",
			Embedding:  []float32{1, 0},
		},
		{
			ID:         "chunk-cooking",
			DocumentID: "doc-cooking",
			Content:    "A cake contains flour and eggs.",
			Embedding:  []float32{0, 1},
		},
	}

	if err := chunkRepo.SaveBatch(
		context.Background(),
		chunks,
	); err != nil {
		t.Fatalf("SaveBatch() returned error: %v", err)
	}

	searchService := NewSearchService(
		chunkRepo,
		staticEmbedder{value: []float32{1, 0}},
	)

	generator := &capturingGenerator{
		answer: "Go uses goroutines and channels [1].",
	}

	service := NewAnswerService(searchService, generator)

	got, err := service.Answer(
		context.Background(),
		"How does Go implement concurrency?",
		1,
	)
	if err != nil {
		t.Fatalf("Answer() returned error: %v", err)
	}

	if got.Answer != "Go uses goroutines and channels [1]." {
		t.Fatalf("answer = %q", got.Answer)
	}

	if len(got.Sources) != 1 {
		t.Fatalf("got %d sources, want 1", len(got.Sources))
	}

	if got.Sources[0].Chunk.ID != "chunk-go" {
		t.Fatalf(
			"source = %q, want chunk-go",
			got.Sources[0].Chunk.ID,
		)
	}

	if len(generator.messages) != 2 {
		t.Fatalf(
			"got %d messages, want 2",
			len(generator.messages),
		)
	}

	userPrompt := generator.messages[1].Content

	if !strings.Contains(
		userPrompt,
		"How does Go implement concurrency?",
	) {
		t.Fatal("prompt does not contain the question")
	}

	if !strings.Contains(
		userPrompt,
		"Go uses goroutines and channels for concurrency.",
	) {
		t.Fatal("prompt does not contain retrieved context")
	}

	if !strings.Contains(userPrompt,
		"[1]") {
		t.Fatal("prompt does not contain citation marker")
	}
}

func TestAnswerServiceReturnsNoKnowledge(t *testing.T) {
	searchService := NewSearchService(
		memory.NewChunkRepository(),
		staticEmbedder{value: []float32{1, 0}},
	)

	generator := &capturingGenerator{
		answer: "This should not be used.",
	}

	service := NewAnswerService(searchService, generator)

	_, err := service.Answer(
		context.Background(),
		"What is RAG?",
		3,
	)

	if err != ErrNoRelevantKnowledge {
		t.Fatalf("expected ErrNoRelevantKnowledge, got %v", err)
	}

	if len(generator.messages) != 0 {
		t.Fatal("generator should not run without retrieved knowledge")
	}
}
