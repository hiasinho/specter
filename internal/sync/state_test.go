package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadLastRevision(t *testing.T) {
	t.Run("no state file", func(t *testing.T) {
		dir := t.TempDir()
		rev, err := ReadLastRevision(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rev != nil {
			t.Errorf("expected nil, got %d", *rev)
		}
	})

	t.Run("valid state file", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".specter-sync"), []byte("42\n"), 0644)

		rev, err := ReadLastRevision(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rev == nil {
			t.Fatal("expected non-nil revision")
		}
		if *rev != 42 {
			t.Errorf("expected 42, got %d", *rev)
		}
	})

	t.Run("invalid content", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".specter-sync"), []byte("not-a-number"), 0644)

		_, err := ReadLastRevision(dir)
		if err == nil {
			t.Fatal("expected error for invalid content")
		}
	})
}

func TestWriteLastRevision(t *testing.T) {
	dir := t.TempDir()

	err := WriteLastRevision(dir, 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".specter-sync"))
	if err != nil {
		t.Fatalf("unexpected error reading file: %v", err)
	}
	if string(data) != "99\n" {
		t.Errorf("expected '99\\n', got %q", string(data))
	}

	// Verify round-trip
	rev, err := ReadLastRevision(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *rev != 99 {
		t.Errorf("expected 99, got %d", *rev)
	}
}
