package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const stateFile = ".specter-sync"

// ReadLastRevision reads the last synced revision from .specter-sync.
func ReadLastRevision(dir string) (*int, error) {
	data, err := os.ReadFile(filepath.Join(dir, stateFile))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading sync state: %w", err)
	}

	rev, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return nil, fmt.Errorf("parsing sync state: %w", err)
	}
	return &rev, nil
}

// WriteLastRevision writes the last synced revision to .specter-sync.
func WriteLastRevision(dir string, revision int) error {
	path := filepath.Join(dir, stateFile)
	return os.WriteFile(path, []byte(fmt.Sprintf("%d\n", revision)), 0644)
}
