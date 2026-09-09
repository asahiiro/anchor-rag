package handler

import (
	"errors"
	"net/http"

	"github.com/asahiiro/anchor-rag/internal/application"
	"github.com/gin-gonic/gin"
)

const defaultSearchLimit = 3

type SearchHandler struct {
	service *application.SearchService
}

func NewSearchHandler(
	service *application.SearchService,
) *SearchHandler {
	return &SearchHandler{
		service: service,
	}
}

type searchRequest struct {
	Query string `json:"query" binding:"required"`
	Limit int    `json:"limit" binding:"omitempty,gte=1,lte=20"`
}

func (h *SearchHandler) Search(c *gin.Context) {
	var req searchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "query is required and limit must be between 1 and 20",
		})
		return
	}

	if req.Limit == 0 {
		req.Limit = defaultSearchLimit
	}

	results, err := h.service.Search(
		c.Request.Context(),
		req.Query,
		req.Limit,
	)

	if errors.Is(err, application.ErrEmptySearchQuery) ||
		errors.Is(err, application.ErrInvalidSearchLimit) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to search knowledge",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}
