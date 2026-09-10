package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/generation"
)

var (
	ErrEmptyQuestion = errors.New(
		"question must not be empty",
	)
	ErrNoRelevantKnowledge = errors.New(
		"no relevant knowledge found",
	)
	ErrEmptyGeneratedAnswer = errors.New(
		"generator returned an empty answer",
	)
)

type AnswerService struct {
	searchService *SearchService
	generator     generation.Generator
}

func NewAnswerService(
	searchService *SearchService,
	textGenerator generation.Generator,
) *AnswerService {
	return &AnswerService{
		searchService: searchService,
		generator:     textGenerator,
	}
}

func (s *AnswerService) Answer(
	ctx context.Context,
	question string,
	limit int,
) (domain.AnswerResult, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return domain.AnswerResult{}, ErrEmptyQuestion
	}

	results, err := s.searchService.Search(
		ctx,
		question,
		limit,
	)
	if err != nil {
		return domain.AnswerResult{}, fmt.Errorf(
			"retrieve knowledge: %w",
			err,
		)
	}

	if len(results) == 0 {
		return domain.AnswerResult{}, ErrNoRelevantKnowledge
	}

	messages := buildAnswerMessages(question, results)

	answer, err := s.generator.Generate(
		ctx,
		messages,
	)
	if err != nil {
		return domain.AnswerResult{}, fmt.Errorf(
			"generate answer: %w",
			err,
		)
	}

	answer = strings.TrimSpace(answer)
	if answer == "" {
		return domain.AnswerResult{}, ErrEmptyGeneratedAnswer
	}

	return domain.AnswerResult{
		Answer:  answer,
		Sources: results,
	}, nil

}

func buildAnswerMessages(
	question string,
	results []domain.SearchResult,
) []generation.Message {

	var contextBuilder strings.Builder

	for index, result := range results {
		fmt.Fprintf(
			&contextBuilder,
			"[%d] document_id=%s chunk_id=%s\n%s\n\n",
			index+1,
			result.Chunk.DocumentID,
			result.Chunk.ID,
			result.Chunk.Content,
		)
	}

	return []generation.Message{
		{
			Role: generation.RoleSystem,
			Content: "Answer the question using only the supplied context. " +
				"If the context is insufficient, say that you do not know. " +
				"Cite supporting passages using [1], [2], and so on.",
		},
		{
			Role: generation.RoleUser,
			Content: fmt.Sprintf(
				"Question:\n%s\n\nContext:\n%s",
				question,
				contextBuilder.String(),
			),
		},
	}
}
