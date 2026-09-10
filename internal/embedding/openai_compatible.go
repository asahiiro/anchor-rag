package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrMissingEmbeddingBaseURL = errors.New(
		"embedding base URL is required",
	)
	ErrMissingEmbeddingModel = errors.New(
		"embedding model is required",
	)
	ErrEmptyEmbeddingInput = errors.New(
		"embedding input must not be empty",
	)
	ErrInvalidEmbeddingResponse = errors.New(
		"invalid embedding response",
	)
)

type OpenAICompatibleConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

type OpenAICompatibleEmbedder struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

var _ Embedder = (*OpenAICompatibleEmbedder)(nil)

func NewOpenAICompatibleEmbedder(
	config OpenAICompatibleConfig,
) (*OpenAICompatibleEmbedder, error) {
	baseURL := strings.TrimRight(
		strings.TrimSpace(config.BaseURL),
		"/",
	)
	if baseURL == "" {
		return nil, ErrMissingEmbeddingBaseURL
	}

	model := strings.TrimSpace(config.Model)
	if model == "" {
		return nil, ErrMissingEmbeddingModel
	}

	client := config.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &OpenAICompatibleEmbedder{
		baseURL: baseURL,
		apiKey:  config.APIKey,
		model:   model,
		client:  client,
	}, nil
}

type embeddingAPIRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format"`
}

type embeddingAPIItem struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

type embeddingAPIResponse struct {
	Data []embeddingAPIItem `json:"data"`
}

func (e *OpenAICompatibleEmbedder) Embed(
	ctx context.Context,
	texts []string,
) ([][]float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	for _, text := range texts {
		if strings.TrimSpace(text) == "" {
			return nil, ErrEmptyEmbeddingInput
		}
	}

	payload, err := json.Marshal(embeddingAPIRequest{
		Input:          texts,
		Model:          e.model,
		EncodingFormat: "float",
	})
	if err != nil {
		return nil, fmt.Errorf("encode embedding request: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		e.baseURL+"/embeddings",
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("create embedding request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		request.Header.Set(
			"Authorization",
			"Bearer "+e.apiKey,
		)
	}
	response, err := e.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send embedding request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 64<<10))

		return nil, fmt.Errorf(
			"embedding API returned %s: %s",
			response.Status,
			strings.TrimSpace(string(body)),
		)
	}

	var decoded embeddingAPIResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}

	if len(decoded.Data) != len(texts) {
		return nil, fmt.Errorf(
			"%w: got %d vectors, want %d",
			ErrInvalidEmbeddingResponse,
			len(decoded.Data),
			len(texts),
		)
	}

	embeddings := make([][]float32, len(texts))
	seen := make([]bool, len(texts))

	for _, item := range decoded.Data {
		if item.Index < 0 || item.Index >= len(texts) {
			return nil, fmt.Errorf(
				"%w: index %d is out of range",
				ErrInvalidEmbeddingResponse,
				item.Index,
			)
		}

		if seen[item.Index] || len(item.Embedding) == 0 {
			return nil, ErrInvalidEmbeddingResponse
		}

		embeddings[item.Index] = item.Embedding
		seen[item.Index] = true
	}

	for _, found := range seen {
		if !found {
			return nil, ErrInvalidEmbeddingResponse
		}
	}

	return embeddings, nil
}
