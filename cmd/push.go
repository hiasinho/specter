package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
	"github.com/hiasinho/specter/internal/documents"
	"github.com/hiasinho/specter/internal/git"
	syncstate "github.com/hiasinho/specter/internal/sync"
	"github.com/spf13/cobra"
)

var pushForce bool

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

		state, err := syncstate.ReadState(repoRoot)
		if err != nil {
			return err
		}

		req := &api.SyncPushRequest{
			Branch:    branch,
			Documents: docs,
		}
		if !pushForce && state.SyncedAt != "" {
			req.BaseRevision = state.SyncedAt
		}

		client := api.NewClient(token)
		result, err := client.Push(cfg.Project, req)
		if err != nil {
			var conflict *api.ConflictError
			if errors.As(err, &conflict) {
				fmt.Fprintf(os.Stderr, "Push rejected: %d document(s) modified on server since last pull:\n", len(conflict.Conflicts))
				for _, c := range conflict.Conflicts {
					fmt.Fprintf(os.Stderr, "  ! %s (server revision %d)\n", c.Path, c.ServerRevision)
				}
				fmt.Fprintf(os.Stderr, "\nRun `specter pull` to sync, then push again.\n")
				return fmt.Errorf("push aborted due to conflicts")
			}
			return err
		}

		if result.SyncedAt != "" {
			state.SyncedAt = result.SyncedAt
			if err := syncstate.WriteState(repoRoot, state); err != nil {
				return fmt.Errorf("updating sync state: %w", err)
			}
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
	pushCmd.Flags().BoolVar(&pushForce, "force", false, "Skip conflict detection (last write wins)")
	rootCmd.AddCommand(pushCmd)
}
