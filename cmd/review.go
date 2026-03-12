package cmd

import (
	"fmt"
	"os"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review <id> <accept|reject>",
	Short: "Accept or reject a proposal",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		action := args[1]

		if action != "accept" && action != "reject" {
			return fmt.Errorf("action must be 'accept' or 'reject', got %q", action)
		}

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

		client := api.NewClient(token)
		// API expects "accepted"/"rejected"
		status := action + "ed"
		if err := client.UpdateProposalStatus(cfg.Project, id, status); err != nil {
			return err
		}

		fmt.Printf("Proposal %s %s.\n", id, status)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}
