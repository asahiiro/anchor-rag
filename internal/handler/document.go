package handler

import (
	"errors"
	"net/http"

	"github.com/asahiiro/anchor-rag/internal/application"
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
