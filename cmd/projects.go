package cmd

import (
	"fmt"
	"os"

	"github.com/hiasinho/specter/internal/api"
	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List projects you are a member of",
	RunE: func(cmd *cobra.Command, args []string) error {
		token := os.Getenv("SPECTER_TOKEN")
		if token == "" {
			return fmt.Errorf("SPECTER_TOKEN environment variable not set")
		}

		client := api.NewClient(token)
		projects, err := client.ListProjects()
		if err != nil {
			return err
		}

		if len(projects) == 0 {
			fmt.Println("No projects found.")
			return nil
		}

		for _, p := range projects {
			fmt.Printf("%s  %s  (%s)\n", p.FullName, p.Name, p.Role)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)
}
