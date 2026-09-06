package rag

import "testing"

func TestRetrieverRanksRelevantJapaneseChunkFirst(t *testing.T) {
	chunks := []Chunk{
		{Source: "go.md", Text: "Goは静的型付けのプログラミング言語です。"},
		{Source: "rag.md", Text: "RAGは関連文書を検索して回答生成に使います。"},
	}
	retriever := NewRetriever(chunks)
	results := retriever.Search("RAGで文書を検索する仕組みは？", 1)
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Chunk.Source != "rag.md" {
		t.Fatalf("got source %q, want rag.md", results[0].Chunk.Source)
	}
}

func TestRetrieverReturnsNoResultForUnknownTerms(t *testing.T) {
	retriever := NewRetriever([]Chunk{{Source: "go.md", Text: "Go language"}})
	if results := retriever.Search("zebra", 3); len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}
}
