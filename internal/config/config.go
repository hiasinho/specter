package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Project string   `yaml:"project"`
	Paths   []string `yaml:"paths"`
	Exclude []string `yaml:"exclude"`
}

// Load reads the .specter config file from the given directory.
func Load(dir string) (*Config, error) {
	path := filepath.Join(dir, ".specter")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read .specter config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("could not parse .specter config: %w", err)
	}

	if cfg.Project == "" {
		return nil, fmt.Errorf(".specter config missing required field: project")
	}

	parts := strings.SplitN(cfg.Project, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf(".specter config: project must be in owner/slug format (e.g. hiasinho/specter)")
	}

	if len(cfg.Paths) == 0 {
		return nil, fmt.Errorf(".specter config missing required field: paths")
	}

	return &cfg, nil
}

// ProjectOwner returns the owner part of the project identifier.
func (c *Config) ProjectOwner() string {
	return strings.SplitN(c.Project, "/", 2)[0]
}

// ProjectSlug returns the slug part of the project identifier.
func (c *Config) ProjectSlug() string {
	return strings.SplitN(c.Project, "/", 2)[1]
}

// FindRepoRoot walks up from dir looking for a .specter file.
func FindRepoRoot(dir string) (string, error) {
	for {
		if _, err := os.Stat(filepath.Join(dir, ".specter")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .specter config found")
		}
		dir = parent
	}
}
