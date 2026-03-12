package documents

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hiasinho/specter/internal/config"
)

func TestCollect(t *testing.T) {
	dir := t.TempDir()

	// Create directory structure
	os.MkdirAll(filepath.Join(dir, "specs"), 0755)
	os.MkdirAll(filepath.Join(dir, "specs", "drafts"), 0755)
	os.MkdirAll(filepath.Join(dir, "docs"), 0755)

	os.WriteFile(filepath.Join(dir, "specs", "api.md"), []byte("# API"), 0644)
	os.WriteFile(filepath.Join(dir, "specs", "cli.md"), []byte("# CLI"), 0644)
	os.WriteFile(filepath.Join(dir, "specs", "drafts", "wip.md"), []byte("# WIP"), 0644)
	os.WriteFile(filepath.Join(dir, "specs", "notes.txt"), []byte("not markdown"), 0644)
	os.WriteFile(filepath.Join(dir, "docs", "readme.md"), []byte("# Readme"), 0644)

	t.Run("collects markdown files from configured paths", func(t *testing.T) {
		cfg := &config.Config{
			Project: "test",
			Paths:   []string{"specs/"},
		}

		docs, err := Collect(dir, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should find api.md, cli.md, and drafts/wip.md but not notes.txt
		if len(docs) != 3 {
			t.Fatalf("expected 3 documents, got %d", len(docs))
		}

		paths := make(map[string]bool)
		for _, d := range docs {
			paths[d.Path] = true
		}
		if !paths["specs/api.md"] {
			t.Error("missing specs/api.md")
		}
		if !paths["specs/cli.md"] {
			t.Error("missing specs/cli.md")
		}
	})

	t.Run("respects exclude patterns", func(t *testing.T) {
		cfg := &config.Config{
			Project: "test",
			Paths:   []string{"specs/"},
			Exclude: []string{"specs/drafts/"},
		}

		docs, err := Collect(dir, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(docs) != 2 {
			t.Fatalf("expected 2 documents, got %d", len(docs))
		}

		for _, d := range docs {
			if d.Path == "specs/drafts/wip.md" {
				t.Error("should have excluded specs/drafts/wip.md")
			}
		}
	})

	t.Run("multiple paths", func(t *testing.T) {
		cfg := &config.Config{
			Project: "test",
			Paths:   []string{"specs/", "docs/"},
			Exclude: []string{"specs/drafts/"},
		}

		docs, err := Collect(dir, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(docs) != 3 {
			t.Fatalf("expected 3 documents, got %d", len(docs))
		}
	})

	t.Run("computes content hash", func(t *testing.T) {
		cfg := &config.Config{
			Project: "test",
			Paths:   []string{"docs/"},
		}

		docs, err := Collect(dir, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(docs) != 1 {
			t.Fatalf("expected 1 document, got %d", len(docs))
		}

		if docs[0].ContentHash == "" {
			t.Error("expected content hash to be set")
		}
		if len(docs[0].ContentHash) != 64 {
			t.Errorf("expected 64-char SHA-256 hash, got %d chars", len(docs[0].ContentHash))
		}
	})

	t.Run("empty directory", func(t *testing.T) {
		emptyDir := filepath.Join(dir, "empty")
		os.MkdirAll(emptyDir, 0755)

		cfg := &config.Config{
			Project: "test",
			Paths:   []string{"empty/"},
		}

		docs, err := Collect(dir, cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(docs) != 0 {
			t.Errorf("expected 0 documents, got %d", len(docs))
		}
	})
}

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		path     string
		patterns []string
		expected bool
	}{
		{"specs/drafts/wip.md", []string{"specs/drafts/"}, true},
		{"specs/api.md", []string{"specs/drafts/"}, false},
		{"specs/_wip_notes.md", []string{"_wip_*.md"}, true},
		{"deep/nested/_wip_foo.md", []string{"**/_wip_*.md"}, true},
		{"specs/api.md", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isExcluded(tt.path, tt.patterns)
			if result != tt.expected {
				t.Errorf("isExcluded(%q, %v) = %v, want %v", tt.path, tt.patterns, result, tt.expected)
			}
		})
	}
}
