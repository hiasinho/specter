package cmd

import (
	"fmt"
	"os"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
	"github.com/spf13/cobra"
)

var initName string

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the project on the Specter service",
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

		name := initName
		if name == "" {
			name = cfg.Project
		}

		client := api.NewClient(token)
		project, err := client.CreateProject(cfg.Project, name)
		if err != nil {
			return err
		}

		fmt.Printf("Project created: %s (%s)\n", project.Name, project.Slug)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initName, "name", "", "Project display name (defaults to slug)")
	rootCmd.AddCommand(initCmd)
}
