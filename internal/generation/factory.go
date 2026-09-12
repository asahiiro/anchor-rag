package generation

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrMissingProvider = errors.New(
		"generation provider is required",
	)
	ErrUnsupportedProvider = errors.New(
		"unsupported generation provider",
	)
)

type Config struct {
	Provider   string
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

func NewFromConfig(config Config) (Generator, error) {
	provider := strings.ToLower(
		strings.TrimSpace(config.Provider),
	)

	if provider == "" {
		return nil, ErrMissingProvider
	}

	switch provider {
	case "openai-compatible":
		return NewOpenAICompatibleGenerator(
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
