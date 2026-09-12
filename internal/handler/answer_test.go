package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/asahiiro/anchor-rag/internal/application"
	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/gin-gonic/gin"
)

type stubAnswerService struct {
	result domain.AnswerResult
	err    error

	receivedQuestion string
	receivedLimit    int
}

func (s *stubAnswerService) Answer(
	_ context.Context,
	question string,
	limit int,
) (domain.AnswerResult, error) {
	s.receivedQuestion = question
	s.receivedLimit = limit

	return s.result, s.err
}

func TestAnswerHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubAnswerService{
		result: domain.AnswerResult{
			Answer: "RAG combines retrieval with generation [1].",
			Sources: []domain.SearchResult{
				{
					Chunk: domain.Chunk{
						ID:         "chunk-1",
						DocumentID: "document-1",
						Content:    "RAG combines retrieval with generation.",
					},
					Score: 0.95,
				},
			},
		},
	}

	answerHandler := NewAnswerHandler(service)

	router := gin.New()
	router.POST(
		"/api/v1/answers",
		answerHandler.Create,
	)

	body := strings.NewReader(`{
		"question": "What is RAG?"
	}`)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/answers",
		body,
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"status = %d, want %d: %s",
			recorder.Code,
			http.StatusOK,
			recorder.Body.String(),
		)
	}

	if service.receivedQuestion != "What is RAG?" {
		t.Fatalf(
			"question = %q",
			service.receivedQuestion,
		)
	}

	if service.receivedLimit != defaultAnswerLimit {
		t.Fatalf(
			"Limit = %d, want %d",
			service.receivedLimit,
			defaultAnswerLimit,
		)
	}

	var response domain.AnswerResult
	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Answer != service.result.Answer {
		t.Fatalf(
			"answer = %q, want %q",
			response.Answer,
			service.result.Answer,
		)
	}

	if len(response.Sources) != 1 {
		t.Fatalf(
			"source count = %d, want 1",
			len(response.Sources),
		)
	}
}

func TestAnswerHandlerReturnsNotFoundWithoutKnowledge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubAnswerService{
		err: application.ErrNoRelevantKnowledge,
	}

	answerHandler := NewAnswerHandler(service)

	router := gin.New()
	router.POST(
		"/api/v1/answers",
		answerHandler.Create,
	)

	body := strings.NewReader(`{
		"question": "What is RAG?"
	}`)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/answers",
		body,
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"status = %d, want %d: %s",
			recorder.Code,
			http.StatusNotFound,
			recorder.Body.String(),
		)
	}
}
