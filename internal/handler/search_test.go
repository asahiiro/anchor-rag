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
)

func TestSearchHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentRepo := memory.NewDocumentRepository()
	chunkRepo := memory.NewChunkRepository()

	textChunker, err := chunker.New(100, 10)
	if err != nil {
		t.Fatalf("chunker.New() returned error: %v", err)
	}

	textEmbedder, err := embedding.NewRuneFrequencyEmbedder(256)
	if err != nil {
		t.Fatalf("embedding constructor returned error: %v", err)
	}

	documentService := application.NewDocumentService(
		documentRepo,
		chunkRepo,
		textChunker,
		textEmbedder,
	)

	if _, err := documentService.Create(
		context.Background(),
		"go.md",
		"golang",
	); err != nil {
		t.Fatalf("failed to create Go document: %v", err)
	}

	if _, err := documentService.Create(
		context.Background(),
		"unrelated.md",
		"zzzzzz",
	); err != nil {
		t.Fatalf("failed to create unrelated document: %v", err)
	}

	searchService := application.NewSearchService(
		chunkRepo,
		textEmbedder,
	)
	searchHandler := NewSearchHandler(searchService)

	router := gin.New()
	router.POST("/api/v1/retrievals", searchHandler.Search)

	body := strings.NewReader(`{
		"query": "golang",
		"limit": 1
		}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/retrievals", body)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response struct {
		Results []domain.SearchResult `json:"results"`
	}

	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Results) != 1 {
		t.Fatalf(
			"got %d results, want 1",
			len(response.Results),
		)
	}

	if response.Results[0].Chunk.Content != "golang" {
		t.Fatalf(
			"first result content = %q, want golang",
			response.Results[0].Chunk.Content,
		)
	}
}
