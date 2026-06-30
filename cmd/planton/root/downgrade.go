package root

import (
	"github.com/plantonhq/planton/internal/cli/upgrade"
	"github.com/spf13/cobra"
)

var Downgrade = &cobra.Command{
	Use:   "downgrade VERSION",
	Short: "install a previous version of the planton CLI",
	Long: `Install a specific previous version of the planton CLI.

This command downloads the specified version directly from GitHub releases.
If Homebrew manages the current installation, you will be prompted to uninstall
via Homebrew first to avoid conflicts.

Arguments:
  VERSION   Required. The version to install (e.g., v0.3.5-cli.20260108.0).

Examples:
  # Downgrade to a specific version
  planton downgrade v0.3.5-cli.20260108.0

  # Force downgrade even if already on the specified version
  planton downgrade v0.3.5-cli.20260108.0 --force`,
	Args: cobra.ExactArgs(1),
	Run:  downgradeHandler,
}

func init() {
	Downgrade.Flags().BoolP("force", "f", false, "force install even if already on the specified version")
}

func downgradeHandler(cmd *cobra.Command, args []string) {
	force, _ := cmd.Flags().GetBool("force")
	targetVersion := args[0]

	// Downgrade is just an alias for upgrade with a specific version
	// checkOnly is always false for downgrade
	upgrade.Run(false, force, targetVersion)
}
