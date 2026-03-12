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

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Sync local documents to the Specter service",
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

		docs, err := documents.Collect(repoRoot, cfg)
		if err != nil {
			return err
		}

		if len(docs) == 0 {
			fmt.Println("No documents to push.")
			return nil
		}

		client := api.NewClient(token)
		result, err := client.Push(cfg.Project, &api.SyncPushRequest{
			Branch:    branch,
			Documents: docs,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Pushed to %s/%s:\n", cfg.Project, branch)
		if len(result.Created) > 0 {
			fmt.Printf("  created: %d\n", len(result.Created))
			for _, p := range result.Created {
				fmt.Printf("    + %s\n", p)
			}
		}
		if len(result.Updated) > 0 {
			fmt.Printf("  updated: %d\n", len(result.Updated))
			for _, p := range result.Updated {
				fmt.Printf("    ~ %s\n", p)
			}
		}
		if len(result.Unchanged) > 0 {
			fmt.Printf("  unchanged: %d\n", len(result.Unchanged))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
