package root

import (
	"os"

	"github.com/plantonhq/planton/apis/dev/planton/shared/iac/pulumi"
	"github.com/plantonhq/planton/apis/dev/planton/shared/iac/terraform"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/iacflags"
	"github.com/plantonhq/planton/internal/cli/iacrunner"
	climanifest "github.com/plantonhq/planton/internal/cli/manifest"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/spf13/cobra"
)

var Destroy = &cobra.Command{
	Use:     "destroy",
	Aliases: []string{"delete"},
	Short:   "destroy infrastructure using the provisioner specified in manifest",
	Long: `Destroy infrastructure by automatically routing to the appropriate provisioner
(Pulumi, Tofu, or Terraform) based on the manifest label 'planton.dev/provisioner'.

If the provisioner label is not present, you will be prompted to select one interactively.

This command has 'delete' as an alias for kubectl-like experience.`,
	Example: `
	# Destroy with manifest file
	planton destroy -f manifest.yaml
	planton delete -f manifest.yaml
	planton destroy --manifest manifest.yaml

	# Destroy with stack input file (extracts manifest from target field)
	planton destroy -i stack-input.yaml

	# Destroy with kustomize
	planton destroy --kustomize-dir _kustomize --overlay prod

	# Destroy with field overrides
	planton destroy -f manifest.yaml --set spec.version=v1.2.3
	`,
	Run: destroyHandler,
}

func init() {
	iacflags.AddManifestSourceFlags(Destroy)
	iacflags.AddProviderConfigFlags(Destroy)
	iacflags.AddExecutionFlags(Destroy)
	iacflags.AddPulumiFlags(Destroy)
	iacflags.AddTofuApplyFlags(Destroy)
	iacflags.AddTofuInitFlags(Destroy)
}

func destroyHandler(cmd *cobra.Command, args []string) {
	ctx, err := iacrunner.ResolveContext(cmd)
	if err != nil {
		// Only print error if it wasn't already handled (clipboard/manifest load errors are pre-handled)
		if !climanifest.IsClipboardError(err) && !manifest.IsManifestLoadError(err) {
			cliprint.PrintError(err.Error())
		}
		os.Exit(1)
	}
	defer ctx.Cleanup()

	switch ctx.ProvisionerType {
	case provisioner.ProvisionerTypePulumi:
		if err := iacrunner.RunPulumi(ctx, cmd, pulumi.PulumiOperationType_destroy, false); err != nil {
			os.Exit(1)
		}
	case provisioner.ProvisionerTypeTofu:
		if err := iacrunner.RunTofu(ctx, cmd, terraform.TerraformOperationType_destroy); err != nil {
			os.Exit(1)
		}
	case provisioner.ProvisionerTypeTerraform:
		if err := iacrunner.RunTerraform(ctx, cmd, terraform.TerraformOperationType_destroy); err != nil {
			os.Exit(1)
		}
	default:
		cliprint.PrintError("Unknown provisioner type")
		os.Exit(1)
	}
}
