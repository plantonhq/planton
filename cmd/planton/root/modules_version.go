package root

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/staging"
	"github.com/spf13/cobra"
)

var ModulesVersion = &cobra.Command{
	Use:   "modules-version",
	Short: "Show the current version of IaC modules in the staging area",
	Long: `Display the currently checked out version of the Planton IaC modules
in the local staging area.

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
