package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hiasinho/specter/internal/api"
	"github.com/spf13/cobra"
)

var projectDeleteForce bool

var projectDeleteCmd = &cobra.Command{
	Use:   "delete <owner/slug>",
	Short: "Delete a project and all its data",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		project := args[0]

		token := os.Getenv("SPECTER_TOKEN")
		if token == "" {
			return fmt.Errorf("SPECTER_TOKEN environment variable not set")
		}

		if !projectDeleteForce {
			fmt.Printf("This will permanently delete %s and all its data. Continue? [y/N] ", project)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(answer)) != "y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		client := api.NewClient(token)
		if err := client.DeleteProject(project); err != nil {
			return err
		}

		fmt.Printf("Deleted %s\n", project)
		return nil
	},
}

func init() {
	projectDeleteCmd.Flags().BoolVar(&projectDeleteForce, "force", false, "Skip confirmation prompt")
	projectCmd.AddCommand(projectDeleteCmd)
	rootCmd.AddCommand(projectCmd)
}

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
}
