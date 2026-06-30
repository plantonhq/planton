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

var Plan = &cobra.Command{
	Use:     "plan",
	Aliases: []string{"preview"},
	Short:   "preview infrastructure changes using the provisioner specified in manifest",
	Long: `Preview infrastructure changes by automatically routing to the appropriate provisioner
(Pulumi, Tofu, or Terraform) based on the manifest label 'planton.dev/provisioner'.

If the provisioner label is not present, you will be prompted to select one interactively.

This command has 'preview' as an alias for Pulumi-style experience.`,
	Example: `
	# Preview changes with manifest file
	planton plan -f manifest.yaml
	planton preview -f manifest.yaml
	planton plan --manifest manifest.yaml

	# Preview with stack input file (extracts manifest from target field)
	planton plan -i stack-input.yaml

	# Preview with kustomize
	planton plan --kustomize-dir _kustomize --overlay prod

	# Preview with field overrides
	planton plan -f manifest.yaml --set spec.version=v1.2.3

	# Preview destroy plan (Tofu)
	planton plan -f manifest.yaml --destroy
	`,
	Run: planHandler,
}

func init() {
	iacflags.AddManifestSourceFlags(Plan)
	iacflags.AddProviderConfigFlags(Plan)
	iacflags.AddExecutionFlags(Plan)
	iacflags.AddPulumiFlags(Plan)
	iacflags.AddTofuPlanFlags(Plan)
	iacflags.AddTofuInitFlags(Plan)
}

func planHandler(cmd *cobra.Command, args []string) {
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
		// For preview, we use update operation with isPreview=true
		if err := iacrunner.RunPulumi(ctx, cmd, pulumi.PulumiOperationType_update, true); err != nil {
			os.Exit(1)
		}
	case provisioner.ProvisionerTypeTofu:
		if err := iacrunner.RunTofu(ctx, cmd, terraform.TerraformOperationType_plan); err != nil {
			os.Exit(1)
		}
	case provisioner.ProvisionerTypeTerraform:
		if err := iacrunner.RunTerraform(ctx, cmd, terraform.TerraformOperationType_plan); err != nil {
			os.Exit(1)
		}
	default:
		cliprint.PrintError("Unknown provisioner type")
		os.Exit(1)
	}
}
