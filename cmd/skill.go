package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hiasinho/specter/internal/config"
	"github.com/spf13/cobra"
)

var skillInstall bool

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Output agent instructions for the Specter service",
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

		block := generateSkillBlock(cfg)

		if skillInstall {
			return installSkillBlock(repoRoot, block)
		}

		fmt.Print(block)
		return nil
	},
}

func generateSkillBlock(cfg *config.Config) string {
	return fmt.Sprintf(`## Specter

This project uses [Specter](https://hiasinho.github.io/specter/) to sync documents.

- **Project:** %s
- **Config:** .specter (YAML)
- **Auth:** Set SPECTER_TOKEN environment variable

### Workflow

1. Run `+"`specter pull`"+` before starting work to get the latest documents.
2. Edit documents in the configured paths.
3. Run `+"`specter push`"+` after making changes to sync them.
4. Use `+"`specter status`"+` to check for local/remote differences.
5. Use `+"`specter diff`"+` to see line-level changes.
`, cfg.Project)
}

func installSkillBlock(repoRoot, block string) error {
	agentsPath := filepath.Join(repoRoot, "AGENTS.md")

	existing, err := os.ReadFile(agentsPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading AGENTS.md: %w", err)
	}

	content := string(existing)
	if len(content) > 0 && content[len(content)-1] != '\n' {
		content += "\n"
	}
	content += "\n" + block

	if err := os.WriteFile(agentsPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing AGENTS.md: %w", err)
	}

	fmt.Println("Specter agent instructions added to AGENTS.md")
	return nil
}

func init() {
	skillCmd.Flags().BoolVar(&skillInstall, "install", false, "Append instructions to AGENTS.md")
	rootCmd.AddCommand(skillCmd)
}
