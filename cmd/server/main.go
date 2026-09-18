package main

import (
	"context"
	"net/http"
	"os"

	"github.com/asahiiro/anchor-rag/internal/application"
	"github.com/asahiiro/anchor-rag/internal/chunker"
	"github.com/asahiiro/anchor-rag/internal/embedding"
	"github.com/asahiiro/anchor-rag/internal/generation"
	"github.com/asahiiro/anchor-rag/internal/handler"
	"github.com/asahiiro/anchor-rag/internal/storage"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	dataStore, err := storage.New(
		context.Background(),
		storage.Config{
			Provider:    os.Getenv("STORAGE_PROVIDER"),
			DatabaseURL: os.Getenv("DATABASE_URL"),
		},
	)
	if err != nil {
		panic(err)
	}
	defer dataStore.Close()

	textChunker, err := chunker.New(500, 50)
	if err != nil {
		panic(err)
	}
	textEmbedder, err := embedding.NewFromConfig(
		embedding.Config{
			Provider:       os.Getenv("EMBEDDING_PROVIDER"),
			BaseURL:        os.Getenv("EMBEDDING_BASE_URL"),
			APIKey:         os.Getenv("EMBEDDING_API_KEY"),
			Model:          os.Getenv("EMBEDDING_MODEL"),
			LocalDimension: 256,
		},
	)
	if err != nil {
		panic(err)
	}
	textGenerator, err := generation.NewFromConfig(
		generation.Config{
			Provider: os.Getenv("GENERATION_PROVIDER"),
			BaseURL:  os.Getenv("GENERATION_BASE_URL"),
			APIKey:   os.Getenv("GENERATION_API_KEY"),
			Model:    os.Getenv("GENERATION_MODEL"),
		},
	)
	if err != nil {
		panic(err)
	}

	documentService := application.NewDocumentService(
		dataStore.KnowledgeWriter,
		textChunker,
		textEmbedder,
	)
	searchService := application.NewSearchService(
		dataStore.ChunkSearcher,
		textEmbedder,
	)
	answerService := application.NewAnswerService(
		searchService,
		textGenerator,
	)

	documentHandler := handler.NewDocumentHandler(documentService)
	searchHandler := handler.NewSearchHandler(searchService)
	answerHandler := handler.NewAnswerHandler(answerService)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
	api := router.Group("/api/v1")
	api.POST("/documents", documentHandler.Create)
	api.POST("/retrievals", searchHandler.Search)
	api.POST("/answers", answerHandler.Create)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}

}
