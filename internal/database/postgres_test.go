package database

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pgvector/pgvector-go"
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

	wantVector := []float32{
		1,
		2,
		3,
	}

	var gotVector pgvector.Vector
	if err := pool.QueryRow(
		ctx,
		"SELECT $1::vector",
		pgvector.NewVector(wantVector),
	).Scan(&gotVector); err != nil {
		t.Fatalf(
			"vector round trip failed: %v",
			err,
		)
	}

	if !reflect.DeepEqual(
		gotVector.Slice(),
		wantVector,
	) {
		t.Fatalf(
			"vector = %#v, want %#v",
			gotVector.Slice(),
			wantVector,
		)
	}
}
