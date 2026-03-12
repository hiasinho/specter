package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hiasinho/specter/internal/config"
)

func TestGenerateSkillBlock(t *testing.T) {
	cfg := &config.Config{
		Project: "test-project",
		Paths:   []string{"docs/"},
	}

	result := generateSkillBlock(cfg)

	t.Run("contains project name", func(t *testing.T) {
		if !strings.Contains(result, "test-project") {
			t.Error("expected output to contain project name")
		}
	})

	t.Run("contains workflow steps", func(t *testing.T) {
		for _, cmd := range []string{"specter pull", "specter push", "specter status", "specter diff"} {
			if !strings.Contains(result, cmd) {
				t.Errorf("expected output to contain %q", cmd)
			}
		}
	})

	t.Run("contains header", func(t *testing.T) {
		if !strings.Contains(result, "## Specter") {
			t.Error("expected output to contain ## Specter header")
		}
	})
}

func TestInstallSkillBlock(t *testing.T) {
	block := "## Specter\n\nTest block\n"

	t.Run("creates AGENTS.md when missing", func(t *testing.T) {
		dir := t.TempDir()

		if err := installSkillBlock(dir, block); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
		if err != nil {
			t.Fatalf("could not read AGENTS.md: %v", err)
		}

		if !strings.Contains(string(data), "Test block") {
			t.Error("expected AGENTS.md to contain skill block")
		}
	})

	t.Run("appends to existing AGENTS.md", func(t *testing.T) {
		dir := t.TempDir()
		existing := "# Agents\n\nExisting content.\n"
		os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(existing), 0644)

		if err := installSkillBlock(dir, block); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
		if err != nil {
			t.Fatalf("could not read AGENTS.md: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "Existing content.") {
			t.Error("expected AGENTS.md to preserve existing content")
		}
		if !strings.Contains(content, "Test block") {
			t.Error("expected AGENTS.md to contain appended skill block")
		}
	})

	t.Run("adds newline before block if missing", func(t *testing.T) {
		dir := t.TempDir()
		existing := "no trailing newline"
		os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(existing), 0644)

		if err := installSkillBlock(dir, block); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		data, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
		content := string(data)

		// Should have added a newline before the separator
		if !strings.Contains(content, "no trailing newline\n\n") {
			t.Error("expected newline to be added before skill block")
		}
	})
}
