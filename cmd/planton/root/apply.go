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

var Apply = &cobra.Command{
	Use:   "apply",
	Short: "apply infrastructure changes using the provisioner specified in manifest",
	Long: `Apply infrastructure changes by automatically routing to the appropriate provisioner
(Pulumi, Tofu, or Terraform) based on the manifest label 'planton.dev/provisioner'.

If the provisioner label is not present, you will be prompted to select one interactively.`,
	Example: `
	# Apply from clipboard (manifest content already copied)
	planton apply --clipboard
	planton apply -c

	# Apply with manifest file
	planton apply -f manifest.yaml
	planton apply --manifest manifest.yaml

	# Apply with stack input file (extracts manifest from target field)
	planton apply -i stack-input.yaml

	# Apply with kustomize
	planton apply --kustomize-dir _kustomize --overlay prod

	# Apply with field overrides
	planton apply -f manifest.yaml --set spec.version=v1.2.3
	`,
	Run: applyHandler,
}

func init() {
	iacflags.AddManifestSourceFlags(Apply)
	iacflags.AddProviderConfigFlags(Apply)
	iacflags.AddExecutionFlags(Apply)
	iacflags.AddPulumiFlags(Apply)
	iacflags.AddTofuApplyFlags(Apply)
	iacflags.AddTofuInitFlags(Apply)
}

func applyHandler(cmd *cobra.Command, args []string) {
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
		if err := iacrunner.RunPulumi(ctx, cmd, pulumi.PulumiOperationType_update, false); err != nil {
			os.Exit(1)
		}
	case provisioner.ProvisionerTypeTofu:
		if err := iacrunner.RunTofu(ctx, cmd, terraform.TerraformOperationType_apply); err != nil {
			os.Exit(1)
		}
	case provisioner.ProvisionerTypeTerraform:
		if err := iacrunner.RunTerraform(ctx, cmd, terraform.TerraformOperationType_apply); err != nil {
			os.Exit(1)
		}
	default:
		cliprint.PrintError("Unknown provisioner type")
		os.Exit(1)
	}
}
