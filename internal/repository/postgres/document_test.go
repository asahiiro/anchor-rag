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

	repo := NewDocumentRepository(pool)

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

	if err := repo.Save(ctx, doc); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	got, err := repo.FindByID(ctx, doc.ID)
	if err != nil {
		t.Fatalf("FindByID() returned error: %v", err)
	}

	if got.ID != doc.ID ||
		got.Name != doc.Name ||
		got.Content != doc.Content ||
		!got.CreatedAt.Equal(doc.CreatedAt) {
		t.Fatalf(
			"FindByID() = %#v, want %#v",
			got,
			doc,
		)
	}

	_, err = repo.FindByID(
		ctx,
		uuid.NewString(),
	)
	if !errors.Is(err, repository.ErrDocumentNotFound) {
		t.Fatalf(
			"expected ErrDocumentNotFound, got %v",
			err,
		)
	}
}
