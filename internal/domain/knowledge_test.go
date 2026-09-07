package domain

import "testing"

func TestNewDocument(t *testing.T) {
	doc, err := NewDocument(
		"doc-1",
		"  internship-notes.md  ",
		"RAG consists of retrieval and generation.",
	)
	if err != nil {
		t.Fatalf("NewDocument() returned error: %v", err)
	}

	if doc.Name != "internship-notes.md" {
		t.Fatalf("expected trimmed name,got %q", doc.Name)
	}

	if doc.Content == "" {
		t.Fatal("expected document content to be preserved")
	}

	if doc.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be initialized")
	}
}

func TestNewDocumentRejectsBlankName(t *testing.T) {
	_, err := NewDocument("doc-2", "   ", "some content")
	if err == nil {
		t.Fatalf("expected an error for blankk document name")
	}
}
