// Package openmcf provides the root command for the OpenMCF CLI.
// Auto-release test: CLI change triggers v{semver}.{YYYYMMDD}.{N} tag format.
package openmcf

import (
	"fmt"
	"os"

	"github.com/plantonhq/openmcf/cmd/openmcf/root"
	"github.com/plantonhq/openmcf/internal/cli/flag"
	"github.com/spf13/cobra"
)

// DefaultOpenMCFGitRepo is the default path for the local openmcf git repository
const DefaultOpenMCFGitRepo = "~/scm/github.com/plantonhq/openmcf"

var rootCmd = &cobra.Command{
	Use:   "openmcf",
	Short: "Unified Interface for Multi-Cloud Infrastructure",
	Run: func(cmd *cobra.Command, args []string) {
		// Check if version flag was passed
		showVersion, _ := cmd.Flags().GetBool("version")
		if showVersion {
			root.PrintVersion()
			return
		}
		// Otherwise show help
		cmd.Help()
	},
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.DisableSuggestions = true

	// Enable -v as shorthand for --version (handled in Run function for colorful output)
	rootCmd.Flags().BoolP("version", "v", false, "show version information")

	// Local module flags - inherited by all subcommands
	rootCmd.PersistentFlags().Bool(string(flag.LocalModule), false,
		"Use local openmcf git repository for IaC modules instead of downloading")
	rootCmd.PersistentFlags().String(string(flag.OpenMCFGitRepo), DefaultOpenMCFGitRepo,
		"Path to local openmcf git repository (used with --local-module)")

	rootCmd.AddCommand(
		root.Apply,
		root.Checkout,
		root.Destroy,
		root.Downgrade,
		root.Init,
		root.LoadManifest,
		root.ModulesVersion,
		root.Plan,
		root.Pull,
		root.Pulumi,
		root.Refresh,
		root.Terraform,
		root.Tofu,
		root.Upgrade,
		root.ValidateManifest,
		root.Version,
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
