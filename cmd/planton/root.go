// Package planton provides the root command for the Planton CLI.
// Auto-release test: CLI change triggers v{semver}.{YYYYMMDD}.{N} tag format.
package planton

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/cmd/planton/root"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/spf13/cobra"
)

// DefaultPlantonGitRepo is the default path for the local planton git repository
const DefaultPlantonGitRepo = "~/scm/github.com/plantonhq/planton"

var rootCmd = &cobra.Command{
	Use:   "planton",
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
	rootCmd.DisableSuggestions = true

	// Enable -v as shorthand for --version (handled in Run function for colorful output)
	rootCmd.Flags().BoolP("version", "v", false, "show version information")

	// Local module flags - inherited by all subcommands
	rootCmd.PersistentFlags().Bool(string(flag.LocalModule), false,
		"Use local planton git repository for IaC modules instead of downloading")
	rootCmd.PersistentFlags().String(string(flag.PlantonGitRepo), DefaultPlantonGitRepo,
		"Path to local planton git repository (used with --local-module)")

	rootCmd.AddCommand(
		root.Apply,
		root.Checkout,
		root.Destroy,
		root.Downgrade,
		root.E2E,
		root.Init,
		root.Kustomize,
		root.LoadManifest,
		root.ModulesVersion,
		root.Plan,
		root.Pull,
		root.Pulumi,
		root.Refresh,
		root.SecretCoverage,
		root.Terraform,
		root.Tofu,
		root.Upgrade,
		root.ValidateManifest,
		root.ValidateOutputs,
		root.ValidateRefs,
		root.Version,
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
