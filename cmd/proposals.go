package cmd

import (
	"fmt"
	"os"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
	"github.com/spf13/cobra"
)

var (
	proposalsDocument string
	proposalsStatus   string
)

var proposalsCmd = &cobra.Command{
	Use:   "proposals",
	Short: "List proposals for the project",
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

		client := api.NewClient(token)
		proposals, err := client.ListProposals(cfg.Project, proposalsDocument, proposalsStatus)
		if err != nil {
			return err
		}

		if len(proposals) == 0 {
			fmt.Println("No proposals found.")
			return nil
		}

		for _, p := range proposals {
			status := p.Status
			if status == "" {
				status = "pending"
			}
			line := ""
			if p.AnchorLineHint > 0 {
				line = fmt.Sprintf(":%d", p.AnchorLineHint)
			}
			fmt.Printf("  %-10s %-8s %-8s %s%s\n", p.ID, status, p.Type, p.DocumentPath, line)
			if p.Body != "" {
				summary := p.Body
				if len(summary) > 80 {
					summary = summary[:77] + "..."
				}
				fmt.Printf("             %s\n", summary)
			}
		}

		return nil
	},
}

func init() {
	proposalsCmd.Flags().StringVar(&proposalsDocument, "document", "", "Filter by document path")
	proposalsCmd.Flags().StringVar(&proposalsStatus, "status", "pending", "Filter by status (pending|accepted|rejected)")
	rootCmd.AddCommand(proposalsCmd)
}
