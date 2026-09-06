package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/taako-502/tensai-rag-example/internal/rag"
)

func main() {
	docsDir := flag.String("docs", "./docs", "knowledge-base directory")
	endpoint := flag.String("endpoint", envOr("TENSAI_ENDPOINT", "http://127.0.0.1:8080"), "tensai server URL")
	topK := flag.Int("top-k", 3, "number of chunks to retrieve")
	chunkSize := flag.Int("chunk-size", 600, "maximum runes per chunk")
	maxTokens := flag.Int("max-tokens", 256, "maximum tokens to generate")
	retrieveOnly := flag.Bool("retrieve-only", false, "print retrieved chunks without calling the LLM")
	flag.Parse()

	docs, err := rag.LoadDocuments(*docsDir)
	if err != nil {
		exit(err)
	}
	chunks := rag.SplitDocuments(docs, *chunkSize)
	retriever := rag.NewRetriever(chunks)
	fmt.Fprintf(os.Stderr, "indexed %d chunks from %d documents\n", len(chunks), len(docs))

	question := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if question == "" {
		fmt.Print("質問: ")
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil && len(line) == 0 {
			exit(fmt.Errorf("read question: %w", err))
		}
		question = strings.TrimSpace(line)
	}
	if question == "" {
		exit(fmt.Errorf("question must not be empty"))
	}

	results := retriever.Search(question, *topK)
	if len(results) == 0 {
		fmt.Println("関連する文書が見つかりませんでした。")
		return
	}
	if *retrieveOnly {
		for i, result := range results {
			fmt.Printf("[%d] %s (score: %.4f)\n%s\n\n", i+1, result.Chunk.Source, result.Score, result.Chunk.Text)
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	generator := rag.Generator{
		Endpoint:  *endpoint,
		APIKey:    os.Getenv("TENSAI_API_KEY"),
		MaxTokens: *maxTokens,
	}
	answer, err := generator.Generate(ctx, question, results)
	if err != nil {
		exit(err)
	}
	fmt.Println(answer)
	fmt.Println("\n検索結果:")
	for i, result := range results {
		fmt.Printf("[%d] %s (score: %.4f)\n", i+1, result.Chunk.Source, result.Score)
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
