package memory

import (
	"context"
	"sync"

	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/repository"
)

type DocumentRepository struct {
	mu        sync.RWMutex
	documents map[string]domain.Document
}

var _ repository.DocumentRepository = (*DocumentRepository)(nil)

func NewDocumentRepository() *DocumentRepository {
	return &DocumentRepository{
		documents: make(map[string]domain.Document),
	}
}

func (r *DocumentRepository) Save(
	ctx context.Context,
	doc domain.Document,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.documents[doc.ID] = doc
	return nil
}

func (r *DocumentRepository) FindByID(
	ctx context.Context,
	id string,
) (domain.Document, error) {
	if err := ctx.Err(); err != nil {
		return domain.Document{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	doc, ok := r.documents[id]
	if !ok {
		return domain.Document{}, repository.ErrDocumentNotFound
	}

	return doc, nil
}
