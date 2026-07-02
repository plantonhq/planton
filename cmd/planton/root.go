// Package planton provides the root command for the Planton CLI.
package planton

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/cmd/planton/root"
	"github.com/spf13/cobra"
)

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

	// The engine command set and its persistent flags come from the shared
	// embedding seam; the standalone binary owns its self-management commands
	// and developer tools on top of it.
	root.RegisterCommands(rootCmd, root.Options{})
	rootCmd.AddCommand(
		root.Downgrade,
		root.E2E,
		root.Upgrade,
		root.Version,
	)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
