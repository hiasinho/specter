package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// CurrentBranch returns the current git branch name.
func CurrentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("could not detect git branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
