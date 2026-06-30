package root

import (
	"github.com/plantonhq/planton/internal/cli/upgrade"
	"github.com/spf13/cobra"
)

var Upgrade = &cobra.Command{
	Use:   "upgrade [VERSION]",
	Short: "upgrade the planton CLI to the latest or specified version",
	Long: `Upgrade the planton CLI to the latest available version, or to a specific version if provided.

On macOS, if planton was installed via Homebrew, this command uses 'brew upgrade --cask'.
On all other platforms (or if Homebrew is not available), it downloads the latest
binary directly from GitHub releases.

When a specific VERSION is provided, the CLI is downloaded directly from GitHub releases,
bypassing Homebrew. If Homebrew manages the current installation, you will be
prompted to uninstall via Homebrew first to avoid conflicts.

Arguments:
  VERSION   Optional. Specific version to install (e.g., v0.3.10-cli.20260110.0).
            If omitted, upgrades to the latest version.

Examples:
  # Upgrade to the latest version
  planton upgrade

  # Upgrade to a specific version
  planton upgrade v0.3.10-cli.20260110.0

  # Check for updates without installing
  planton upgrade --check

  # Force upgrade even if already on latest version
  planton upgrade --force

  # Force install a specific version
  planton upgrade v0.3.10-cli.20260110.0 --force`,
	Args: cobra.MaximumNArgs(1),
	Run:  upgradeHandler,
}

func init() {
	Upgrade.Flags().BoolP("check", "c", false, "check for updates without installing")
	Upgrade.Flags().BoolP("force", "f", false, "force upgrade even if already on latest/specified version")
}

func upgradeHandler(cmd *cobra.Command, args []string) {
	checkOnly, _ := cmd.Flags().GetBool("check")
	force, _ := cmd.Flags().GetBool("force")

	var targetVersion string
	if len(args) > 0 {
		targetVersion = args[0]
	}

	upgrade.Run(checkOnly, force, targetVersion)
}
