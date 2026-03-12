package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	specterFile := filepath.Join(dir, ".specter")

	t.Run("valid config", func(t *testing.T) {
		content := "project: hiasinho/my-project\npaths:\n  - specs/\n  - docs/\nexclude:\n  - drafts/\n"
		os.WriteFile(specterFile, []byte(content), 0644)

		cfg, err := Load(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Project != "hiasinho/my-project" {
			t.Errorf("expected project 'hiasinho/my-project', got %q", cfg.Project)
		}
		if cfg.ProjectOwner() != "hiasinho" {
			t.Errorf("expected owner 'hiasinho', got %q", cfg.ProjectOwner())
		}
		if cfg.ProjectSlug() != "my-project" {
			t.Errorf("expected slug 'my-project', got %q", cfg.ProjectSlug())
		}
		if len(cfg.Paths) != 2 {
			t.Errorf("expected 2 paths, got %d", len(cfg.Paths))
		}
		if len(cfg.Exclude) != 1 {
			t.Errorf("expected 1 exclude, got %d", len(cfg.Exclude))
		}
	})

	t.Run("missing project", func(t *testing.T) {
		os.WriteFile(specterFile, []byte("paths:\n  - specs/\n"), 0644)
		_, err := Load(dir)
		if err == nil {
			t.Fatal("expected error for missing project")
		}
	})

	t.Run("missing paths", func(t *testing.T) {
		os.WriteFile(specterFile, []byte("project: owner/foo\n"), 0644)
		_, err := Load(dir)
		if err == nil {
			t.Fatal("expected error for missing paths")
		}
	})

	t.Run("bare slug without owner", func(t *testing.T) {
		os.WriteFile(specterFile, []byte("project: my-project\npaths:\n  - specs/\n"), 0644)
		_, err := Load(dir)
		if err == nil {
			t.Fatal("expected error for bare slug without owner")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := Load(t.TempDir())
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		os.WriteFile(specterFile, []byte(":::invalid"), 0644)
		_, err := Load(dir)
		if err == nil {
			t.Fatal("expected error for invalid yaml")
		}
	})
}

func TestFindRepoRoot(t *testing.T) {
	t.Run("finds root in current dir", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".specter"), []byte("project: x\n"), 0644)

		root, err := FindRepoRoot(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != dir {
			t.Errorf("expected %q, got %q", dir, root)
		}
	})

	t.Run("finds root in parent dir", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".specter"), []byte("project: x\n"), 0644)
		child := filepath.Join(dir, "sub", "deep")
		os.MkdirAll(child, 0755)

		root, err := FindRepoRoot(child)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != dir {
			t.Errorf("expected %q, got %q", dir, root)
		}
	})

	t.Run("no config found", func(t *testing.T) {
		dir := t.TempDir()
		_, err := FindRepoRoot(dir)
		if err == nil {
			t.Fatal("expected error when no .specter found")
		}
	})
}
