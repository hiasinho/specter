package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hiasinho/specter/internal/config"
	"github.com/spf13/cobra"
)

//go:embed skill.md
var skillTemplate string

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
	tmpl, err := template.New("skill").Parse(skillTemplate)
	if err != nil {
		return fmt.Sprintf("error parsing skill template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return fmt.Sprintf("error executing skill template: %v", err)
	}

	return buf.String()
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
