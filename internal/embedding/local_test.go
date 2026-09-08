package embedding

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestRuneFrequencyEmbedder(t *testing.T) {
	embedder, err := NewRuneFrequencyEmbedder(32)
	if err != nil {
		t.Fatalf("constructor returned error: %v", err)
	}

	got, err := embedder.Embed(
		context.Background(),
		[]string{"Go!", "go"},
	)
	if err != nil {
		t.Fatalf("Embed() returned error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("got %d embeddings, want 2", len(got))
	}

	if len(got[0]) != 32 {
		t.Fatalf("dimension = %d, want 32", len(got[0]))
	}

	if !reflect.DeepEqual(got[0], got[1]) {
		t.Fatalf(
			"case and punctuation should be ignored: %#v != %#v",
			got[0],
			got[1],
		)
	}
}

func TestRuneFrequencyEmbedderRejectsInvalidDimension(t *testing.T) {
	_, err := NewRuneFrequencyEmbedder(0)

	if !errors.Is(err, ErrInvalidDimension) {
		t.Fatalf("expected ErrInvalidDimension, got %v", err)
	}
}

func TestRuneFrequencyEmbedderRespectsCancelledContext(t *testing.T) {
	embedder, err := NewRuneFrequencyEmbedder(32)
	if err != nil {
		t.Fatalf("constructor returned error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = embedder.Embed(ctx, []string{"RAG"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
