package embedding

import (
	"context"
	"errors"
	"strings"
	"unicode"
)

var ErrInvalidDimension = errors.New(
	"embedding dimension must be greater than zero",
)

type RuneFrequencyEmbedder struct {
	dimension int
}

var _ Embedder = (*RuneFrequencyEmbedder)(nil)

func NewRuneFrequencyEmbedder(
	dimension int,
) (*RuneFrequencyEmbedder, error) {
	if dimension <= 0 {
		return nil, ErrInvalidDimension
	}

	return &RuneFrequencyEmbedder{
		dimension: dimension,
	}, nil
}

func (e *RuneFrequencyEmbedder) Embed(
	ctx context.Context,
	texts []string,
) ([][]float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(texts))
	for textIndex, text := range texts {
		vector := make([]float32, e.dimension)

		for _, currentRune := range strings.ToLower(text) {
			if unicode.IsSpace(currentRune) ||
				unicode.IsPunct(currentRune) {
				continue
			}
			vectorIndex := int(currentRune) % e.dimension
			vector[vectorIndex]++
		}

		embeddings[textIndex] = vector
	}

	return embeddings, nil
}
