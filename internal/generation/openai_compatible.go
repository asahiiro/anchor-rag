package generation

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
	ErrMissingGenerationBaseURL = errors.New(
		"generation base URL is required",
	)
	ErrMissingGenerationModel = errors.New(
		"generation model is required",
	)
	ErrEmptyGenerationMessages = errors.New(
		"generation messages must not be empty",
	)
	ErrInvalidGenerationResponse = errors.New(
		"invalid generation response",
	)
)

type OpenAICompatibleConfig struct {
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

type OpenAICompatibleGenerator struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

var _ Generator = (*OpenAICompatibleGenerator)(nil)

func NewOpenAICompatibleGenerator(
	config OpenAICompatibleConfig,
) (*OpenAICompatibleGenerator, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(config.BaseURL), "/")

	if baseURL == "" {
		return nil, ErrMissingGenerationBaseURL
	}

	model := strings.TrimSpace(config.Model)
	if model == "" {
		return nil, ErrMissingGenerationModel
	}

	client := config.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: 60 * time.Second,
		}
	}

	return &OpenAICompatibleGenerator{
		baseURL: baseURL,
		apiKey:  config.APIKey,
		model:   model,
		client:  client,
	}, nil
}

type chatCompletionAPIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type chatCompletionAPIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (g *OpenAICompatibleGenerator) Generate(
	ctx context.Context,
	messages []Message,
) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	if len(messages) == 0 {
		return "", ErrEmptyGenerationMessages
	}

	payload, err := json.Marshal(chatCompletionAPIRequest{
		Model:    g.model,
		Messages: messages,
	})
	if err != nil {
		return "", fmt.Errorf(
			"encode generation request: %w",
			err,
		)
	}

	endpoint := g.baseURL + "/chat/completions"

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return "", fmt.Errorf(
			"create generation request: %w",
			err,
		)
	}

	request.Header.Set("Content-Type", "application/json")

	if g.apiKey != "" {
		request.Header.Set(
			"Authorization",
			"Bearer "+g.apiKey,
		)
	}

	response, err := g.client.Do(request)
	if err != nil {
		return "", fmt.Errorf(
			"send generation request: %w",
			err,
		)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(
			io.LimitReader(response.Body, 64<<10),
		)

		return "", fmt.Errorf(
			"generation API returned %s: %s",
			response.Status,
			strings.TrimSpace(string(body)),
		)
	}

	var decoded chatCompletionAPIResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf(
			"decode generation response: %w",
			err,
		)
	}

	if len(decoded.Choices) == 0 {
		return "", ErrInvalidGenerationResponse
	}

	answer := strings.TrimSpace(decoded.Choices[0].Message.Content)

	if answer == "" {
		return "", ErrInvalidGenerationResponse
	}

	return answer, nil
}
