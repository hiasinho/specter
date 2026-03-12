package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
	"github.com/hiasinho/specter/internal/git"
	"github.com/spf13/cobra"
)

var historyLimit int

var historyCmd = &cobra.Command{
	Use:   "history <path> [revision]",
	Short: "Show document revision history or a specific revision",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, project, branch, err := historySetup()
		if err != nil {
			return err
		}

		docPath := args[0]

		// If a revision number is given, show that revision's content
		if len(args) == 2 {
			rev, err := strconv.Atoi(args[1])
			if err != nil {
				return fmt.Errorf("invalid revision number: %s", args[1])
			}
			revision, err := client.GetDocumentRevision(project, docPath, rev)
			if err != nil {
				return err
			}
			fmt.Printf("# %s (revision %d)\n", revision.Path, revision.Revision)
			fmt.Printf("# Author: %s <%s>\n", revision.Author.Username, revision.Author.Email)
			fmt.Printf("# Date:   %s\n\n", revision.CreatedAt)
			fmt.Print(revision.ContentMD)
			if len(revision.ContentMD) > 0 && revision.ContentMD[len(revision.ContentMD)-1] != '\n' {
				fmt.Println()
			}
			return nil
		}

		// Otherwise list revision history
		history, err := client.ListDocumentHistory(project, docPath, branch, historyLimit)
		if err != nil {
			return err
		}

		fmt.Printf("History for %s:\n\n", history.Path)
		for _, rev := range history.Revisions {
			fmt.Printf("  r%-4d  %s  %s\n", rev.Revision, rev.CreatedAt, rev.Author.Username)
		}

		return nil
	},
}

var historyDiffFrom int
var historyDiffTo int

var historyDiffCmd = &cobra.Command{
	Use:   "diff <path>",
	Short: "Compare two revisions of a document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, project, branch, err := historySetup()
		if err != nil {
			return err
		}

		if historyDiffFrom == 0 {
			return fmt.Errorf("--from is required")
		}

		result, err := client.GetDocumentDiff(project, args[0], branch, historyDiffFrom, historyDiffTo)
		if err != nil {
			return err
		}

		fmt.Printf("%s: revision %d → %d\n\n", result.Path, result.FromRevision, result.ToRevision)
		fmt.Println(result.Diff)
		return nil
	},
}

func historySetup() (*api.Client, string, string, error) {
	token := os.Getenv("SPECTER_TOKEN")
	if token == "" {
		return nil, "", "", fmt.Errorf("SPECTER_TOKEN environment variable not set")
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, "", "", err
	}

	repoRoot, err := config.FindRepoRoot(wd)
	if err != nil {
		return nil, "", "", err
	}

	cfg, err := config.Load(repoRoot)
	if err != nil {
		return nil, "", "", err
	}

	branch, err := git.CurrentBranch()
	if err != nil {
		return nil, "", "", err
	}

	return api.NewClient(token), cfg.Project, branch, nil
}

func init() {
	historyCmd.Flags().IntVar(&historyLimit, "limit", 10, "Maximum number of revisions to show")
	historyDiffCmd.Flags().IntVar(&historyDiffFrom, "from", 0, "Starting revision (required)")
	historyDiffCmd.Flags().IntVar(&historyDiffTo, "to", 0, "Ending revision (defaults to latest)")

	historyCmd.AddCommand(historyDiffCmd)
	rootCmd.AddCommand(historyCmd)
}
