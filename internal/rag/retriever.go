package rag

import (
	"sort"

	"github.com/mattn/tensai"
)

type Result struct {
	Chunk Chunk
	Score float32
}

type Retriever struct {
	chunks     []Chunk
	vectors    [][]tensai.Float
	vectorizer *Vectorizer
}

func NewRetriever(chunks []Chunk) *Retriever {
	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Text
	}
	vectorizer := NewVectorizer(texts)
	vectors := make([][]tensai.Float, len(chunks))
	for i, text := range texts {
		vectors[i] = vectorizer.Transform(text)
	}
	return &Retriever{chunks: chunks, vectors: vectors, vectorizer: vectorizer}
}

// Searchは、クエリに関連するチャンクを検索します。
func (r *Retriever) Search(query string, topK int) []Result {
	if topK <= 0 {
		return nil
	}
	// queryVector は、検索クエリの正規化済みベクトル
	queryVector := r.vectorizer.Transform(query)
	results := make([]Result, 0, len(r.chunks))
	for i, vector := range r.vectors {
		// 正規化済みベクトルの内積からコサイン類似度を計算します。
		score := float32(tensai.DotVec(queryVector, vector))
		if score > 0 {
			results = append(results, Result{Chunk: r.chunks[i], Score: score})
		}
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].Score > results[j].Score })
	if len(results) > topK {
		results = results[:topK]
	}
	return results
}
