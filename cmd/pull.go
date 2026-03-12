package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
	"github.com/hiasinho/specter/internal/git"
	syncstate "github.com/hiasinho/specter/internal/sync"
	"github.com/spf13/cobra"
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Sync documents from the Specter service to local",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		repoRoot, err := config.FindRepoRoot(wd)
		if err != nil {
			return err
		}

		cfg, err := config.Load(repoRoot)
		if err != nil {
			return err
		}

		token := os.Getenv("SPECTER_TOKEN")
		if token == "" {
			return fmt.Errorf("SPECTER_TOKEN environment variable not set")
		}

		branch, err := git.CurrentBranch()
		if err != nil {
			return err
		}

		lastRevision, err := syncstate.ReadLastRevision(repoRoot)
		if err != nil {
			return err
		}

		client := api.NewClient(token)
		result, err := client.Pull(cfg.Project, branch, lastRevision)
		if err != nil {
			return err
		}

		if len(result.Documents) == 0 {
			fmt.Println("Already up to date.")
			return nil
		}

		var maxRevision int
		for _, doc := range result.Documents {
			if err := writeDocument(repoRoot, doc); err != nil {
				return err
			}
			fmt.Printf("  <- %s\n", doc.Path)
			if doc.Revision > maxRevision {
				maxRevision = doc.Revision
			}
		}

		if maxRevision > 0 {
			if err := syncstate.WriteLastRevision(repoRoot, maxRevision); err != nil {
				return err
			}
		}

		fmt.Printf("Pulled %d document(s) from %s/%s\n", len(result.Documents), cfg.Project, branch)
		return nil
	},
}

func writeDocument(repoRoot string, doc api.Document) error {
	path := filepath.Join(repoRoot, doc.Path)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating directory for %s: %w", doc.Path, err)
	}
	return os.WriteFile(path, []byte(doc.ContentMD), 0644)
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
