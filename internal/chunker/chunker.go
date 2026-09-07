package chunker

import (
	"errors"
	"strings"
)

var (
	ErrInvalidSize = errors.New(
		"chunk size must be greater than zero",
	)
	ErrInvalidOverlap = errors.New(
		"chunk overlap must be non-negative and smaller than chunk size",
	)
)

type Chunker struct {
	size    int
	overlap int
}

func New(size, overlap int) (*Chunker, error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}

	if overlap < 0 || overlap >= size {
		return nil, ErrInvalidOverlap
	}

	return &Chunker{
		size:    size,
		overlap: overlap,
	}, nil
}

func (c *Chunker) Split(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	runes := []rune(text)
	step := c.size - c.overlap

	chunks := make([]string, 0)

	for start := 0; start < len(runes); start += step {
		end := start + c.size
		if end > len(runes) {
			end = len(runes)
		}

		content := strings.TrimSpace(string(runes[start:end]))
		if content != "" {
			chunks = append(chunks, content)
		}

		if end == len(runes) {
			break
		}
	}
	return chunks
}
