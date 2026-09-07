package domain

import (
	"errors"
	"strings"
	"time"
)

type Document struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
}

type Chunk struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	Content    string    `json:"content"`
	Position   int       `json:"position"`
	Embedding  []float32 `json:"-"`
}

func NewDocument(id, name, content string) (Document, error) {
	normalizedName := strings.TrimSpace(name)

	if normalizedName == "" {
		return Document{}, errors.New("document name is required")
	}

	return Document{
		ID:        id,
		Name:      normalizedName,
		Content:   content,
		CreatedAt: time.Now(),
	}, nil
}
