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
	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/asahiiro/anchor-rag/internal/embedding"
	"github.com/asahiiro/anchor-rag/internal/repository/memory"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	writer := memory.NewKnowledgeWriter(
		documentRepo,
		chunkRepo,
	)
	deleter := memory.NewKnowledgeDeleter(
		documentRepo,
		chunkRepo,
	)

	textChunker, err := chunker.New(5, 2)
	if err != nil {
		t.Fatalf("chunker.New() returned error: %v", err)
	}

	textEmbedder, err := embedding.NewRuneFrequencyEmbedder(32)
	if err != nil {
		t.Fatalf("embedding constructor returned error: %v", err)
	}

	service := application.NewDocumentService(
		writer,
		documentRepo,
		deleter,
		textChunker,
		textEmbedder,
	)

	return NewDocumentHandler(service), documentRepo, chunkRepo
}

func TestDocumentHandlerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, documentRepo, chunkRepo :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.POST("/api/v1/documents", documentHandler.Create)

	body := strings.NewReader(`{
		"name": "rag-notes.md",
		"content": "RAG combines retrieval with generation."
		}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/documents",
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
	router.POST("/api/v1/documents", documentHandler.Create)

	body := strings.NewReader(`{
		"name": "rag-notes.md"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/documents",
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
	router.POST("/api/v1/documents", documentHandler.Create)

	body := strings.NewReader(`{
		"name": "   ",
		"content": "valid content"
	}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/documents",
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

func TestDocumentHandlerGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, documentRepo, _ :=
		newTestDocumentHandler(t)

	doc, err := domain.NewDocument(
		uuid.NewString(),
		"retrievable.md",
		"retrievable document content",
	)
	if err != nil {
		t.Fatalf("NewDocument() returned error: %v", err)
	}

	if err := documentRepo.Save(
		context.Background(),
		doc,
	); err != nil {
		t.Fatalf("save document: %v", err)
	}

	router := gin.New()
	router.GET(
		"/api/v1/documents/:id",
		documentHandler.Get,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/documents/"+doc.ID,
		nil,
	)

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

	var response struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Content string `json:"content"`
	}

	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.ID != doc.ID {
		t.Fatalf(
			"document ID = %q, want %q",
			response.ID,
			doc.ID,
		)
	}

	if response.Name != doc.Name {
		t.Fatalf(
			"document name = %q, want %q",
			response.Name,
			doc.Name,
		)
	}

	if response.Content != doc.Content {
		t.Fatalf(
			"document content = %q, want %q",
			response.Content,
			doc.Content,
		)
	}
}

func TestDocumentHandlerGetRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, _, _ :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.GET(
		"/api/v1/documents/:id",
		documentHandler.Get,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/documents/not-a-uuid",
		nil,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d: %s",
			recorder.Code,
			http.StatusBadRequest,
			recorder.Body.String(),
		)
	}

	wantBody := `{"error":"invalid document ID"}`
	if recorder.Body.String() != wantBody {
		t.Fatalf(
			"body = %q, want %q",
			recorder.Body.String(),
			wantBody,
		)
	}
}

func TestDocumentHandlerGetReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, _, _ :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.GET(
		"/api/v1/documents/:id",
		documentHandler.Get,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/documents/"+uuid.NewString(),
		nil,
	)

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

	wantBody := `{"error":"document not found"}`
	if recorder.Body.String() != wantBody {
		t.Fatalf(
			"body = %q, want %q",
			recorder.Body.String(),
			wantBody,
		)
	}
}

func TestDocumentHandlerDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, documentRepo, _ :=
		newTestDocumentHandler(t)

	doc, err := domain.NewDocument(
		uuid.NewString(),
		"retrievable.md",
		"retrievable document content",
	)
	if err != nil {
		t.Fatalf("NewDocument() returned error: %v", err)
	}

	if err := documentRepo.Save(
		context.Background(),
		doc,
	); err != nil {
		t.Fatalf("save document: %v", err)
	}

	router := gin.New()
	router.DELETE(
		"/api/v1/documents/:id",
		documentHandler.Delete,
	)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/documents/"+doc.ID,
		nil,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"status = %d, want %d: %s",
			recorder.Code,
			http.StatusNoContent,
			recorder.Body.String(),
		)
	}

	if len(recorder.Body.Bytes()) != 0 {
		t.Fatalf(
			"body length = %d, want 0",
			len(recorder.Body.Bytes()),
		)
	}

	router.GET(
		"/api/v1/documents/:id",
		documentHandler.Get,
	)

	request = httptest.NewRequest(
		http.MethodGet,
		"/api/v1/documents/"+doc.ID,
		nil,
	)

	recorder = httptest.NewRecorder()
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

func TestDocumentHandlerDeleteRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, _, _ :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.DELETE(
		"/api/v1/documents/:id",
		documentHandler.Delete,
	)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/documents/not-a-uuid",
		nil,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"status = %d, want %d: %s",
			recorder.Code,
			http.StatusBadRequest,
			recorder.Body.String(),
		)
	}

	wantBody := `{"error":"invalid document ID"}`
	if recorder.Body.String() != wantBody {
		t.Fatalf(
			"body = %q, want %q",
			recorder.Body.String(),
			wantBody,
		)
	}
}

func TestDocumentHandlerDeleteReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, _, _ :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.DELETE(
		"/api/v1/documents/:id",
		documentHandler.Delete,
	)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/api/v1/documents/"+uuid.NewString(),
		nil,
	)

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

	wantBody := `{"error":"document not found"}`
	if recorder.Body.String() != wantBody {
		t.Fatalf(
			"body = %q, want %q",
			recorder.Body.String(),
			wantBody,
		)
	}
}
