package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newDocumentUploadRequest(
	t *testing.T,
	filename string,
	content []byte,
) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile(
		"file",
		filename,
	)
	if err != nil {
		t.Fatalf("CreateFormFile() returned error: %v", err)
	}

	if _, err := part.Write(content); err != nil {
		t.Fatalf("write multipart content: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/documents/upload",
		&body,
	)
	request.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
	)

	return request
}

func TestDocumentHandlerUpload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, documentRepo, chunkRepo :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.POST(
		"/api/v1/documents/upload",
		documentHandler.Upload,
	)

	request := newDocumentUploadRequest(
		t,
		"rag-notes.md",
		[]byte("# RAG\n\nRetrieval augmented generation."),
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"status = %d, want %d: %s",
			recorder.Code,
			http.StatusCreated,
			recorder.Body.String(),
		)
	}

	var response struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Name != "rag-notes.md" {
		t.Fatalf(
			"name = %q, want rag-notes.md",
			response.Name,
		)
	}

	storedDocument, err := documentRepo.FindByID(
		context.Background(),
		response.ID,
	)
	if err != nil {
		t.Fatalf("document was not saved: %v", err)
	}

	if storedDocument.Content !=
		"# RAG\n\nRetrieval augmented generation." {
		t.Fatalf(
			"stored content = %q",
			storedDocument.Content,
		)
	}

	storedChunks, err := chunkRepo.FindByDocumentID(
		context.Background(),
		response.ID,
	)
	if err != nil {
		t.Fatalf("find chunks: %v", err)
	}

	if len(storedChunks) == 0 {
		t.Fatal("expected uploaded document chunks")
	}
}

func TestDocumentHandlerUploadAcceptsTextFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, _, _ :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.POST(
		"/api/v1/documents/upload",
		documentHandler.Upload,
	)

	request := newDocumentUploadRequest(
		t,
		"notes.TXT",
		[]byte("plain text document"),
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"status = %d, want %d: %s",
			recorder.Code,
			http.StatusCreated,
			recorder.Body.String(),
		)
	}
}

func TestDocumentHandlerUploadRejectsMissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, _, _ :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.POST(
		"/api/v1/documents/upload",
		documentHandler.Upload,
	)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/documents/upload",
		&body,
	)
	request.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
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

	wantBody := `{"error":"file is required"}`
	if recorder.Body.String() != wantBody {
		t.Fatalf(
			"body = %q, want %q",
			recorder.Body.String(),
			wantBody,
		)
	}

}

func TestDocumentHandlerUploadRejectsUnsupportedType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, _, _ :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.POST(
		"/api/v1/documents/upload",
		documentHandler.Upload,
	)

	request := newDocumentUploadRequest(
		t,
		"rag-notes.pdf",
		[]byte("# RAG\n\nRetrieval augmented generation."),
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

	wantBody := `{"error":"only .txt and .md files are supported"}`
	if recorder.Body.String() != wantBody {
		t.Fatalf(
			"body = %q, want %q",
			recorder.Body.String(),
			wantBody,
		)
	}
}

func TestDocumentHandlerUploadRejectsEmptyFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, _, _ :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.POST(
		"/api/v1/documents/upload",
		documentHandler.Upload,
	)

	request := newDocumentUploadRequest(
		t,
		"rag-notes.md",
		[]byte("      "),
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

	wantBody := `{"error":"file must not be empty"}`
	if recorder.Body.String() != wantBody {
		t.Fatalf(
			"body = %q, want %q",
			recorder.Body.String(),
			wantBody,
		)
	}
}

func TestDocumentHandlerUploadRejectsInvalidUTF8(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, _, _ :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.POST(
		"/api/v1/documents/upload",
		documentHandler.Upload,
	)

	request := newDocumentUploadRequest(
		t,
		"rag-notes.md",
		[]byte{0xff, 0xfe},
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

	wantBody := `{"error":"file must contain valid UTF-8 text"}`
	if recorder.Body.String() != wantBody {
		t.Fatalf(
			"body = %q, want %q",
			recorder.Body.String(),
			wantBody,
		)
	}
}

func TestDocumentHandlerUploadRejectsLargeFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	documentHandler, _, _ :=
		newTestDocumentHandler(t)

	router := gin.New()
	router.POST(
		"/api/v1/documents/upload",
		documentHandler.Upload,
	)

	content := bytes.Repeat(
		[]byte("a"),
		int(maxDocumentFileSize)+1,
	)
	request := newDocumentUploadRequest(
		t,
		"rag-notes.md",
		content,
	)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf(
			"status = %d, want %d: %s",
			recorder.Code,
			http.StatusRequestEntityTooLarge,
			recorder.Body.String(),
		)
	}

	wantBody := `{"error":"file must not exceed 2 MiB"}`
	if recorder.Body.String() != wantBody {
		t.Fatalf(
			"body = %q, want %q",
			recorder.Body.String(),
			wantBody,
		)
	}
}
