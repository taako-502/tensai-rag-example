package rag

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAndSplitDocuments(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("first paragraph\nsecond paragraph"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ignored.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	docs, err := LoadDocuments(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("got %d documents, want 1", len(docs))
	}
	chunks := SplitDocuments(docs, 20)
	if len(chunks) != 2 {
		t.Fatalf("got %d chunks, want 2", len(chunks))
	}
}
