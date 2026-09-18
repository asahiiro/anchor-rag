package postgres

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/asahiiro/anchor-rag/internal/database"
	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestDocumentRepositoryIntegration(t *testing.T) {
	databaseURL := strings.TrimSpace(
		os.Getenv("DATABASE_URL"),
	)
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not configured")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(pool.Close)

	documentRepo := NewDocumentRepository(pool)
	chunkRepo := NewChunkRepository(pool)

	doc, err := domain.NewDocument(
		uuid.NewString(),
		"postgres-rag.md",
		"PostgreSQL can persist RAG documents.",
	)
	if err != nil {
		t.Fatalf("NewDocument() returned error: %v", err)
	}

	doc.CreatedAt = doc.CreatedAt.
		UTC().
		Truncate(time.Microsecond)

	embedding := make([]float32, 1024)
	embedding[0] = 1

	chunks := []domain.Chunk{
		{
			ID:         uuid.NewString(),
			DocumentID: doc.ID,
			Content:    "committed chunk",
			Position:   0,
			Embedding:  embedding,
		},
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cleanupCancel()

		if _, err = pool.Exec(
			cleanupCtx,
			"DELETE FROM documents WHERE id = $1",
			doc.ID,
		); err != nil {
			t.Errorf(
				"cleanup test document: %v",
				err,
			)
		}
	})

	if err := documentRepo.SaveDocument(ctx, doc, chunks); err != nil {
		t.Fatalf("SaveDocument() returned error: %v", err)
	}

	storedDocument, err := documentRepo.FindByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("FindByID() returned error: %v", err)
	}

	if storedDocument.ID != doc.ID ||
		storedDocument.Name != doc.Name ||
		storedDocument.Content != doc.Content ||
		!storedDocument.CreatedAt.Equal(doc.CreatedAt) {
		t.Fatalf(
			"FindByID() = %#v, want %#v",
			storedDocument,
			doc,
		)
	}

	_, err = documentRepo.FindByID(
		ctx,
		uuid.NewString(),
	)
	if !errors.Is(err, repository.ErrDocumentNotFound) {
		t.Fatalf(
			"expected ErrDocumentNotFound, got %v",
			err,
		)
	}

	storedChunks, err := chunkRepo.FindByDocumentID(
		ctx,
		doc.ID,
	)
	if err != nil {
		t.Fatalf("FindByDocumentID() returned error: %v", err)
	}

	if len(storedChunks) != 1 {
		t.Fatalf(
			"stored chunk count = %d, want 1",
			len(storedChunks),
		)
	}

	if storedChunks[0].ID != chunks[0].ID {
		t.Fatalf(
			"stored chunk ID = %q, want %q",
			storedChunks[0].ID,
			chunks[0].ID,
		)
	}
}

func TestDocumentRepositorySaveDocumentRollsBack(
	t *testing.T,
) {
	databaseURL := strings.TrimSpace(
		os.Getenv("DATABASE_URL"),
	)
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not configured")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	pool, err := database.NewPostgresPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(pool.Close)

	repo := NewDocumentRepository(pool)

	doc, err := domain.NewDocument(
		uuid.NewString(),
		"transaction-test.md",
		"Document and chunks must be saved atomically.",
	)
	if err != nil {
		t.Fatalf("NewDocument() returned error: %v", err)
	}

	embedding := make([]float32, 1024)
	embedding[0] = 1

	chunks := []domain.Chunk{
		{
			ID:         uuid.NewString(),
			DocumentID: doc.ID,
			Content:    "first chunk",
			Position:   0,
			Embedding:  embedding,
		},
		{
			ID:         uuid.NewString(),
			DocumentID: doc.ID,
			Content:    "duplicate position",
			Position:   0,
			Embedding:  embedding,
		},
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cleanupCancel()

		if _, cleanupErr := pool.Exec(
			cleanupCtx,
			"DELETE FROM documents WHERE id = $1",
			doc.ID,
		); cleanupErr != nil {
			t.Errorf("cleanup document: %v", cleanupErr)
		}
	})

	err = repo.SaveDocument(ctx, doc, chunks)
	if err == nil {
		t.Fatal("expected SaveDocument() to fail")
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("expected PostgreSQL error, got %v", err)
	}

	if pgErr.Code != "23505" {
		t.Fatalf(
			"PostgreSQL error code = %s, want unique violation",
			pgErr.Code,
		)
	}

	var documentCount int
	if err := pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM documents WHERE id = $1",
		doc.ID,
	).Scan(&documentCount); err != nil {
		t.Fatalf("count documents: %v", err)
	}

	var chunkCount int
	if err := pool.QueryRow(
		ctx,
		"SELECT COUNT(*) FROM chunks WHERE document_id = $1",
		doc.ID,
	).Scan(&chunkCount); err != nil {
		t.Fatalf("count chunks: %v", err)
	}

	if documentCount != 0 || chunkCount != 0 {
		t.Fatalf(
			"after rollback: documents = %d, chunks = %d; want both zero",
			documentCount,
			chunkCount,
		)
	}
}
