package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/asahiiro/anchor-rag/internal/chunker"
	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/repository"
	"github.com/google/uuid"
)

var ErrInvalidDocument = errors.New("invalid document")

type DocumentService struct {
	documentRepo repository.DocumentRepository
	chunkRepo    repository.ChunkRepository
	chunker      *chunker.Chunker
}

func NewDocumentService(
	documentRepo repository.DocumentRepository,
	chunkRepo repository.ChunkRepository,
	textChunker *chunker.Chunker,
) *DocumentService {
	return &DocumentService{
		documentRepo: documentRepo,
		chunkRepo:    chunkRepo,
		chunker:      textChunker,
	}
}

func (s *DocumentService) Create(
	ctx context.Context,
	name string,
	content string,
) (domain.Document, error) {
	doc, err := domain.NewDocument(
		uuid.NewString(),
		name,
		content,
	)
	if err != nil {
		return domain.Document{}, fmt.Errorf(
			"%w: %v",
			ErrInvalidDocument,
			err,
		)
	}

	parts := s.chunker.Split(doc.Content)

	chunks := make([]domain.Chunk, 0, len(parts))
	for position, chunkContent := range parts {
		chunks = append(chunks, domain.Chunk{
			ID:         uuid.NewString(),
			DocumentID: doc.ID,
			Content:    chunkContent,
			Position:   position,
		})
	}

	if err := s.documentRepo.Save(ctx, doc); err != nil {
		return domain.Document{}, fmt.Errorf(
			"save document: %w",
			err,
		)
	}

	if err := s.chunkRepo.SaveBatch(ctx, chunks); err != nil {
		return domain.Document{}, fmt.Errorf(
			"save chunks: %w",
			err,
		)
	}

	return doc, nil
}
