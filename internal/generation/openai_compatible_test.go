package generation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAICompatibleGenerator(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}

			if r.URL.Path != "/v1/chat/completions" {
				t.Errorf(
					"path = %q, want /v1/chat/completions",
					r.URL.Path,
				)
			}

			if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
				t.Errorf(
					"Authorization = %q, want Bearer test-key",
					got,
				)
			}

			var body chatCompletionAPIRequest
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode request: %v", err)
			}

			if body.Model != "test-model" {
				t.Errorf(
					"model = %q, want test-model",
					body.Model,
				)
			}

			if len(body.Messages) != 2 {
				t.Fatalf(
					"message count = %d, want 2",
					len(body.Messages),
				)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"choices": [
				{
					"message": {
						"role": "assistant",
						"content": "Go uses goroutines [1]."
					}
				}
				]
			}`))
		},
	))
	defer server.Close()

	generator, err := NewOpenAICompatibleGenerator(
		OpenAICompatibleConfig{
			BaseURL: server.URL + "/v1/",
			APIKey:  "test-key",
			Model:   "test-model",
		},
	)
	if err != nil {
		t.Fatalf("constructor returned error: %v", err)
	}

	got, err := generator.Generate(
		context.Background(),
		[]Message{
			{
				Role:    RoleSystem,
				Content: "Answer using the context.",
			},
			{
				Role:    RoleUser,
				Content: "What is Go?",
			},
		},
	)
	if err != nil {
		t.Fatalf("generate() returned error: %v", err)
	}

	want := "Go uses goroutines [1]."
	if got != want {
		t.Fatalf("Generate() = %q, want %q", got, want)
	}
}

func TestOpenAICompatibleGeneratorTrimTrailingSlashes(t *testing.T) {
	generator, err := NewOpenAICompatibleGenerator(
		OpenAICompatibleConfig{
			BaseURL: "http://example.com/v1///",
			Model:   "test-model",
		},
	)
	if err != nil {
		t.Fatalf("constructor returned error: %v", err)
	}

	want := "http://example.com/v1"
	if generator.baseURL != want {
		t.Fatalf(
			"baseURL = %q, want %q",
			generator.baseURL,
			want,
		)
	}
}

func TestOpenAICompatibleGeneratorReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			http.Error(
				w,
				"rate limit exceeded",
				http.StatusTooManyRequests,
			)
		},
	))
	defer server.Close()

	generator, err := NewOpenAICompatibleGenerator(
		OpenAICompatibleConfig{
			BaseURL: server.URL,
			Model:   "test-model",
		},
	)
	if err != nil {
		t.Fatalf("constructor returned error: %v", err)
	}

	_, err = generator.Generate(
		context.Background(),
		[]Message{
			{
				Role:    RoleUser,
				Content: "What is RAG?",
			},
		},
	)
	if err == nil {
		t.Fatal("expected API error")
	}

	if !strings.Contains(err.Error(), "429") {
		t.Fatalf("error = %q, want status 429", err.Error())
	}
}

func TestOpenAICompatibleGeneratorRejectsEmptyMessages(t *testing.T) {
	generator, err := NewOpenAICompatibleGenerator(
		OpenAICompatibleConfig{
			BaseURL: "http://example.com/v1",
			Model:   "test-model",
		},
	)
	if err != nil {
		t.Fatalf("constructor returned error: %v", err)
	}

	_, err = generator.Generate(
		context.Background(),
		nil,
	)
	if !errors.Is(err, ErrEmptyGenerationMessages) {
		t.Fatalf(
			"expected ErrEmptyGenerationMessages, got %v",
			err,
		)
	}
}

func TestOpenAICompatibleGeneratorValidatesConfig(t *testing.T) {
	_, err := NewOpenAICompatibleGenerator(
		OpenAICompatibleConfig{
			Model: "test-model",
		},
	)
	if !errors.Is(err, ErrMissingGenerationBaseURL) {
		t.Fatalf(
			"expected ErrMissingGenerationBaseURL, got %v",
			err,
		)
	}

	_, err = NewOpenAICompatibleGenerator(
		OpenAICompatibleConfig{
			BaseURL: "http://example.com/v1",
		},
	)
	if !errors.Is(err, ErrMissingGenerationModel) {
		t.Fatalf(
			"expected ErrMissingGenerationModel, got %v",
			err,
		)
	}
}
