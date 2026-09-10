package embedding

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestOpenAICompatibleEmbedder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}

			if r.URL.Path != "/v1/embeddings" {
				t.Errorf(
					"path = %q, want /v1/embeddings",
					r.URL.Path,
				)
			}
			if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
				t.Errorf(
					"Authorization = %q, want Bearer test-key",
					got,
				)
			}

			var requestBody embeddingAPIRequest
			if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
				t.Errorf("failed to decode request: %v", err)
			}

			if requestBody.Model != "test-model" {
				t.Errorf(
					"model = %q, want test-model",
					requestBody.Model,
				)
			}

			if !reflect.DeepEqual(
				requestBody.Input,
				[]string{"golang", "RAG"},
			) {
				t.Errorf(
					"input = %#v",
					requestBody.Input,
				)
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"data": [
				{
					"index": 1,
					"embedding": [0,1]
				},
				{
					"index": 0,
					"embedding": [1, 0]
					}
				]
			}`))
		},
	))
	defer server.Close()

	embedder, err := NewOpenAICompatibleEmbedder(
		OpenAICompatibleConfig{
			BaseURL: server.URL + "/v1/",
			APIKey:  "test-key",
			Model:   "test-model",
		},
	)
	if err != nil {
		t.Fatalf("constructor returned error: %v", err)
	}

	got, err := embedder.Embed(
		context.Background(),
		[]string{"golang", "RAG"},
	)
	if err != nil {
		t.Fatalf("Embed() returned error: %v", err)
	}

	want := [][]float32{
		{1, 0},
		{0, 1},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Embed() = %#v, want %#v", got, want)
	}
}

func TestOpenAICompatibleEmbedderReturnsAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		},
	))
	defer server.Close()

	embedder, err := NewOpenAICompatibleEmbedder(
		OpenAICompatibleConfig{
			BaseURL: server.URL,
			Model:   "test-model",
		},
	)
	if err != nil {
		t.Fatalf("constructor returned error: %v", err)
	}

	_, err = embedder.Embed(
		context.Background(),
		[]string{"RAG"},
	)
	if err == nil {
		t.Fatalf("expected API error")
	}

	if !strings.Contains(err.Error(), "429") {
		t.Fatalf("error = %q, want status 429", err.Error())
	}
}

func TestOpenAICompatibleEmbedderRejectsEmptyInput(t *testing.T) {
	embedder, err := NewOpenAICompatibleEmbedder(
		OpenAICompatibleConfig{
			BaseURL: "http://example.com/v1",
			Model:   "test-model",
		},
	)
	if err != nil {
		t.Fatalf("constructor returned error: %v", err)
	}

	_, err = embedder.Embed(
		context.Background(),
		[]string{"     "},
	)
	if !errors.Is(err, ErrEmptyEmbeddingInput) {
		t.Fatalf("expected ErrEmptyEmbeddingInput, got %v", err)
	}
}

func TestOpenAICompatibleEmbedderValidatesConfig(t *testing.T) {
	_, err := NewOpenAICompatibleEmbedder(
		OpenAICompatibleConfig{Model: "test-model"},
	)
	if !errors.Is(err, ErrMissingEmbeddingBaseURL) {
		t.Fatalf("expected ErrMissingEmbeddingBaseURL, got %v", err)
	}

	_, err = NewOpenAICompatibleEmbedder(
		OpenAICompatibleConfig{BaseURL: "http://example.com/v1"},
	)
	if !errors.Is(err, ErrMissingEmbeddingModel) {
		t.Fatalf("expected ErrMissingEmbeddingModel, got %v", err)
	}
}
