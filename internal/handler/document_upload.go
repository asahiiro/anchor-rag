package handler

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

const (
	maxDocumentFileSize   = 2 << 20
	maxDocumentUploadSize = maxDocumentFileSize + 64<<10
)

func (h *DocumentHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(
		c.Writer,
		c.Request.Body,
		maxDocumentUploadSize,
	)

	fileHeader, err := c.FormFile("file")
	if err != nil {
		var maxBytesError *http.MaxBytesError

		if errors.As(err, &maxBytesError) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "file must not exceed 2 MiB",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file is required",
		})
		return
	}

	if fileHeader.Size > maxDocumentFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "file must not exceed 2 MiB",
		})
		return
	}

	name := filepath.Base(fileHeader.Filename)
	extension := strings.ToLower(filepath.Ext(name))

	switch extension {
	case ".md", ".txt":
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "only .txt and .md files are supported",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read document",
		})
		return
	}
	defer file.Close()

	content, err := io.ReadAll(
		io.LimitReader(
			file,
			maxDocumentFileSize+1,
		),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to read document",
		})
		return
	}

	if int64(len(content)) > maxDocumentFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "file must not exceed 2 MiB",
		})
		return
	}

	if !utf8.Valid(content) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file must contain valid UTF-8 text",
		})
		return
	}

	if strings.TrimSpace(string(content)) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file must not be empty",
		})
		return
	}

	h.createDocument(
		c,
		name,
		string(content),
	)
}
