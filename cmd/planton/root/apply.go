package root

import (
	"os"

	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/internal/cli/iacflags"
	"github.com/plantonhq/planton/internal/cli/iacrunner"
	climanifest "github.com/plantonhq/planton/internal/cli/manifest"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/pkg/kustomize/builder"
	"github.com/plantonhq/planton/pkg/setdeploy"
	"github.com/plantonhq/planton/shared/iac/pulumi"
	"github.com/plantonhq/planton/shared/iac/terraform"
	"github.com/spf13/cobra"
)

var Apply = &cobra.Command{
	Use:   "apply",
	Short: "apply infrastructure changes using the provisioner specified in manifest",
	Long: `Apply infrastructure changes by automatically routing to the appropriate provisioner
(Pulumi, Tofu, or Terraform) based on the manifest label 'planton.dev/provisioner'.

If the provisioner label is not present, you will be prompted to select one interactively.

When the input holds MORE than one manifest — a multi-document file, a directory of
manifests, or a kustomize overlay that renders several resources — apply deploys the
whole set: a preflight report verifies everything verifiable first, then the resources
deploy sequentially in dependency order derived from their own references, each node's
outputs resolving the next nodes' valueFrom references. One manifest behaves exactly as
before.

Exit codes: 0 success, 1 deploy failure, 2 refused at preflight (nothing was handed to
an IaC engine).`,
	Example: `
	# Apply from clipboard (manifest content already copied)
	planton apply --clipboard
	planton apply -c

	# Apply with manifest file
	planton apply -f manifest.yaml
	planton apply --manifest manifest.yaml

	# Apply a directory of manifests as one set (dependency-ordered)
	planton apply -f manifests/

	# Apply a kustomize overlay (multi-resource overlays deploy as one set)
	planton apply --kustomize-dir _kustomize --overlay prod

	# Apply with IaC input file (extracts manifest from target field)
	planton apply -i iac-input.yaml

	# Apply with field overrides (single manifest only)
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
	// The set lane engages exactly when the input is plural; a single
	// document falls through to the existing path untouched.
	docs, isSet, err := resolveSetDocs(cmd)
	if err != nil {
		cliprint.PrintError(err.Error())
		os.Exit(1)
	}
	if isSet {
		os.Exit(iacrunner.RunSet(cmd, docs))
	}

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

// resolveSetDocs detects plural input without disturbing the single-manifest
// path: a -f DIRECTORY is always a set (kubectl's mental model — a set of one
// is still a set); a -f file or a rendered kustomize overlay is a set exactly
// when it holds more than one document. Clipboard and iac-input stay
// single-manifest sources — their loaders refuse plural content with the
// sentence that points here.
func resolveSetDocs(cmd *cobra.Command) ([]setdeploy.Doc, bool, error) {
	// Clipboard and iac-input take priority in the single lane's source
	// ladder; when either is chosen, the set lane stays out of the way.
	if clipboard, _ := cmd.Flags().GetBool(string(flag.Clipboard)); clipboard {
		return nil, false, nil
	}
	if iacInput, _ := cmd.Flags().GetString(string(flag.IacInput)); iacInput != "" {
		return nil, false, nil
	}

	if manifestPath, _ := cmd.Flags().GetString(string(flag.Manifest)); manifestPath != "" {
		info, err := os.Stat(manifestPath)
		if err == nil && info.IsDir() {
			docs, err := setdeploy.CollectDocsFromDir(manifestPath)
			if err != nil {
				return nil, false, err
			}
			return docs, true, nil
		}
		if err != nil {
			// Not stat-able (a URL, or a missing file the single lane will
			// diagnose): not the set lane's call to make.
			return nil, false, nil
		}
		b, err := os.ReadFile(manifestPath)
		if err != nil {
			return nil, false, nil
		}
		docs, err := setdeploy.CollectDocsFromBytes(b, manifestPath)
		if err != nil || len(docs) <= 1 {
			// Malformed YAML gets the single lane's richer diagnosis.
			return nil, false, nil
		}
		return docs, true, nil
	}

	kustomizeDir, _ := cmd.Flags().GetString(string(flag.KustomizeDir))
	overlay, _ := cmd.Flags().GetString(string(flag.Overlay))
	if kustomizeDir != "" && overlay != "" {
		tempPath, err := builder.BuildManifest(kustomizeDir, overlay)
		if err != nil {
			return nil, false, err
		}
		defer os.Remove(tempPath)
		b, err := os.ReadFile(tempPath)
		if err != nil {
			return nil, false, err
		}
		// The doc source label points at the overlay the author edits, not
		// the temp file that no longer exists when the report prints.
		docs, err := setdeploy.CollectDocsFromBytes(b, kustomizeDir+"/overlays/"+overlay)
		if err != nil {
			return nil, false, err
		}
		if len(docs) <= 1 {
			// One rendered document: the single lane re-renders and proceeds
			// exactly as before.
			return nil, false, nil
		}
		return docs, true, nil
	}

	return nil, false, nil
}
