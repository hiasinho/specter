package cmd

import (
	"fmt"
	"os"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
	"github.com/hiasinho/specter/internal/git"
	"github.com/spf13/cobra"
)

var (
	proposeType   string
	proposeAnchor string
	proposeLine   int
	proposeBody   string
)

var proposeCmd = &cobra.Command{
	Use:   "propose <document-path>",
	Short: "Create a proposal for a document",
	Args:  cobra.ExactArgs(1),
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

		client := api.NewClient(token)
		proposal, err := client.CreateProposal(cfg.Project, &api.Proposal{
			DocumentPath:   args[0],
			Branch:         branch,
			Type:           proposeType,
			AnchorContent:  proposeAnchor,
			AnchorLineHint: proposeLine,
			Body:           proposeBody,
		})
		if err != nil {
			return err
		}

		fmt.Printf("Proposal created: %s\n", proposal.ID)
		return nil
	},
}

func init() {
	proposeCmd.Flags().StringVar(&proposeType, "type", "", "Proposal type (replace|insert|delete|note)")
	proposeCmd.Flags().StringVar(&proposeAnchor, "anchor", "", "Text snippet to anchor to")
	proposeCmd.Flags().IntVar(&proposeLine, "line", 0, "Line number hint for the anchor")
	proposeCmd.Flags().StringVar(&proposeBody, "body", "", "Proposed content or rationale")
	_ = proposeCmd.MarkFlagRequired("type")
	_ = proposeCmd.MarkFlagRequired("anchor")
	_ = proposeCmd.MarkFlagRequired("body")
	rootCmd.AddCommand(proposeCmd)
}
