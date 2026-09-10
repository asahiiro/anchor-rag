package embedding

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var ErrUnsupportedProvider = errors.New(
	"unsupported embedding provider",
)

type Config struct {
	Provider       string
	BaseURL        string
	APIKey         string
	Model          string
	LocalDimension int
	HTTPClient     *http.Client
}

func NewFromConfig(config Config) (Embedder, error) {
	provider := strings.ToLower(
		strings.TrimSpace(config.Provider),
	)

	switch provider {
	case "", "local":
		return NewRuneFrequencyEmbedder(config.LocalDimension)

	case "openai-compatible":
		return NewOpenAICompatibleEmbedder(
			OpenAICompatibleConfig{
				BaseURL:    config.BaseURL,
				APIKey:     config.APIKey,
				Model:      config.Model,
				HTTPClient: config.HTTPClient,
			},
		)

	default:
		return nil, fmt.Errorf(
			"%w: %s",
			ErrUnsupportedProvider,
			provider,
		)
	}
}
