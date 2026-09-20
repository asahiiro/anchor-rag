package application

import (
	"context"
	"errors"
	"testing"

	"github.com/asahiiro/anchor-rag/internal/chunker"
	"github.com/asahiiro/anchor-rag/internal/embedding"
	"github.com/asahiiro/anchor-rag/internal/repository"
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
	textEmbedder, err := embedding.NewRuneFrequencyEmbedder(32)
	if err != nil {
		t.Fatalf("embedding constructor returned error: %v", err)
	}

	writer := memory.NewKnowledgeWriter(
		documentRepo,
		chunkRepo,
	)
	deleter := memory.NewKnowledgeDeleter(
		documentRepo,
		chunkRepo,
	)

	service := NewDocumentService(
		writer,
		documentRepo,
		deleter,
		textChunker,
		textEmbedder,
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

		if len(storedChunk.Embedding) != 32 {
			t.Fatalf(
				"chunk embedding dimension = %d, want 32",
				len(storedChunk.Embedding),
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

	textEmbedder, err := embedding.NewRuneFrequencyEmbedder(32)
	if err != nil {
		t.Fatalf("embedding constructor returned error: %v", err)
	}

	writer := memory.NewKnowledgeWriter(
		documentRepo,
		chunkRepo,
	)
	deleter := memory.NewKnowledgeDeleter(
		documentRepo,
		chunkRepo,
	)

	service := NewDocumentService(
		writer,
		documentRepo,
		deleter,
		textChunker,
		textEmbedder,
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

func TestDocumentServiceGet(t *testing.T) {
	documentRepo := memory.NewDocumentRepository()
	chunkRepo := memory.NewChunkRepository()

	writer := memory.NewKnowledgeWriter(
		documentRepo,
		chunkRepo,
	)
	deleter := memory.NewKnowledgeDeleter(
		documentRepo,
		chunkRepo,
	)

	textChunker, err := chunker.New(5, 2)
	if err != nil {
		t.Fatalf("chunker.New() returned error: %v", err)
	}

	textEmbedder, err := embedding.NewRuneFrequencyEmbedder(32)
	if err != nil {
		t.Fatalf(
			"embedding constructor returned error: %v",
			err,
		)
	}

	service := NewDocumentService(
		writer,
		documentRepo,
		deleter,
		textChunker,
		textEmbedder,
	)

	created, err := service.Create(
		context.Background(),
		"get-document.md",
		"retrievable document content",
	)
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	got, err := service.Get(
		context.Background(),
		created.ID,
	)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	if got.ID != created.ID {
		t.Fatalf(
			"document ID = %q, want %q",
			got.ID,
			created.ID,
		)
	}

	if got.Name != created.Name {
		t.Fatalf(
			"document name = %q, want %q",
			got.Name,
			created.Name,
		)
	}

	if got.Content != created.Content {
		t.Fatalf(
			"document content = %q, want %q",
			got.Content,
			created.Content,
		)
	}
}

func TestDocumentServiceGetReturnsNotFound(t *testing.T) {
	documentRepo := memory.NewDocumentRepository()
	chunkRepo := memory.NewChunkRepository()

	writer := memory.NewKnowledgeWriter(
		documentRepo,
		chunkRepo,
	)
	deleter := memory.NewKnowledgeDeleter(
		documentRepo,
		chunkRepo,
	)

	textChunker, err := chunker.New(5, 2)
	if err != nil {
		t.Fatalf("chunker.New() returned error: %v", err)
	}

	textEmbedder, err := embedding.NewRuneFrequencyEmbedder(32)
	if err != nil {
		t.Fatalf(
			"embedding constructor returned error: %v",
			err,
		)
	}

	service := NewDocumentService(
		writer,
		documentRepo,
		deleter,
		textChunker,
		textEmbedder,
	)

	_, err = service.Get(
		context.Background(),
		uuid.NewString(),
	)
	if !errors.Is(err, repository.ErrDocumentNotFound) {
		t.Fatalf(
			"expected ErrDocumentNotFound, got %v",
			err,
		)
	}

}

func TestDocumentServiceGetRejectsInvalidID(t *testing.T) {
	documentRepo := memory.NewDocumentRepository()
	chunkRepo := memory.NewChunkRepository()

	writer := memory.NewKnowledgeWriter(
		documentRepo,
		chunkRepo,
	)
	deleter := memory.NewKnowledgeDeleter(
		documentRepo,
		chunkRepo,
	)

	textChunker, err := chunker.New(5, 2)
	if err != nil {
		t.Fatalf("chunker.New() returned error: %v", err)
	}

	textEmbedder, err := embedding.NewRuneFrequencyEmbedder(32)
	if err != nil {
		t.Fatalf(
			"embedding constructor returned error: %v",
			err,
		)
	}

	service := NewDocumentService(
		writer,
		documentRepo,
		deleter,
		textChunker,
		textEmbedder,
	)

	_, err = service.Get(
		context.Background(),
		"not-a-uuid",
	)
	if !errors.Is(err, ErrInvalidDocumentID) {
		t.Fatalf(
			"expected ErrInvalidDocumentID, got %v",
			err,
		)
	}

}

func TestDocumentServiceDelete(t *testing.T) {

	documentRepo := memory.NewDocumentRepository()
	chunkRepo := memory.NewChunkRepository()

	writer := memory.NewKnowledgeWriter(
		documentRepo,
		chunkRepo,
	)
	deleter := memory.NewKnowledgeDeleter(
		documentRepo,
		chunkRepo,
	)

	textChunker, err := chunker.New(5, 2)
	if err != nil {
		t.Fatalf("chunker.New() returned error: %v", err)
	}

	textEmbedder, err := embedding.NewRuneFrequencyEmbedder(32)
	if err != nil {
		t.Fatalf(
			"embedding constructor returned error: %v",
			err,
		)
	}

	service := NewDocumentService(
		writer,
		documentRepo,
		deleter,
		textChunker,
		textEmbedder,
	)

	created, err := service.Create(
		context.Background(),
		"delete-document.md",
		"document content",
	)
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	_, err = documentRepo.FindByID(
		context.Background(),
		created.ID,
	)

	if err != nil {
		t.Fatalf("document was not saved: %v", err)
	}

	storedChunks, err := chunkRepo.FindByDocumentID(
		context.Background(),
		created.ID,
	)
	if err != nil {
		t.Fatalf("find chunks before deletion: %v", err)
	}

	if len(storedChunks) == 0 {
		t.Fatal("expected chunks before deletion")
	}

	if err = service.Delete(
		context.Background(),
		created.ID,
	); err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	_, err = service.Get(
		context.Background(),
		created.ID,
	)

	if !errors.Is(err, repository.ErrDocumentNotFound) {
		t.Fatalf("expected ErrDocumentNotFound, got %v", err)
	}

	gotChunks, err := chunkRepo.FindByDocumentID(
		context.Background(),
		created.ID,
	)
	if err != nil {
		t.Fatalf("find chunks after deletion: %v", err)
	}

	if len(gotChunks) != 0 {
		t.Fatalf("expected 0 chunks, got %v", len(gotChunks))
	}
}

func TestDocumentServiceDeleteRejectsInvalidID(t *testing.T) {
	documentRepo := memory.NewDocumentRepository()
	chunkRepo := memory.NewChunkRepository()

	writer := memory.NewKnowledgeWriter(
		documentRepo,
		chunkRepo,
	)
	deleter := memory.NewKnowledgeDeleter(
		documentRepo,
		chunkRepo,
	)

	textChunker, err := chunker.New(5, 2)
	if err != nil {
		t.Fatalf("chunker.New() returned error: %v", err)
	}

	textEmbedder, err := embedding.NewRuneFrequencyEmbedder(32)
	if err != nil {
		t.Fatalf(
			"embedding constructor returned error: %v",
			err,
		)
	}

	service := NewDocumentService(
		writer,
		documentRepo,
		deleter,
		textChunker,
		textEmbedder,
	)

	err = service.Delete(
		context.Background(),
		"not-a-uuid",
	)
	if !errors.Is(err, ErrInvalidDocumentID) {
		t.Fatalf(
			"expected ErrInvalidDocumentID, got %v",
			err,
		)
	}
}

func TestDocumentServiceDeleteReturnsNotFound(t *testing.T) {
	documentRepo := memory.NewDocumentRepository()
	chunkRepo := memory.NewChunkRepository()

	writer := memory.NewKnowledgeWriter(
		documentRepo,
		chunkRepo,
	)
	deleter := memory.NewKnowledgeDeleter(
		documentRepo,
		chunkRepo,
	)

	textChunker, err := chunker.New(5, 2)
	if err != nil {
		t.Fatalf("chunker.New() returned error: %v", err)
	}

	textEmbedder, err := embedding.NewRuneFrequencyEmbedder(32)
	if err != nil {
		t.Fatalf(
			"embedding constructor returned error: %v",
			err,
		)
	}

	service := NewDocumentService(
		writer,
		documentRepo,
		deleter,
		textChunker,
		textEmbedder,
	)

	err = service.Delete(
		context.Background(),
		uuid.NewString(),
	)
	if !errors.Is(err, repository.ErrDocumentNotFound) {
		t.Fatalf(
			"expected ErrDocumentNotFound, got %v",
			err,
		)
	}
}
