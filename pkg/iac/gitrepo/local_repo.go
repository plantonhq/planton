package gitrepo

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/spf13/cobra"
)

const (
	// PlantonGitRepoEnvVar is the environment variable name for the local planton repo path
	PlantonGitRepoEnvVar = "PLANTON_GIT_REPO"
)

// GetLocalRepoPath determines the planton repo location with priority:
// 1. --planton-git-repo flag if explicitly set
// 2. PLANTON_GIT_REPO environment variable
// 3. Flag's default value
// Always returns a valid path string (expands ~ to home directory). Never returns an error.
func GetLocalRepoPath(cmd *cobra.Command) string {
	// Priority 1: Flag explicitly set by user
	if cmd.Flags().Changed(string(flag.PlantonGitRepo)) {
		val, _ := cmd.Flags().GetString(string(flag.PlantonGitRepo))
		return expandHomePath(val)
	}

	// Priority 2: Environment variable
	if envVal := os.Getenv(PlantonGitRepoEnvVar); envVal != "" {
		return expandHomePath(envVal)
	}

	// Priority 3: Flag's default value
	val, _ := cmd.Flags().GetString(string(flag.PlantonGitRepo))
	return expandHomePath(val)
}

// expandHomePath expands ~ to the user's home directory
func expandHomePath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
