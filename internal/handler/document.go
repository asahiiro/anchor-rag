package handler

import (
	"errors"
	"net/http"
	"time"

	"github.com/asahiiro/anchor-rag/internal/application"
	"github.com/asahiiro/anchor-rag/internal/repository"
	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	service *application.DocumentService
}

func NewDocumentHandler(
	service *application.DocumentService,
) *DocumentHandler {
	return &DocumentHandler{
		service: service,
	}
}

type createDocumentRequest struct {
	Name    string `json:"name" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type documentDetailResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *DocumentHandler) Create(c *gin.Context) {
	var req createDocumentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name and content are required",
		})
		return
	}

	doc, err := h.service.Create(
		c.Request.Context(),
		req.Name,
		req.Content,
	)

	if errors.Is(err, application.ErrInvalidDocument) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create document",
		})
		return
	}

	c.JSON(http.StatusCreated, doc)

}

func (h *DocumentHandler) Get(c *gin.Context) {
	doc, err := h.service.Get(
		c.Request.Context(),
		c.Param("id"),
	)

	switch {
	case errors.Is(err, application.ErrInvalidDocumentID):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": application.ErrInvalidDocumentID.Error(),
		})
		return

	case errors.Is(err, repository.ErrDocumentNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": repository.ErrDocumentNotFound.Error(),
		})
		return

	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get document",
		})
		return
	}

	c.JSON(http.StatusOK, documentDetailResponse{
		ID:        doc.ID,
		Name:      doc.Name,
		Content:   doc.Content,
		CreatedAt: doc.CreatedAt,
	})
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	err := h.service.Delete(
		c.Request.Context(),
		c.Param("id"),
	)

	switch {
	case errors.Is(err, application.ErrInvalidDocumentID):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": application.ErrInvalidDocumentID.Error(),
		})
		return

	case errors.Is(err, repository.ErrDocumentNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": repository.ErrDocumentNotFound.Error(),
		})
		return

	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete document",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
