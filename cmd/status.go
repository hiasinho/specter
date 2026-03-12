package cmd

import (
	"fmt"
	"os"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
	"github.com/hiasinho/specter/internal/documents"
	"github.com/hiasinho/specter/internal/git"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show what's changed locally or remotely",
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

		localDocs, err := documents.Collect(repoRoot, cfg)
		if err != nil {
			return err
		}

		client := api.NewClient(token)
		remoteDocs, err := client.ListDocuments(cfg.Project, branch)
		if err != nil {
			return err
		}

		localByPath := make(map[string]api.Document)
		for _, d := range localDocs {
			localByPath[d.Path] = d
		}

		remoteByPath := make(map[string]api.Document)
		for _, d := range remoteDocs {
			remoteByPath[d.Path] = d
		}

		var modified, localOnly, remoteOnly []string

		for path, local := range localByPath {
			remote, exists := remoteByPath[path]
			if !exists {
				localOnly = append(localOnly, path)
			} else if local.ContentHash != remote.ContentHash {
				modified = append(modified, path)
			}
		}

		for path := range remoteByPath {
			if _, exists := localByPath[path]; !exists {
				remoteOnly = append(remoteOnly, path)
			}
		}

		if len(modified) == 0 && len(localOnly) == 0 && len(remoteOnly) == 0 {
			fmt.Printf("In sync with %s/%s\n", cfg.Project, branch)
			showPendingProposals(client, cfg.Project)
			return nil
		}

		fmt.Printf("Status for %s/%s:\n", cfg.Project, branch)
		for _, p := range modified {
			fmt.Printf("  ~ %s (modified)\n", p)
		}
		for _, p := range localOnly {
			fmt.Printf("  + %s (local only)\n", p)
		}
		for _, p := range remoteOnly {
			fmt.Printf("  - %s (remote only)\n", p)
		}

		showPendingProposals(client, cfg.Project)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
