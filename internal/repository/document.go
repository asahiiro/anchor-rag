package repository

import (
	"context"
	"errors"

	"github.com/asahiiro/anchor-rag/internal/domain"
)

var ErrDocumentNotFound = errors.New("document not found")

type DocumentRepository interface {
	Save(ctx context.Context, doc domain.Document) error
	FindByID(ctx context.Context, id string) (domain.Document, error)
}

type DocumentReader interface {
	FindByID(
		ctx context.Context,
		id string,
	) (domain.Document, error)
}
