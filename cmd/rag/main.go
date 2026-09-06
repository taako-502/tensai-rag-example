package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/taako-502/tensai-rag-example/internal/rag"
)

func main() {
	if err := loadDotEnv(".env"); err != nil {
		exit(err)
	}
	cfg := parseConfig()
	if err := run(cfg, flag.Args()); err != nil {
		exit(err)
	}
}

type config struct {
	docsDir      string
	endpoint     string
	topK         int
	chunkSize    int
	maxTokens    int
	retrieveOnly bool
}

func parseConfig() config {
	docsDir := flag.String("docs", "./docs", "knowledge-base directory")
	endpoint := flag.String("endpoint", envOr("TENSAI_ENDPOINT", "http://127.0.0.1:8080"), "tensai server URL")
	topK := flag.Int("top-k", 3, "number of chunks to retrieve")
	chunkSize := flag.Int("chunk-size", 600, "maximum runes per chunk")
	maxTokens := flag.Int("max-tokens", 256, "maximum tokens to generate")
	retrieveOnly := flag.Bool("retrieve-only", false, "print retrieved chunks without calling the LLM")
	flag.Parse()
	return config{
		docsDir:      *docsDir,
		endpoint:     *endpoint,
		topK:         *topK,
		chunkSize:    *chunkSize,
		maxTokens:    *maxTokens,
		retrieveOnly: *retrieveOnly,
	}
}

func run(cfg config, questionArgs []string) error {
	retriever, documentCount, chunkCount, err := prepareRetriever(cfg)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "indexed %d chunks from %d documents\n", chunkCount, documentCount)

	question, err := readQuestion(questionArgs)
	if err != nil {
		return err
	}

	results := retriever.Search(question, cfg.topK)
	if len(results) == 0 {
		fmt.Println("関連する文書が見つかりませんでした。")
		return nil
	}
	if cfg.retrieveOnly {
		printRetrievedChunks(results)
		return nil
	}

	return generateAnswer(cfg, question, results)
}

func prepareRetriever(cfg config) (*rag.Retriever, int, int, error) {
	docs, err := rag.LoadDocuments(cfg.docsDir)
	if err != nil {
		return nil, 0, 0, err
	}
	chunks := rag.SplitDocuments(docs, cfg.chunkSize)
	return rag.NewRetriever(chunks), len(docs), len(chunks), nil
}

// readQuestionは、コマンドライン引数または標準入力から質問を読み取ります。
func readQuestion(args []string) (string, error) {
	question := strings.TrimSpace(strings.Join(args, " "))
	if question != "" {
		return question, nil
	}
	fmt.Print("質問: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", fmt.Errorf("read question: %w", err)
	}
	question = strings.TrimSpace(line)
	if question == "" {
		return "", fmt.Errorf("question must not be empty")
	}
	return question, nil
}

func printRetrievedChunks(results []rag.Result) {
	for i, result := range results {
		fmt.Printf("[%d] %s (score: %.4f)\n%s\n\n", i+1, result.Chunk.Source, result.Score, result.Chunk.Text)
	}
}

func generateAnswer(cfg config, question string, results []rag.Result) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	generator := rag.Generator{
		Endpoint:  cfg.endpoint,
		APIKey:    os.Getenv("TENSAI_API_KEY"),
		MaxTokens: cfg.maxTokens,
	}
	answer, err := generator.Generate(ctx, question, results)
	if err != nil {
		return err
	}
	fmt.Println(answer)
	fmt.Println("\n検索結果:")
	for i, result := range results {
		fmt.Printf("[%d] %s (score: %.4f)\n", i+1, result.Chunk.Source, result.Score)
	}
	return nil
}

func loadDotEnv(path string) error {
	if err := godotenv.Load(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("load %s: %w", path, err)
	}
	return nil
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
