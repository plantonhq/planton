package root

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/staging"
	"github.com/plantonhq/planton/internal/cli/version"
	"github.com/plantonhq/planton/pkg/downloads"
	"github.com/spf13/cobra"
)

var ModulesVersion = &cobra.Command{
	Use:   "modules-version",
	Short: "Show which release this binary downloads IaC modules from, and the staging area's version",
	Long: `Display the catalog release this binary is pinned to for IaC module
downloads, and the version currently checked out in the local staging area.

Every apply, plan, and destroy downloads the published module for a kind from
the pinned release (downloads.planton.ai/releases/<release>/...) so the module
always matches the schemas compiled into this binary. A binary with no release
pin (a development build) falls back to the staging area instead.

The staging area (~/.planton/staging/planton) maintains a cached copy
of the Planton repository containing all IaC modules (Pulumi and Terraform/OpenTofu).
This command reads the version from the staging area's .version file and displays it.
If the staging area doesn't exist, it will indicate that no modules are cached yet.

Use 'planton checkout <version>' to switch to a different version.
Use 'planton pull' to update to the latest version from upstream.`,
	Example: `  # Check current modules version
  planton modules-version

  # Typical workflow
  planton modules-version     # Check current version
  planton checkout v0.2.273   # Switch to specific version
  planton modules-version     # Verify the switch`,
	Run: modulesVersionHandler,
}

func modulesVersionHandler(cmd *cobra.Command, args []string) {
	printPinnedRelease()

	exists, version, repoPath, err := staging.GetStagingInfo()
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to get staging info: %v", err))
		os.Exit(1)
	}

	if !exists {
		fmt.Println("No IaC modules cached yet.")
		fmt.Println("")
		fmt.Println("Run 'planton pull' to clone the modules to the staging area,")
		fmt.Println("or run any apply/preview/destroy command to automatically set up staging.")
		return
	}

	fmt.Println("IaC Modules Staging Area")
	fmt.Println("========================")
	fmt.Printf("Location: %s\n", repoPath)
	if version != "" {
		fmt.Printf("Version:  %s\n", version)
	} else {
		fmt.Println("Version:  (unknown - .version file not found)")
	}
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  planton pull                  Update to latest from upstream")
	fmt.Println("  planton checkout <version>    Switch to a specific version")
}

// printPinnedRelease names the release the module runtimes download from, or
// says plainly that this binary has none. The pin is the binary's own
// version for the standalone CLI and the embedded catalog release for a host
// binary; either way it is the one fact that decides where modules come from.
func printPinnedRelease() {
	fmt.Println("IaC Modules Release Pin")
	fmt.Println("=======================")
	if version.Version == "" || version.Version == version.DefaultVersion {
		fmt.Println("Release:  (none -- a development build; modules come from the staging area below)")
	} else {
		fmt.Printf("Release:  %s\n", version.Version)
		fmt.Printf("Source:   %s/%s/modules/{terraform,pulumi}/<kind>/\n", downloads.BaseURL, version.Version)
	}
	fmt.Println("")
}
