package module

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/ui/validateoutputs"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/eject"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/spf13/cobra"
)

var Eject = &cobra.Command{
	Use:   "eject <kind>",
	Short: "copy the official IaC module for a kind into your own directory",
	Long: `Ejects the official OpenTofu/Terraform or Pulumi module behind a catalog
kind into a directory you own, ready to customize.

The copy is fully yours: edit it freely, keep it in your own git repository,
and register it so your organization's deployments of the kind run your
module instead of the official one. A generated CONTRACT.md inside the copy
explains the input/outputs contract the module must keep honoring, and
'planton module verify' proves a customization still conforms.

For pulumi modules the copy becomes a standalone Go module: its go.mod is
generated (pinned to the release the module was ejected from) and its
self-imports are rewritten onto the module path you choose with --go-module.`,
	Example: `  # Eject the OpenTofu module for AWS S3 Bucket
  planton module eject AwsS3Bucket --provisioner tofu

  # Eject the Pulumi module under your own Go module path
  planton module eject AwsS3Bucket --provisioner pulumi \
    --go-module github.com/my-org/awss3bucket-module

  # Choose where the copy lands
  planton module eject GcpGkeCluster --provisioner terraform --output-dir ./modules/gke`,
	Args: cobra.ExactArgs(1),
	RunE: ejectHandler,
	// The handler renders rich errors itself; a returned error is a one-line
	// summary that must not be buried under usage output.
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	Eject.Flags().String("provisioner", "", "which module to eject: tofu, terraform, or pulumi")
	Eject.Flags().String("output-dir", "", "directory to receive the copy (default: ./<kind>-<provisioner>, must be new or empty)")
	Eject.Flags().String("go-module", "", "Go module path for an ejected pulumi module (default: github.com/your-org/<kind>-module)")
	Eject.Flags().Bool("skip-go-mod-tidy", false, "skip dependency resolution in the ejected pulumi module (offline use)")

	_ = Eject.MarkFlagRequired("provisioner")
}

func ejectHandler(cmd *cobra.Command, args []string) error {
	kindName := args[0]
	provisionerName, _ := cmd.Flags().GetString("provisioner")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	goModulePath, _ := cmd.Flags().GetString("go-module")
	skipGoModTidy, _ := cmd.Flags().GetBool("skip-go-mod-tidy")

	kind := catalogkindreflect.KindFromString(kindName)
	if kind == 0 {
		validateoutputs.RenderUnknownKind(kindName)
		return errors.Errorf("unknown catalog kind %q", kindName)
	}
	canonicalKindName := catalogkindreflect.ExtractKindNameByKind(kind)

	prov, err := provisioner.FromString(provisionerName)
	if err != nil {
		return err
	}

	if outputDir == "" {
		outputDir = fmt.Sprintf("./%s-%s", strings.ToLower(canonicalKindName), prov.String())
	}
	if goModulePath == "" {
		// An obvious placeholder: it compiles and deploys as-is, and the
		// generated contract notes tell the user how to rename it.
		goModulePath = fmt.Sprintf("github.com/your-org/%s-module", strings.ToLower(canonicalKindName))
	}

	result, err := eject.Eject(eject.Input{
		KindName:      canonicalKindName,
		Provisioner:   prov,
		OutputDir:     outputDir,
		GoModulePath:  goModulePath,
		SkipGoModTidy: skipGoModTidy,
	})
	if err != nil {
		cliprint.PrintError(fmt.Sprintf("Failed to eject the %s module for %s", prov.String(), canonicalKindName))
		return err
	}

	cliprint.PrintSuccess(fmt.Sprintf("Ejected the official %s module for %s (from %s)",
		prov.String(), result.KindName, result.SourceVersion))
	cliprint.PrintModulePath(result.OutputDir)

	if prov == provisioner.ProvisionerTypePulumi && !result.GoModTidyRan {
		cliprint.PrintInfo("Dependencies are not resolved yet — run 'go mod tidy' in the module directory before building")
	}

	fmt.Fprintln(os.Stdout)
	cliprint.PrintInfo(fmt.Sprintf("Start with %s — it explains the module's contract, how to verify changes, and how to register the module", eject.NotesFileName))

	return nil
}
