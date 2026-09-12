package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/asahiiro/anchor-rag/internal/application"
	"github.com/asahiiro/anchor-rag/internal/domain"
	"github.com/gin-gonic/gin"
)

const defaultAnswerLimit = 3

type answerService interface {
	Answer(
		ctx context.Context,
		question string,
		limit int,
	) (domain.AnswerResult, error)
}

type AnswerHandler struct {
	service answerService
}

func NewAnswerHandler(service answerService) *AnswerHandler {
	return &AnswerHandler{
		service: service,
	}
}

type answerRequest struct {
	Question string `json:"question" binding:"required"`
	Limit    int    `json:"limit" binding:"omitempty,gte=1,lte=20"`
}

func (h *AnswerHandler) Create(c *gin.Context) {
	var req answerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "question is required and limit must be between 1 and 20",
		})
		return
	}

	if req.Limit == 0 {
		req.Limit = defaultAnswerLimit
	}

	result, err := h.service.Answer(
		c.Request.Context(),
		req.Question,
		req.Limit,
	)

	switch {
	case errors.Is(err, application.ErrEmptyQuestion),
		errors.Is(err, application.ErrInvalidSearchLimit):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return

	case errors.Is(err, application.ErrNoRelevantKnowledge):
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return

	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate answer",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
