package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
	"github.com/hiasinho/specter/internal/documents"
	"github.com/hiasinho/specter/internal/git"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff [path]",
	Short: "Show differences between local and remote documents",
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

		localByPath := make(map[string]api.Document)
		for _, d := range localDocs {
			localByPath[d.Path] = d
		}

		// If a specific path is given, only diff that document
		var pathFilter string
		if len(args) > 0 {
			pathFilter = args[0]
		}

		hasDiff := false
		for path, local := range localByPath {
			if pathFilter != "" && path != pathFilter {
				continue
			}

			remote, err := client.GetDocument(cfg.Project, branch, path)
			if err != nil {
				// Document doesn't exist remotely yet
				fmt.Printf("--- /dev/null\n+++ %s\n", path)
				for _, line := range strings.Split(local.ContentMD, "\n") {
					fmt.Printf("+ %s\n", line)
				}
				hasDiff = true
				continue
			}

			if local.ContentMD == remote.ContentMD {
				continue
			}

			hasDiff = true
			fmt.Printf("--- remote/%s\n+++ local/%s\n", path, path)
			printSimpleDiff(remote.ContentMD, local.ContentMD)
		}

		if !hasDiff {
			fmt.Println("No differences.")
		}

		return nil
	},
}

func printSimpleDiff(remote, local string) {
	remoteLines := strings.Split(remote, "\n")
	localLines := strings.Split(local, "\n")

	maxLen := len(remoteLines)
	if len(localLines) > maxLen {
		maxLen = len(localLines)
	}

	for i := 0; i < maxLen; i++ {
		var rLine, lLine string
		if i < len(remoteLines) {
			rLine = remoteLines[i]
		}
		if i < len(localLines) {
			lLine = localLines[i]
		}

		if rLine != lLine {
			if i < len(remoteLines) {
				fmt.Printf("- %s\n", rLine)
			}
			if i < len(localLines) {
				fmt.Printf("+ %s\n", lLine)
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(diffCmd)
}
