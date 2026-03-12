package cmd

import (
	"fmt"

	"github.com/hiasinho/specter/internal/api"
	"github.com/spf13/cobra"
)

var (
	registerEmail      string
	registerUsername   string
	registerInviteCode string
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Create a new Specter user account",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := api.NewClient("")
		result, err := client.Register(registerEmail, registerUsername, registerInviteCode)
		if err != nil {
			return err
		}

		fmt.Printf("Registration successful!\n\n")
		fmt.Printf("Username: %s\n", result.Username)
		fmt.Printf("Email:    %s\n", result.Email)
		fmt.Printf("ID:       %s\n", result.ID)
		fmt.Printf("\nToken (save this — it won't be shown again):\n  %s\n", result.Token)

		if len(result.InviteCodes) > 0 {
			fmt.Println("\nInvite codes:")
			for _, code := range result.InviteCodes {
				fmt.Printf("  %s\n", code)
			}
		}

		return nil
	},
}

func init() {
	registerCmd.Flags().StringVar(&registerEmail, "email", "", "Email address (required)")
	registerCmd.Flags().StringVar(&registerUsername, "username", "", "Username (required)")
	registerCmd.Flags().StringVar(&registerInviteCode, "invite-code", "", "Invite code (required)")
	_ = registerCmd.MarkFlagRequired("email")
	_ = registerCmd.MarkFlagRequired("username")
	_ = registerCmd.MarkFlagRequired("invite-code")
	rootCmd.AddCommand(registerCmd)
}
