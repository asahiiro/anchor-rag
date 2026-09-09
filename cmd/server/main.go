package main

import (
	"net/http"

	"github.com/asahiiro/anchor-rag/internal/application"
	"github.com/asahiiro/anchor-rag/internal/chunker"
	"github.com/asahiiro/anchor-rag/internal/embedding"
	"github.com/asahiiro/anchor-rag/internal/handler"
	"github.com/asahiiro/anchor-rag/internal/repository/memory"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	documentRepo := memory.NewDocumentRepository()
	chunkRepo := memory.NewChunkRepository()
	textChunker, err := chunker.New(500, 50)
	if err != nil {
		panic(err)
	}
	textEmbedder, err := embedding.NewRuneFrequencyEmbedder(256)
	if err != nil {
		panic(err)
	}

	documentService := application.NewDocumentService(
		documentRepo,
		chunkRepo,
		textChunker,
		textEmbedder,
	)
	searchService := application.NewSearchService(
		chunkRepo,
		textEmbedder,
	)

	documentHandler := handler.NewDocumentHandler(documentService)
	searchHandler := handler.NewSearchHandler(searchService)

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
	api := router.Group("/api/v1")
	api.POST("/documents", documentHandler.Create)
	api.POST("/retrievals", searchHandler.Search)

	if err := router.Run(":8080"); err != nil {
		panic(err)
	}

}
