package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReadState(t *testing.T) {
	t.Run("no state file", func(t *testing.T) {
		dir := t.TempDir()
		state, err := ReadState(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.LastRevision != nil {
			t.Errorf("expected nil LastRevision, got %d", *state.LastRevision)
		}
		if state.SyncedAt != "" {
			t.Errorf("expected empty SyncedAt, got %q", state.SyncedAt)
		}
	})

	t.Run("legacy integer format", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".specter-sync"), []byte("42\n"), 0644)

		state, err := ReadState(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.LastRevision == nil || *state.LastRevision != 42 {
			t.Errorf("expected LastRevision 42, got %v", state.LastRevision)
		}
		if state.SyncedAt != "" {
			t.Errorf("expected empty SyncedAt for legacy format")
		}
	})

	t.Run("json format", func(t *testing.T) {
		dir := t.TempDir()
		rev := 10
		state := SyncState{LastRevision: &rev, SyncedAt: "2025-01-01T00:00:00Z"}
		data, _ := json.Marshal(state)
		os.WriteFile(filepath.Join(dir, ".specter-sync"), data, 0644)

		got, err := ReadState(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.LastRevision == nil || *got.LastRevision != 10 {
			t.Errorf("expected LastRevision 10, got %v", got.LastRevision)
		}
		if got.SyncedAt != "2025-01-01T00:00:00Z" {
			t.Errorf("expected SyncedAt '2025-01-01T00:00:00Z', got %q", got.SyncedAt)
		}
	})

	t.Run("invalid content", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".specter-sync"), []byte("not-valid"), 0644)

		_, err := ReadState(dir)
		if err == nil {
			t.Fatal("expected error for invalid content")
		}
	})
}

func TestWriteState(t *testing.T) {
	dir := t.TempDir()
	rev := 99
	state := &SyncState{LastRevision: &rev, SyncedAt: "2025-06-15T12:00:00Z"}

	err := WriteState(dir, state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify round-trip
	got, err := ReadState(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.LastRevision == nil || *got.LastRevision != 99 {
		t.Errorf("expected LastRevision 99, got %v", got.LastRevision)
	}
	if got.SyncedAt != "2025-06-15T12:00:00Z" {
		t.Errorf("expected SyncedAt '2025-06-15T12:00:00Z', got %q", got.SyncedAt)
	}
}
