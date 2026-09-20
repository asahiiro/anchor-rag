package memory

import (
	"context"

	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/repository"
)

type KnowledgeWriter struct {
	documentRepo *DocumentRepository
	chunkRepo    *ChunkRepository
}

type KnowledgeDeleter struct {
	documentRepo *DocumentRepository
	chunkRepo    *ChunkRepository
}

var _ repository.KnowledgeWriter = (*KnowledgeWriter)(nil)
var _ repository.KnowledgeDeleter = (*KnowledgeDeleter)(nil)

func NewKnowledgeWriter(
	documentRepo *DocumentRepository,
	chunkRepo *ChunkRepository,
) *KnowledgeWriter {
	return &KnowledgeWriter{
		documentRepo: documentRepo,
		chunkRepo:    chunkRepo,
	}
}

func NewKnowledgeDeleter(
	documentRepo *DocumentRepository,
	chunkRepo *ChunkRepository,
) *KnowledgeDeleter {
	return &KnowledgeDeleter{
		documentRepo: documentRepo,
		chunkRepo:    chunkRepo,
	}
}

func (w *KnowledgeWriter) SaveDocument(
	ctx context.Context,
	doc domain.Document,
	chunks []domain.Chunk,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	w.documentRepo.mu.Lock()
	w.chunkRepo.mu.Lock()

	defer w.chunkRepo.mu.Unlock()
	defer w.documentRepo.mu.Unlock()

	w.documentRepo.documents[doc.ID] = doc

	for _, chunk := range chunks {
		w.chunkRepo.chunksByDocument[chunk.DocumentID] = append(
			w.chunkRepo.chunksByDocument[chunk.DocumentID],
			chunk,
		)
	}

	return nil
}

func (d *KnowledgeDeleter) DeleteDocument(
	ctx context.Context,
	id string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	d.documentRepo.mu.Lock()
	d.chunkRepo.mu.Lock()

	defer d.chunkRepo.mu.Unlock()
	defer d.documentRepo.mu.Unlock()

	if _, ok := d.documentRepo.documents[id]; !ok {
		return repository.ErrDocumentNotFound
	}

	delete(d.documentRepo.documents, id)
	delete(d.chunkRepo.chunksByDocument, id)

	return nil
}
