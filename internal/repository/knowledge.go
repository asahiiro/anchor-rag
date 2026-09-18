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
