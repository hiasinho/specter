package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const stateFile = ".specter-sync"

// SyncState holds the persisted sync state.
type SyncState struct {
	LastRevision *int   `json:"last_revision,omitempty"`
	SyncedAt     string `json:"synced_at,omitempty"`
}

// ReadState reads the sync state from .specter-sync.
// Handles both the new JSON format and the legacy integer format.
func ReadState(dir string) (*SyncState, error) {
	data, err := os.ReadFile(filepath.Join(dir, stateFile))
	if os.IsNotExist(err) {
		return &SyncState{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading sync state: %w", err)
	}

	content := strings.TrimSpace(string(data))

	// Try JSON first (new format)
	var state SyncState
	if err := json.Unmarshal([]byte(content), &state); err == nil {
		return &state, nil
	}

	// Fall back to legacy integer format
	rev, err := strconv.Atoi(content)
	if err != nil {
		return nil, fmt.Errorf("parsing sync state: %w", err)
	}
	return &SyncState{LastRevision: &rev}, nil
}

// WriteState writes the sync state to .specter-sync.
func WriteState(dir string, state *SyncState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("writing sync state: %w", err)
	}
	return os.WriteFile(filepath.Join(dir, stateFile), append(data, '\n'), 0644)
}
