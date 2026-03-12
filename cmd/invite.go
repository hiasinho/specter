package cmd

import (
	"fmt"
	"os"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
	"github.com/spf13/cobra"
)

var inviteRole string

var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Manage project invites",
}

var inviteCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an invite for the current project",
	RunE: func(cmd *cobra.Command, args []string) error {
		token := os.Getenv("SPECTER_TOKEN")
		if token == "" {
			return fmt.Errorf("SPECTER_TOKEN environment variable not set")
		}

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
			return fmt.Errorf("loading config: %w", err)
		}

		client := api.NewClient(token)
		invite, err := client.CreateInvite(cfg.Project, inviteRole)
		if err != nil {
			return err
		}

		fmt.Printf("Invite created for %s\n", cfg.Project)
		fmt.Printf("Role: %s\n", invite.Role)
		fmt.Printf("Code: %s\n", invite.Code)
		return nil
	},
}

var inviteRedeemCmd = &cobra.Command{
	Use:   "redeem",
	Short: "Redeem an invite code to join a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		token := os.Getenv("SPECTER_TOKEN")
		if token == "" {
			return fmt.Errorf("SPECTER_TOKEN environment variable not set")
		}

		code, _ := cmd.Flags().GetString("code")
		if code == "" {
			return fmt.Errorf("--code is required")
		}

		client := api.NewClient(token)
		result, err := client.RedeemInvite(code)
		if err != nil {
			return err
		}

		fmt.Printf("Joined %s as %s\n", result.Project.FullName, result.Role)
		return nil
	},
}

func init() {
	inviteCreateCmd.Flags().StringVar(&inviteRole, "role", "reader", "Role for the invite (editor|reviewer|reader)")
	inviteRedeemCmd.Flags().String("code", "", "Invite code to redeem")

	inviteCmd.AddCommand(inviteCreateCmd)
	inviteCmd.AddCommand(inviteRedeemCmd)
	rootCmd.AddCommand(inviteCmd)
}
