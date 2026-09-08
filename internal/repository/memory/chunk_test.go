package memory

import (
	"context"
	"reflect"
	"testing"

	"github.com/asahiiro/anchor-rag/internal/domain"
)

func TestChunkRepositorySaveAndFind(t *testing.T) {
	repo := NewChunkRepository()

	chunks := []domain.Chunk{
		{
			ID:         "chunk-1",
			DocumentID: "doc-1",
			Content:    "abcde",
			Position:   0,
		},
		{
			ID:         "chunk-2",
			DocumentID: "doc-1",
			Content:    "defgh",
			Position:   1,
		},
	}

	if err := repo.SaveBatch(context.Background(), chunks); err != nil {
		t.Fatalf("SaveBatch() returned error: %v", err)
	}

	got, err := repo.FindByDocumentID(
		context.Background(),
		"doc-1",
	)
	if err != nil {
		t.Fatalf("FindByDocumentID() returned error: %v", err)
	}

	if !reflect.DeepEqual(got, chunks) {
		t.Fatalf("got %#v, want %#v", got, chunks)
	}
}

func TestChunkRepositorySeparatesDocuments(t *testing.T) {
	repo := NewChunkRepository()

	chunks := []domain.Chunk{
		{
			ID:         "chunk-1",
			DocumentID: "doc-1",
			Content:    "first document",
		},
		{
			ID:         "chunk-2",
			DocumentID: "doc-2",
			Content:    "second document",
		},
	}

	if err := repo.SaveBatch(context.Background(), chunks); err != nil {
		t.Fatalf("SaveBatch() returned error: %v", err)
	}

	got, err := repo.FindByDocumentID(
		context.Background(),
		"doc-1",
	)
	if err != nil {
		t.Fatalf("FindByDocumentID() returned error: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(got))
	}

	if got[0].DocumentID != "doc-1" {
		t.Fatalf("unexpected document ID: %q", got[0].DocumentID)
	}
}

func TestChunkRepositoryReturnsCopy(t *testing.T) {
	repo := NewChunkRepository()

	chunks := []domain.Chunk{
		{
			ID:         "chunk-1",
			DocumentID: "doc-1",
			Content:    "original",
		},
	}

	if err := repo.SaveBatch(context.Background(), chunks); err != nil {
		t.Fatalf("SaveBatch() returned error: %v", err)
	}

	got, err := repo.FindByDocumentID(
		context.Background(),
		"doc-1",
	)
	if err != nil {
		t.Fatalf("FindByDocumentID() returned error: %v", err)
	}

	got[0].Content = "modified outside repository"

	gotAgain, err := repo.FindByDocumentID(
		context.Background(),
		"doc-1",
	)
	if err != nil {
		t.Fatalf("FindByDocumentID() returned error: %v", err)
	}

	if gotAgain[0].Content != "original" {
		t.Fatal("repository exposed its internal slice")
	}
}

func TestChunkRepositoryFindAll(t *testing.T) {
	repo := NewChunkRepository()

	input := []domain.Chunk{
		{
			ID:         "chunk-1",
			DocumentID: "doc-1",
			Content:    "first",
		},
		{
			ID:         "chunk-2",
			DocumentID: "doc-2",
			Content:    "second",
		},
	}

	if err := repo.SaveBatch(context.Background(), input); err != nil {
		t.Fatalf("SaveBatch() returned error: %v", err)
	}

	got, err := repo.FindAll(context.Background())
	if err != nil {
		t.Fatalf("FindAll() returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d chunks, want 2", len(got))
	}

	gotByID := make(map[string]domain.Chunk, len(got))
	for _, chunk := range got {
		gotByID[chunk.ID] = chunk
	}

	for _, want := range input {
		gotChunk, ok := gotByID[want.ID]
		if !ok {
			t.Fatalf("missing chunk %q", want.ID)
		}

		if gotChunk.Content != want.Content {
			t.Fatalf(
				"chunk %q content = %q, want %q",
				want.ID,
				gotChunk.Content,
				want.Content,
			)
		}
	}
}
