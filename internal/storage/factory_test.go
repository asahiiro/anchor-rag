package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/asahiiro/anchor-rag/internal/database"
	"github.com/asahiiro/anchor-rag/internal/repository/memory"
)

func TestNewCreatesMemoryStore(t *testing.T) {
	store, err := New(
		context.Background(),
		Config{
			Provider: " memory ",
		},
	)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	t.Cleanup(store.Close)

	if _, ok := store.DocumentRepository.(*memory.DocumentRepository); !ok {
		t.Fatalf(
			"unexpected document repository: %T",
			store.DocumentRepository,
		)
	}

	if _, ok := store.ChunkRepository.(*memory.ChunkRepository); !ok {
		t.Fatalf(
			"unexpected chunk repository: %T",
			store.ChunkRepository,
		)
	}
}

func TestNewPostgresRequiresDatabaseURL(t *testing.T) {
	_, err := New(
		context.Background(),
		Config{
			Provider: "postgres",
		},
	)

	if !errors.Is(err, database.ErrMissingDatabaseURL) {
		t.Fatalf(
			"expected ErrMissingDatabaseURL, got %v",
			err,
		)
	}
}

func TestNewRejectsUnknownProvider(t *testing.T) {
	_, err := New(
		context.Background(),
		Config{
			Provider: "unknown",
		},
	)

	if !errors.Is(err, ErrUnsupportedProvider) {
		t.Fatalf(
			"expected ErrUnsupportedProvider, got %v",
			err,
		)
	}
}
