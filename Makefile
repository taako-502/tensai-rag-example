.PHONY: test run tensai

test:
	go test ./...

run:
	go run ./cmd/rag -docs ./docs

tensai:
	go run github.com/mattn/tensai/cmd/tensai@v0.0.26 serve -q8 -addr 127.0.0.1:8080
