package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Generator struct {
	Endpoint  string
	APIKey    string
	MaxTokens int
	Client    *http.Client
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

// Generateは、検索結果をもとにtensaiサーバーで回答を生成します。
func (g Generator) Generate(ctx context.Context, question string, results []Result) (string, error) {
	var contextText strings.Builder
	for i, result := range results {
		fmt.Fprintf(&contextText, "[%d] source: %s\n%s\n\n", i+1, result.Chunk.Source, result.Chunk.Text)
	}
	requestBody := chatRequest{
		Model: "tensai",
		Messages: []chatMessage{
			{Role: "system", Content: "あなたは検索結果だけを根拠に回答するアシスタントです。根拠がない場合は、情報が見つからないと明記してください。回答末尾に参照した番号を [1] の形式で示してください。"},
			{Role: "user", Content: "検索結果:\n\n" + contextText.String() + "質問:\n" + question},
		},
		MaxTokens: g.MaxTokens,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("encode request: %w", err)
	}

	endpoint := strings.TrimRight(g.Endpoint, "/") + "/v1/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if g.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+g.APIKey)
	}
	client := g.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call tensai server: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("tensai server returned %s: %s", resp.Status, strings.TrimSpace(string(message)))
	}
	var decoded chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("tensai server returned no choices")
	}
	return strings.TrimSpace(decoded.Choices[0].Message.Content), nil
}
