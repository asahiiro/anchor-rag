package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/repository"
)

func TestDocumentRepositorySaveAndFind(t *testing.T) {
	repo := NewDocumentRepository()

	doc, err := domain.NewDocument(
		"doc-1",
		"rag-notes.md",
		"RAG combines retrieval with generation.",
	)
	if err != nil {
		t.Fatalf("NewDocument() returned error: %v", err)
	}

	if err := repo.Save(context.Background(), doc); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	got, err := repo.FindByID(context.Background(), doc.ID)

	if err != nil {
		t.Fatalf("FindByID() returned error: %v", err)
	}

	if got != doc {
		t.Fatalf("FindByID() = %#v, want %#v", got, doc)
	}
}

func TestDocumentRepositoryReturnsNotFound(t *testing.T) {
	repo := NewDocumentRepository()

	_, err := repo.FindByID(context.Background(), "missing")
	if !errors.Is(err, repository.ErrDocumentNotFound) {
		t.Fatalf("expected ErrDocumentNotFound, got %v:", err)
	}
}

func TestDocumentRepositoryRespectsCancelledContext(t *testing.T) {
	repo := NewDocumentRepository()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.FindByID(ctx, "doc-1")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v:", err)
	}
}
