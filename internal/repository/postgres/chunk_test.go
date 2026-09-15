package postgres

import (
	"context"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/asahiiro/anchor-rag/internal/database"
	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/google/uuid"
)

func TestChunkRepositoryIntegration(t *testing.T) {
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

	pool, err := database.NewPostgresPool(
		ctx,
		databaseURL,
	)
	if err != nil {
		t.Fatalf("connect to database: %v", err)
	}
	t.Cleanup(pool.Close)

	documentRepo := NewDocumentRepository(pool)
	chunkRepo := NewChunkRepository(pool)

	doc, err := domain.NewDocument(
		uuid.NewString(),
		"vector-storage.md",
		"Chunks are stored with vectors.",
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

		if _, err := pool.Exec(
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

	if err := documentRepo.Save(ctx, doc); err != nil {
		t.Fatalf("save document: %v", err)
	}

	firstEmbedding := make([]float32, 1024)
	firstEmbedding[0] = 1

	secondEmbedding := make([]float32, 1024)
	secondEmbedding[1] = 1

	chunks := []domain.Chunk{
		{
			ID:         uuid.NewString(),
			DocumentID: doc.ID,
			Content:    "first chunk",
			Position:   0,
			Embedding:  firstEmbedding,
		},
		{
			ID:         uuid.NewString(),
			DocumentID: doc.ID,
			Content:    "second chunk",
			Position:   1,
			Embedding:  secondEmbedding,
		},
	}

	if err := chunkRepo.SaveBatch(ctx, chunks); err != nil {
		t.Fatalf("SaveBatch() returned error: %v", err)
	}

	got, err := chunkRepo.FindByDocumentID(
		ctx,
		doc.ID,
	)
	if err != nil {
		t.Fatalf(
			"FindByDocumentID() returned error: %v",
			err,
		)
	}

	if !reflect.DeepEqual(got, chunks) {
		t.Fatalf(
			"FindByDocumentID() = %#v, want %#v",
			got,
			chunks,
		)
	}

	searchResults, err := chunkRepo.SearchSimilar(
		ctx,
		firstEmbedding,
		1,
	)
	if err != nil {
		t.Fatalf(
			"SearchSimilar() returned error: %v",
			err,
		)
	}

	if len(searchResults) != 1 {
		t.Fatalf(
			"got %d search results, want 1",
			len(searchResults),
		)
	}

	if searchResults[0].Chunk.ID != chunks[0].ID {
		t.Fatalf(
			"first result ID = %q, want %q",
			searchResults[0].Chunk.ID,
			chunks[0].ID,
		)
	}

	if math.Abs(searchResults[0].Score-1) > 1e-9 {
		t.Fatalf(
			"first result score = %v, want 1",
			searchResults[0].Score,
		)
	}

	allChunks, err := chunkRepo.FindAll(ctx)
	if err != nil {
		t.Fatalf("FindAll() returned error: %v", err)
	}

	found := make(map[string]bool)
	for _, chunk := range allChunks {
		found[chunk.ID] = true
	}

	for _, chunk := range chunks {
		if !found[chunk.ID] {
			t.Fatalf(
				"FindAll() is missing chunk %s",
				chunk.ID,
			)
		}
	}
}
