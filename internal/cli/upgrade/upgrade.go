package upgrade

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/version"
)

// Run executes the upgrade command
// If targetVersion is empty, upgrades to the latest version
// If targetVersion is specified, installs that specific version
func Run(checkOnly bool, force bool, targetVersion string) {
	currentVersion := version.Version
	if currentVersion == "" {
		currentVersion = "dev"
	}

	// If a specific version is requested, use the version-specific flow
	if targetVersion != "" {
		runWithTargetVersion(currentVersion, targetVersion, force)
		return
	}

	// Original flow: upgrade to latest
	runUpgradeToLatest(currentVersion, checkOnly, force)
}

// runUpgradeToLatest handles the original upgrade flow (to latest version)
func runUpgradeToLatest(currentVersion string, checkOnly bool, force bool) {
	// Step 1: Check for latest version
	cliprint.PrintStep("Checking for updates...")

	latestVersion, err := GetLatestVersion()
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to check for updates: %v", err))
		fmt.Println()
		fmt.Println("You can manually download the latest version from:")
		fmt.Println("  https://github.com/plantonhq/planton/releases")
		os.Exit(1)
	}

	// Step 2: Compare versions
	needsUpgrade := CompareVersions(currentVersion, latestVersion)

	fmt.Println()
	if needsUpgrade {
		// Show versions with color distinction when update is available
		yellow := color.New(color.FgYellow).SprintFunc()
		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		fmt.Printf("Current version: %s\n", yellow(currentVersion))
		fmt.Printf("Latest version:  %s\n", green(latestVersion))
	} else {
		// Show versions in matching green when up-to-date
		green := color.New(color.FgGreen, color.Bold).SprintFunc()
		fmt.Printf("Current version: %s\n", green(currentVersion))
		fmt.Printf("Latest version:  %s\n", green(latestVersion))
	}

	if !needsUpgrade && !force {
		fmt.Println()
		cliprint.PrintSuccess(fmt.Sprintf("planton is already up to date (%s)", currentVersion))
		return
	}

	if checkOnly {
		if needsUpgrade {
			fmt.Println()
			orange := color.New(color.FgYellow, color.Bold).SprintFunc()
			fmt.Printf("%s A new version is available!\n", orange("⚡"))
			fmt.Println()
			blue := color.New(color.FgCyan, color.Bold).SprintFunc()
			fmt.Printf("Run %s to update.\n", blue("planton upgrade"))
		}
		return
	}

	if !needsUpgrade && force {
		fmt.Println()
		cliprint.PrintStep("Forcing upgrade...")
	}

	// Step 3: Download and install directly from GitHub releases
	fmt.Println()
	cliprint.PrintStep("Upgrade method: Direct Download")

	if err := UpgradeViaDirect(latestVersion); err != nil {
		handleUpgradeError(err, latestVersion)
		os.Exit(1)
	}

	// Step 4: Success message
	fmt.Println()
	cliprint.PrintSuccess(fmt.Sprintf("Successfully upgraded to %s", latestVersion))

	fmt.Println()
	cliprint.PrintStep("Note: You may need to restart your terminal for changes to take effect.")
}

// runWithTargetVersion handles installing a specific version
func runWithTargetVersion(currentVersion, targetVersion string, force bool) {
	cliprint.PrintStep("Checking for updates...")

	// Step 1: Validate the target version exists
	normalizedVersion, err := ValidateVersion(targetVersion)
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to validate version: %v", err))
		fmt.Println()
		fmt.Println("You can view available versions at:")
		fmt.Println("  https://github.com/plantonhq/planton/releases")
		os.Exit(1)
	}

	// Step 2: Show current and target versions
	fmt.Println()
	cyan := color.New(color.FgCyan, color.Bold).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Printf("Current version: %s\n", yellow(currentVersion))
	fmt.Printf("Target version:  %s\n", cyan(normalizedVersion))

	// Step 3: Check if already on target version
	if currentVersion == normalizedVersion && !force {
		fmt.Println()
		cliprint.PrintSuccess(fmt.Sprintf("planton is already at version %s", normalizedVersion))
		return
	}

	if currentVersion == normalizedVersion && force {
		fmt.Println()
		cliprint.PrintStep("Forcing reinstall...")
	}

	// Step 4: Direct download and install
	fmt.Println()
	cliprint.PrintStep("Upgrade method: Direct Download")

	if err := UpgradeViaDirect(normalizedVersion); err != nil {
		handleUpgradeError(err, normalizedVersion)
		os.Exit(1)
	}

	// Step 5: Success message
	fmt.Println()
	cliprint.PrintSuccess(fmt.Sprintf("Successfully installed %s", normalizedVersion))

	fmt.Println()
	cliprint.PrintStep("Note: You may need to restart your terminal for changes to take effect.")
}

// handleUpgradeError handles and displays upgrade errors with helpful suggestions
func handleUpgradeError(err error, latestVersion string) {
	fmt.Println()

	// Check for permission errors
	var permErr *PermissionError
	if errors.As(err, &permErr) {
		cliprint.PrintError(permErr.Error())
		fmt.Println()
		fmt.Println("Try running with sudo:")
		fmt.Println("  sudo planton upgrade")
		fmt.Println()
		fmt.Println("Or download manually to a user directory:")
		goos, goarch := GetPlatformInfo()
		downloadURL := BuildDownloadURL(latestVersion, goos, goarch)
		fmt.Printf("  curl -LO %s\n", downloadURL)
		if goos == "windows" {
			fmt.Println("  # Extract the zip file and move planton.exe to your PATH")
		} else {
			fmt.Println("  tar -xzf cli_*.tar.gz")
			fmt.Println("  chmod +x planton")
			fmt.Println("  mv planton ~/.local/bin/")
		}
		return
	}

	// Generic error
	cliprint.PrintError(fmt.Sprintf("Upgrade failed: %v", err))
	fmt.Println()
	fmt.Println("You can manually download the latest version from:")
	fmt.Println("  https://github.com/plantonhq/planton/releases")

	// Show platform-specific instructions
	goos, goarch := GetPlatformInfo()
	downloadURL := BuildDownloadURL(latestVersion, goos, goarch)

	fmt.Println()
	fmt.Println("Or download directly:")
	if runtime.GOOS == "windows" {
		fmt.Printf("  Invoke-WebRequest -Uri \"%s\" -OutFile \"cli.zip\"\n", downloadURL)
		fmt.Println("  Expand-Archive -Path \"cli.zip\" -DestinationPath \".\"")
		fmt.Println("  Move-Item -Path \"planton.exe\" -Destination \"C:\\Windows\\System32\\\"")
	} else {
		fmt.Printf("  curl -LO %s\n", downloadURL)
		fmt.Println("  tar -xzf cli_*.tar.gz")
		fmt.Println("  chmod +x planton")
		if runtime.GOOS == "darwin" {
			fmt.Println("  xattr -dr com.apple.quarantine planton  # Remove macOS quarantine")
		}
		fmt.Println("  sudo mv planton /usr/local/bin/")
	}
}
