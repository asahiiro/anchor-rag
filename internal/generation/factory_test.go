package generation

import (
	"errors"
	"testing"
)

func TestNewFromConfigCreatesOpenAICompatibleGenerator(t *testing.T) {
	got, err := NewFromConfig(Config{
		Provider: " OpenAI-Compatible ",
		BaseURL:  "https://api.example.com/v1",
		APIKey:   "test-key",
		Model:    "test-model",
	})
	if err != nil {
		t.Fatalf("NewFromConfig() returned error: %v", err)
	}

	if _, ok := got.(*OpenAICompatibleGenerator); !ok {
		t.Fatalf("unexpected generator type: %T", got)
	}
}

func TestNewFromConfigRejectsMissingProvider(t *testing.T) {
	_, err := NewFromConfig(Config{})

	if !errors.Is(err, ErrMissingProvider) {
		t.Fatalf(
			"expected ErrMissingProvider, got %v",
			err,
		)
	}
}

func TestNewFromConfigRejectsUnknownProvider(t *testing.T) {
	_, err := NewFromConfig(Config{
		Provider: "unknown",
	})

	if !errors.Is(err, ErrUnsupportedProvider) {
		t.Fatalf(
			"expected ErrUnsupportedProvider, got %v",
			err,
		)
	}
}

func TestNewFromConfigValidatesCompatibleConfig(t *testing.T) {
	_, err := NewFromConfig(Config{
		Provider: "openai-compatible",
		Model:    "test-model",
	})
	if !errors.Is(err, ErrMissingGenerationBaseURL) {
		t.Fatalf(
			"expected ErrMissingGenerationBaseURL, got %v",
			err,
		)
	}
}
