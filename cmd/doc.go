package cmd

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
	"github.com/hiasinho/specter/internal/git"
	"github.com/spf13/cobra"
)

var docCmd = &cobra.Command{
	Use:   "doc",
	Short: "Manage individual documents",
}

var docPutCmd = &cobra.Command{
	Use:   "put <path>",
	Short: "Create or update a single document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docPath := args[0]

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

		contentBytes, err := os.ReadFile(docPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", docPath, err)
		}
		contentMD := string(contentBytes)
		hash := sha256.Sum256(contentBytes)
		_ = hash // computed for reference; API does not require it in the body

		client := api.NewClient(token)
		if err := client.PutDocument(cfg.Project, branch, docPath, contentMD); err != nil {
			return err
		}

		fmt.Printf("put %s -> %s/%s\n", docPath, cfg.Project, branch)
		return nil
	},
}

var docDeleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a single document",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		docPath := args[0]

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
		if err := client.DeleteDocument(cfg.Project, branch, docPath); err != nil {
			return err
		}

		fmt.Printf("deleted %s from %s/%s\n", docPath, cfg.Project, branch)
		return nil
	},
}

func init() {
	docCmd.AddCommand(docPutCmd)
	docCmd.AddCommand(docDeleteCmd)
	rootCmd.AddCommand(docCmd)
}
