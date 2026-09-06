package rag

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGeneratorCallsOpenAICompatibleEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("got path %q", r.URL.Path)
		}
		var request chatRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(request.Messages[1].Content, "rag.md") {
			t.Error("request does not contain retrieved source")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"回答 [1]"}}]}`))
	}))
	defer server.Close()

	generator := Generator{Endpoint: server.URL, MaxTokens: 64}
	answer, err := generator.Generate(context.Background(), "質問", []Result{{Chunk: Chunk{Source: "rag.md", Text: "根拠"}}})
	if err != nil {
		t.Fatal(err)
	}
	if answer != "回答 [1]" {
		t.Fatalf("got %q", answer)
	}
}
