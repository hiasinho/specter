package cmd

import (
	"fmt"
	"os"

	"github.com/hiasinho/specter/internal/api"
	"github.com/spf13/cobra"
)

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Show current user info",
	RunE: func(cmd *cobra.Command, args []string) error {
		token := os.Getenv("SPECTER_TOKEN")
		if token == "" {
			return fmt.Errorf("SPECTER_TOKEN environment variable not set")
		}

		client := api.NewClient(token)
		user, err := client.Me()
		if err != nil {
			return err
		}

		fmt.Printf("Username: %s\n", user.Username)
		fmt.Printf("Email:    %s\n", user.Email)
		fmt.Printf("ID:       %s\n", user.ID)

		if len(user.InviteCodes) > 0 {
			fmt.Println("\nInvite codes:")
			for _, code := range user.InviteCodes {
				status := "available"
				if code.Redeemed {
					status = "redeemed"
				}
				fmt.Printf("  %s (%s)\n", code.Code, status)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(meCmd)
}
