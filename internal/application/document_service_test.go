package application

import (
	"context"
	"errors"
	"testing"

	"github.com/asahiiro/anchor-rag/internal/chunker"
	"github.com/asahiiro/anchor-rag/internal/repository/memory"
	"github.com/google/uuid"
)

func TestDocumentServiceCreate(t *testing.T) {
	documentRepo := memory.NewDocumentRepository()
	chunkRepo := memory.NewChunkRepository()
	textChunker, err := chunker.New(5, 2)
	if err != nil {
		t.Fatalf("chunker.New() returned error: %v", err)
	}

	service := NewDocumentService(
		documentRepo,
		chunkRepo,
		textChunker,
	)

	doc, err := service.Create(
		context.Background(),
		"  rag-notes.md  ",
		"abcdefghij",
	)
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	if _, err := uuid.Parse(doc.ID); err != nil {
		t.Fatalf("document ID is not a UUID: %q", doc.ID)
	}

	if doc.Name != "rag-notes.md" {
		t.Fatalf("unexpected document name: %q", doc.Name)
	}

	storedDocument, err := documentRepo.FindByID(
		context.Background(),
		doc.ID,
	)

	if err != nil {
		t.Fatalf("document was not saved: %v", err)
	}

	if storedDocument.ID != doc.ID {
		t.Fatalf(
			"stored document ID = %q, want %q",
			storedDocument.ID,
			doc.ID,
		)
	}

	storedChunks, err := chunkRepo.FindByDocumentID(
		context.Background(),
		doc.ID,
	)
	if err != nil {
		t.Fatalf("chunks were not saved: %v", err)
	}

	wantContents := []string{
		"abcde",
		"defgh",
		"ghij",
	}

	if len(storedChunks) != len(wantContents) {
		t.Fatalf(
			"got %d chunks, want %d",
			len(storedChunks),
			len(wantContents),
		)
	}
	for position, storedChunk := range storedChunks {
		if _, err := uuid.Parse(storedChunk.ID); err != nil {
			t.Fatalf(
				"chunk %d ID is not a UUID: %q",
				position,
				storedChunk.ID,
			)
		}

		if storedChunk.DocumentID != doc.ID {
			t.Fatalf(
				"chunk %d points to document %q",
				position,
				storedChunk.DocumentID,
			)
		}

		if storedChunk.Position != position {
			t.Fatalf(
				"chunk position = %d, want %d",
				storedChunk.Position,
				position,
			)
		}

		if storedChunk.Content != wantContents[position] {
			t.Fatalf(
				"chunk content = %q, want %q",
				storedChunk.Content,
				wantContents[position],
			)
		}
	}

}

func TestDocumentServiceRejectsInvalidDocument(t *testing.T) {
	documentRepo := memory.NewDocumentRepository()
	chunkRepo := memory.NewChunkRepository()
	textChunker, err := chunker.New(5, 2)
	if err != nil {
		t.Fatalf("chunker.New() returned error: %v", err)
	}

	service := NewDocumentService(
		documentRepo,
		chunkRepo,
		textChunker,
	)

	_, err = service.Create(
		context.Background(),
		"   ",
		"content",
	)
	if !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("expected ErrInvalidDocument, got %v", err)
	}
}
