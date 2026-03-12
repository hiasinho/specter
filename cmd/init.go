package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/hiasinho/specter/internal/api"
	"github.com/hiasinho/specter/internal/config"
	"github.com/spf13/cobra"
)

var (
	initName    string
	initProject string
	initPaths   []string
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Specter project",
	Long: `Creates a .specter config file and registers the project on the Specter service.

Defaults:
  --project   current directory name
  --name      project slug, title-cased
  --path      specs/

The owner is always your Specter username.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := os.Getenv("SPECTER_TOKEN")
		if token == "" {
			return fmt.Errorf("SPECTER_TOKEN environment variable not set")
		}

		repoRoot, err := os.Getwd()
		if err != nil {
			return err
		}

		// Check if .specter already exists
		specterPath := filepath.Join(repoRoot, ".specter")
		if _, err := os.Stat(specterPath); err == nil {
			return fmt.Errorf(".specter config already exists in %s", repoRoot)
		}

		// Fetch user for owner
		client := api.NewClient(token)
		user, err := client.Me()
		if err != nil {
			return fmt.Errorf("could not fetch user info: %w", err)
		}

		// Resolve project slug
		slug := initProject
		if slug == "" {
			slug = filepath.Base(repoRoot)
		}
		project := user.Username + "/" + slug

		if err := config.ValidateProject(project); err != nil {
			return err
		}

		// Resolve display name
		name := initName
		if name == "" {
			name = titleCase(slug)
		}

		// Default paths
		paths := initPaths
		if len(paths) == 0 {
			paths = []string{"specs/"}
		}

		// Write .specter config
		cfgContent := "project: " + project + "\npaths:\n"
		for _, p := range paths {
			cfgContent += "  - " + p + "\n"
		}
		if err := os.WriteFile(specterPath, []byte(cfgContent), 0644); err != nil {
			return fmt.Errorf("could not write .specter config: %w", err)
		}

		// Register project on service
		proj, err := client.CreateProject(slug, name)
		if err != nil {
			os.Remove(specterPath)
			return err
		}

		fmt.Printf("Initialized Specter project: %s (%s)\n", proj.Name, proj.FullName)
		fmt.Printf("Config written to %s\n", specterPath)
		return nil
	},
}

// titleCase converts a slug like "my-cool-project" to "My Cool Project".
func titleCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || unicode.IsSpace(r)
	})
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func init() {
	initCmd.Flags().StringVar(&initName, "name", "", "Display name (default: slug, title-cased)")
	initCmd.Flags().StringVar(&initProject, "project", "", "Project slug (default: directory name)")
	initCmd.Flags().StringSliceVar(&initPaths, "path", nil, `Paths to scan for documents (default: "specs/")`)
	rootCmd.AddCommand(initCmd)
}
