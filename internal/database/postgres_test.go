package database

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewPostgresPoolRejectsMissingURL(t *testing.T) {
	_, err := NewPostgresPool(
		context.Background(),
		"     ",
	)

	if !errors.Is(err, ErrMissingDatabaseURL) {
		t.Fatalf(
			"expected ErrMissingDatabaseURL, got %v",
			err,
		)
	}
}

func TestNewPostgresPoolIntegration(t *testing.T) {
	databaseURL := strings.TrimSpace(
		os.Getenv("DATABASE_URL"),
	)
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not configured")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	pool, err := NewPostgresPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf(
			"NewPostgresPool() returned error: %v",
			err,
		)
	}
	defer pool.Close()

	var result int
	if err := pool.QueryRow(
		ctx,
		"SELECT 1",
	).Scan(&result); err != nil {
		t.Fatalf("SELECT 1 failed: %v", err)
	}

	if result != 1 {
		t.Fatalf("result = %d, want 1", result)
	}
}
