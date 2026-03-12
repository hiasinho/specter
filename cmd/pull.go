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

var pullForce bool

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

		state, err := syncstate.ReadState(repoRoot)
		if err != nil {
			return err
		}

		client := api.NewClient(token)
		result, err := client.Pull(cfg.Project, branch, state.LastRevision)
		if err != nil {
			return err
		}

		if len(result.Documents) == 0 {
			fmt.Println("Already up to date.")
			showPendingProposals(client, cfg.Project)
			return nil
		}

		var conflicts []string
		var maxRevision int
		for _, doc := range result.Documents {
			localPath := filepath.Join(repoRoot, doc.Path)
			if existing, err := os.ReadFile(localPath); err == nil {
				if string(existing) != doc.ContentMD && !pullForce {
					conflicts = append(conflicts, doc.Path)
					if doc.Revision > maxRevision {
						maxRevision = doc.Revision
					}
					continue
				}
			}

			if err := writeDocument(repoRoot, doc); err != nil {
				return err
			}
			fmt.Printf("  <- %s\n", doc.Path)
			if doc.Revision > maxRevision {
				maxRevision = doc.Revision
			}
		}

		if maxRevision > 0 {
			state.LastRevision = &maxRevision
			state.SyncedAt = result.SyncedAt
			if err := syncstate.WriteState(repoRoot, state); err != nil {
				return err
			}
		}

		pulled := len(result.Documents) - len(conflicts)
		if pulled > 0 {
			fmt.Printf("Pulled %d document(s) from %s/%s\n", pulled, cfg.Project, branch)
		}

		if len(conflicts) > 0 {
			fmt.Printf("\nSkipped %d file(s) with local changes:\n", len(conflicts))
			for _, p := range conflicts {
				fmt.Printf("  ! %s\n", p)
			}
			fmt.Println("Use --force to overwrite.")
		}

		showPendingProposals(client, cfg.Project)
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

func showPendingProposals(client *api.Client, project string) {
	proposals, err := client.ListProposals(project, "", "pending")
	if err == nil && len(proposals) > 0 {
		fmt.Printf("\n%d pending proposal(s) — run `specter proposals` to view\n", len(proposals))
	}
}

func init() {
	pullCmd.Flags().BoolVar(&pullForce, "force", false, "Overwrite local changes")
	rootCmd.AddCommand(pullCmd)
}
