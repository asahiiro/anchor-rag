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

var _ repository.KnowledgeWriter = (*KnowledgeWriter)(nil)

func NewKnowledgeWriter(
	documentRepo *DocumentRepository,
	chunkRepo *ChunkRepository,
) *KnowledgeWriter {
	return &KnowledgeWriter{
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
