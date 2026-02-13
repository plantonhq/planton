package gitrepo

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/plantonhq/openmcf/internal/cli/flag"
	"github.com/spf13/cobra"
)

const (
	// OpenMCFGitRepoEnvVar is the environment variable name for the local openmcf repo path
	OpenMCFGitRepoEnvVar = "OPENMCF_GIT_REPO"
)

// GetLocalRepoPath determines the openmcf repo location with priority:
// 1. --openmcf-git-repo flag if explicitly set
// 2. OPENMCF_GIT_REPO environment variable
// 3. Flag's default value
// Always returns a valid path string (expands ~ to home directory). Never returns an error.
func GetLocalRepoPath(cmd *cobra.Command) string {
	// Priority 1: Flag explicitly set by user
	if cmd.Flags().Changed(string(flag.OpenMCFGitRepo)) {
		val, _ := cmd.Flags().GetString(string(flag.OpenMCFGitRepo))
		return expandHomePath(val)
	}

	// Priority 2: Environment variable
	if envVal := os.Getenv(OpenMCFGitRepoEnvVar); envVal != "" {
		return expandHomePath(envVal)
	}

	// Priority 3: Flag's default value
	val, _ := cmd.Flags().GetString(string(flag.OpenMCFGitRepo))
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
