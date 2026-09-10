package embedding

import (
	"errors"
	"testing"
)

func TestNewFromConfigCreatesLocalEmbedder(t *testing.T) {
	got, err := NewFromConfig(Config{
		Provider:       "local",
		LocalDimension: 256,
	})
	if err != nil {
		t.Fatalf("NewFromConfig() returned error: %v", err)
	}

	if _, ok := got.(*RuneFrequencyEmbedder); !ok {
		t.Fatalf("unexpected embedder type: %T", got)
	}
}

func TestNewFromConfigCreatesOpenAICompatibleEmbedder(t *testing.T) {
	got, err := NewFromConfig(Config{
		Provider: "openai-compatible",
		BaseURL:  "https://api.openai.com/v1",
		APIKey:   "test-key",
		Model:    "test-model",
	})
	if err != nil {
		t.Fatalf("NewFromConfig() returned error: %v", err)
	}

	if _, ok := got.(*OpenAICompatibleEmbedder); !ok {
		t.Fatalf("unexpected embedder type: %T", got)
	}
}

func TestNewFromConfigRejectsUnknownProvider(t *testing.T) {
	_, err := NewFromConfig(Config{
		Provider: "unknown",
	})

	if !errors.Is(err, ErrUnsupportedProvider) {
		t.Fatalf("expected ErrUnsupportedProvider, got %v", err)
	}
}

func TestNewFromConfigValidatesLocalDimension(t *testing.T) {
	_, err := NewFromConfig(Config{
		Provider:       "local",
		LocalDimension: 0,
	})

	if !errors.Is(err, ErrInvalidDimension) {
		t.Fatalf("expected ErrInvalidDimension, got %v", err)
	}
}
