package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/asahiiro/anchor-rag/internal/application"
	"github.com/asahiiro/anchor-rag/internal/chunker"
	"github.com/asahiiro/anchor-rag/internal/repository/memory"
	"github.com/gin-gonic/gin"
)

func newTestDocumentHandler(
	t *testing.T,
) (
	*DocumentHandler,
	*memory.DocumentRepository,
	*memory.ChunkRepository,
) {
	t.Helper()

	documentRepo := memory.NewDocumentRepository()
	chunkRepo := memory.NewChunkRepository()
	textChunker, err := chunker.New(5, 2)
	if err != nil {
		t.Fatalf("chunker.New() returned error: %v", err)
	}

	service := application.NewDocumentService(
		documentRepo,
		chunkRepo,
		textChunker,
	)

	return NewDocumentHandler(service), documentRepo, chunkRepo
}

func TestDocumentHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, documentRepo, chunkRepo :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.POST("/api/documents", documentHandler.Create)

	body := strings.NewReader(`{
		"name": "rag-notes.md",
		"content": "RAG combines retrieval with generation."
		}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/documents",
		body,
	)

	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d got %d: %s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v:", err)
	}

	if response.ID == "" {
		t.Fatalf("expected generated document ID")
	}

	if response.Name != "rag-notes.md" {
		t.Fatalf("unexpected document name: %q", response.Name)
	}

	stored, err := documentRepo.FindByID(context.Background(), response.ID)
	if err != nil {
		t.Fatalf("document was not saved: %v", err)
	}

	if stored.Content != "RAG combines retrieval with generation." {
		t.Fatalf("unexpected stored content: %q", stored.Content)
	}
	chunks, err := chunkRepo.FindByDocumentID(
		context.Background(),
		response.ID,
	)
	if err != nil {
		t.Fatalf("failed to find chunks: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected document chunks to be created")
	}
}

func TestDocumentHandlerCreateRejectsInvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, _, _ :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.POST("/api/documents", documentHandler.Create)

	// 缺少必填的 content 字段。
	body := strings.NewReader(`{
		"name": "rag-notes.md"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/documents",
		body,
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestDocumentHandlerCreateRejectsBlankName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, _, _ := newTestDocumentHandler(t)

	router := gin.New()
	router.POST("/api/documents", documentHandler.Create)

	body := strings.NewReader(`{
		"name": "   ",
		"content": "valid content"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/documents",
		body,
	)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}

	wantBody := `{"error":"invalid document: document name is required"}`
	if recorder.Body.String() != wantBody {
		t.Fatalf(
			"response body = %q, want %q",
			recorder.Body.String(),
			wantBody,
		)
	}
}
