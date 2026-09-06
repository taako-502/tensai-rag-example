package rag

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/mattn/tensai"
)

// Vectorizer converts text to normalized TF-IDF vectors. Japanese text is
// represented with rune bigrams; Latin words and numbers use word tokens.
type Vectorizer struct {
	vocabulary map[string]int
	idf        []float32
}

func NewVectorizer(texts []string) *Vectorizer {
	documentFrequency := make(map[string]int)
	for _, text := range texts {
		seen := make(map[string]struct{})
		for _, token := range tokenize(text) {
			seen[token] = struct{}{}
		}
		for token := range seen {
			documentFrequency[token]++
		}
	}

	tokens := make([]string, 0, len(documentFrequency))
	for token := range documentFrequency {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)

	v := &Vectorizer{
		vocabulary: make(map[string]int, len(tokens)),
		idf:        make([]float32, len(tokens)),
	}
	for i, token := range tokens {
		v.vocabulary[token] = i
		v.idf[i] = float32(math.Log((1+float64(len(texts)))/(1+float64(documentFrequency[token]))) + 1)
	}
	return v
}

// Transformは、テキストをTF-IDFベクトルに変換します。
func (v *Vectorizer) Transform(text string) []tensai.Float {
	vector := make([]tensai.Float, len(v.vocabulary))
	for _, token := range tokenize(text) {
		if i, ok := v.vocabulary[token]; ok {
			vector[i] += tensai.Float(v.idf[i])
		}
	}

	// Unit normalization changes the tensai dot product into cosine similarity.
	norm := float32(math.Sqrt(float64(tensai.DotVec(vector, vector))))
	if norm == 0 {
		return vector
	}
	for i := range vector {
		vector[i] /= tensai.Float(norm)
	}
	return vector
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var word []rune
	var japanese []rune
	flushWord := func() {
		if len(word) > 0 {
			tokens = append(tokens, "w:"+string(word))
			word = word[:0]
		}
	}
	flushJapanese := func() {
		if len(japanese) == 1 {
			tokens = append(tokens, "j:"+string(japanese))
		}
		for i := 0; i+1 < len(japanese); i++ {
			tokens = append(tokens, "j:"+string(japanese[i:i+2]))
		}
		japanese = japanese[:0]
	}

	for _, r := range text {
		switch {
		case unicode.In(r, unicode.Hiragana, unicode.Katakana, unicode.Han):
			flushWord()
			japanese = append(japanese, r)
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			flushJapanese()
			word = append(word, r)
		default:
			flushWord()
			flushJapanese()
		}
	}
	flushWord()
	flushJapanese()
	return tokens
}
