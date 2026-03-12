package documents

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
)

// Collect reads all matching markdown files from the configured paths,
// respecting exclude patterns. Returns documents with paths relative to repoRoot.
func Collect(repoRoot string, cfg *config.Config) ([]api.Document, error) {
	var docs []api.Document

	for _, syncPath := range cfg.Paths {
		absPath := filepath.Join(repoRoot, syncPath)
		err := filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".md") {
				return nil
			}

			relPath, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}

			if isExcluded(relPath, cfg.Exclude) {
				return nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading %s: %w", relPath, err)
			}

			hash := sha256.Sum256(content)
			docs = append(docs, api.Document{
				Path:        relPath,
				ContentMD:   string(content),
				ContentHash: fmt.Sprintf("%x", hash),
			})
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walking %s: %w", syncPath, err)
		}
	}

	return docs, nil
}

func isExcluded(path string, patterns []string) bool {
	for _, pattern := range patterns {
		// Match against the full path
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		// Match against each path segment's suffix (e.g. pattern "*.tmp" matches "a/b/foo.tmp")
		base := filepath.Base(path)
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
		// Handle ** patterns: try matching the non-** part against the base name
		if strings.Contains(pattern, "**") {
			trimmed := strings.TrimPrefix(pattern, "**/")
			if matched, _ := filepath.Match(trimmed, base); matched {
				return true
			}
		}
		// Check if path starts with an excluded directory
		if strings.HasSuffix(pattern, "/") && strings.HasPrefix(path, pattern) {
			return true
		}
	}
	return false
}
