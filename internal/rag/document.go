package rag

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

// Document is a text file loaded from the knowledge base.
type Document struct {
	Path string
	Text string
}

// Chunk is a searchable part of a document.
type Chunk struct {
	Source string
	Text   string
}

// LoadDocumentsは、指定ディレクトリ（デフォルト: ./docs）内のMarkdownとテキストファイルを読み込みます。
func LoadDocuments(root string) ([]Document, error) {
	var docs []Document
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		docs = append(docs, Document{Path: filepath.ToSlash(rel), Text: string(body)})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load documents: %w", err)
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].Path < docs[j].Path })
	if len(docs) == 0 {
		return nil, fmt.Errorf("no .md or .txt files found under %s", root)
	}
	return docs, nil
}

// SplitDocumentsは、文書を指定された文字数以内のチャンクに分割します。
func SplitDocuments(docs []Document, maxRunes int) []Chunk {
	if maxRunes <= 0 {
		maxRunes = 600
	}
	var chunks []Chunk
	for _, doc := range docs {
		paragraphs := strings.FieldsFunc(strings.ReplaceAll(doc.Text, "\r\n", "\n"), func(r rune) bool {
			return r == '\n'
		})
		var current strings.Builder
		flush := func() {
			text := strings.TrimSpace(current.String())
			if text != "" {
				chunks = append(chunks, Chunk{Source: doc.Path, Text: text})
			}
			current.Reset()
		}
		for _, paragraph := range paragraphs {
			paragraph = strings.TrimSpace(paragraph)
			if paragraph == "" {
				continue
			}
			for utf8.RuneCountInString(paragraph) > maxRunes {
				flush()
				runes := []rune(paragraph)
				chunks = append(chunks, Chunk{Source: doc.Path, Text: string(runes[:maxRunes])})
				paragraph = string(runes[maxRunes:])
			}
			additional := utf8.RuneCountInString(paragraph)
			if current.Len() > 0 {
				additional++
			}
			if utf8.RuneCountInString(current.String())+additional > maxRunes {
				flush()
			}
			if current.Len() > 0 {
				current.WriteByte('\n')
			}
			current.WriteString(paragraph)
		}
		flush()
	}
	return chunks
}
