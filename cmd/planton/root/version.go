package root

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/plantonhq/planton/internal/cli/upgrade"
	"github.com/plantonhq/planton/internal/cli/version"
	"github.com/spf13/cobra"
)

var Version = &cobra.Command{
	Use:   "version",
	Short: "check the version of the cli",
	Run:   versionHandler,
}

func versionHandler(cmd *cobra.Command, args []string) {
	PrintVersion()
}

// PrintVersion prints a colorful version display with update check
func PrintVersion() {
	currentVersion := version.Version
	if currentVersion == "" {
		currentVersion = version.DefaultVersion
	}

	// Fetch latest version (silently fail if network unavailable)
	latestVersion, err := upgrade.GetLatestVersion()

	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen, color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()

	fmt.Println()

	// Check if update is available
	if err == nil && upgrade.CompareVersions(currentVersion, latestVersion) {
		// Update available - show yellow current, green latest
		fmt.Printf("Current version: %s\n", yellow(currentVersion))
		fmt.Printf("Latest version:  %s\n", green(latestVersion))
		fmt.Println()
		orange := color.New(color.FgYellow, color.Bold).SprintFunc()
		fmt.Printf("%s A new version is available!\n", orange("⚡"))
		fmt.Println()
		fmt.Printf("Run %s to update.\n", cyan("planton upgrade"))
	} else {
		// Up to date or couldn't check - show green current version
		fmt.Printf("Current version: %s\n", green(currentVersion))
		if err == nil {
			fmt.Printf("Latest version:  %s\n", green(latestVersion))
			fmt.Println()
			fmt.Printf("%s You're up to date!\n", green("✔"))
		}
	}
	fmt.Println()
}
