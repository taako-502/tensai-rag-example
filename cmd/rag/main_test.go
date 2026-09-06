package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	const name = "TENSAI_RAG_TEST_VALUE"
	t.Setenv(name, "")
	if err := os.Unsetenv(name); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(name+"=from-dotenv\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := loadDotEnv(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv(name); got != "from-dotenv" {
		t.Fatalf("got %q, want %q", got, "from-dotenv")
	}
}

func TestLoadDotEnvDoesNotOverrideEnvironment(t *testing.T) {
	const name = "TENSAI_RAG_TEST_VALUE"
	t.Setenv(name, "from-environment")

	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(name+"=from-dotenv\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := loadDotEnv(path); err != nil {
		t.Fatal(err)
	}
	if got := os.Getenv(name); got != "from-environment" {
		t.Fatalf("got %q, want %q", got, "from-environment")
	}
}

func TestLoadDotEnvAllowsMissingFile(t *testing.T) {
	if err := loadDotEnv(filepath.Join(t.TempDir(), ".env")); err != nil {
		t.Fatal(err)
	}
}
