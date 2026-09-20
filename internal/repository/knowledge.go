package repository

import (
	"context"

	"github.com/asahiiro/anchor-rag/internal/domain"
)

type KnowledgeWriter interface {
	SaveDocument(
		ctx context.Context,
		doc domain.Document,
		chunks []domain.Chunk,
	) error
}

type KnowledgeDeleter interface {
	DeleteDocument(
		ctx context.Context,
		id string,
	) error
}
