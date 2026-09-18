package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/asahiiro/anchor-rag/internal/chunker"
	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/embedding"
	"github.com/asahiiro/anchor-rag/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrInvalidDocument = errors.New("invalid document")

	ErrEmbeddingCountMismatch = errors.New(
		"embedding count does not match chunk count",
	)
)

type DocumentService struct {
	writer   repository.KnowledgeWriter
	chunker  *chunker.Chunker
	embedder embedding.Embedder
}

func NewDocumentService(
	writer repository.KnowledgeWriter,
	textChunker *chunker.Chunker,
	textEmbedder embedding.Embedder,
) *DocumentService {
	return &DocumentService{
		writer:   writer,
		chunker:  textChunker,
		embedder: textEmbedder,
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

	embeddings, err := s.embedder.Embed(
		ctx,
		parts,
	)
	if err != nil {
		return domain.Document{}, fmt.Errorf(
			"embed chunks: %w",
			err,
		)
	}
	if len(embeddings) != len(parts) {
		return domain.Document{}, fmt.Errorf(
			"%w: got %d, want %d",
			ErrEmbeddingCountMismatch,
			len(embeddings),
			len(parts),
		)
	}

	chunks := make([]domain.Chunk, 0, len(parts))
	for position, chunkContent := range parts {
		chunks = append(chunks, domain.Chunk{
			ID:         uuid.NewString(),
			DocumentID: doc.ID,
			Content:    chunkContent,
			Position:   position,
			Embedding:  embeddings[position],
		})
	}

	if err := s.writer.SaveDocument(
		ctx,
		doc,
		chunks,
	); err != nil {
		return domain.Document{}, fmt.Errorf(
			"save document with chunks: %w",
			err,
		)
	}

	return doc, nil
}
