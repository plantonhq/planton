package tofu

import (
	"fmt"
	"github.com/plantonhq/planton/internal/cli/ui"
	"os"
	"path/filepath"
	"sort"

	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var GenerateModule = &cobra.Command{
	Use:   "generate-module <kind>",
	Short: "Generate the full thin Terraform module for a Kubernetes-CRD-projection kind",
	Long: `The "generate-module" command emits the complete iac/tf Terraform module
(variables.tf, backend.tf, locals.tf, main.tf, provider.tf, outputs.tf) for a
kind whose spec is a direct projection of a single Kubernetes
custom resource -- i.e. a kind annotated with kubernetes_manifest_projection in
CatalogKindMeta (Istio, Gateway API, etc.).

The module is a thin kubectl_manifest (alekc/kubectl) passthrough: variable
"spec" is typed 'any' and handed to the CR verbatim, because the proto->tfvars
converter already emits the manifest-shaped (camelCase, null-pruned) spec with
StringValueOrRef foreign keys resolved to literal strings. kubectl_manifest
needs no cluster connection at plan time, so the CR can be planned before its
CRDs exist (single-run infra charts, offline plan proofs). Kinds without the
projection annotation are rejected (use the standard 'generate-variables' +
provider-resource module pattern instead).`,
	Example: `
  # Write the module into the kind's iac/tf directory
  planton tofu generate-module KubernetesDestinationRule \
    --output-dir catalog/kubernetes/kubernetesdestinationrule/iac/tf
`,
	Args: cobra.ExactArgs(1),
	Run:  generateModuleHandler,
}

func init() {
	GenerateModule.Flags().String(string(flag.OutputDir), "", "output directory (the kind's iac/tf dir); required")
}

func generateModuleHandler(cmd *cobra.Command, args []string) {
	kindName := args[0]

	outputDir, err := cmd.Flags().GetString(string(flag.OutputDir))
	flag.Require(err, flag.OutputDir, outputDir, "--output-dir catalog/<provider>/<kind>/iac/tf")

	catalogKind := catalogkindreflect.KindFromString(kindName)
	manifestObject := catalogkindreflect.ToMessageMap[catalogKind]
	if manifestObject == nil {
		ui.Failure(
			fmt.Sprintf("no spec message is registered for kind %s", catalogKind.String()),
			"the kind exists in the catalog enum but its proto package is not linked into this binary",
			"run `make generate-catalog-kind-map` and rebuild, then retry",
		)
	}

	files, err := generators.GenerateManifestModule(catalogKind, manifestObject)
	if err != nil {
		ui.Failure(
			fmt.Sprintf("the Terraform module for %s could not be generated: %v", kindName, err),
			"the kind's spec message could not be walked into a module skeleton",
			"report it at https://github.com/plantonhq/planton/issues naming the kind",
		)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		ui.Failure(
			fmt.Sprintf("the output directory %s could not be created: %v", outputDir, err),
			"the path is not writable, or a file already sits where the directory should be",
			"point --output-dir at a writable directory path",
		)
	}

	// Deterministic write order for stable logs.
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(outputDir, name)
		if err := os.WriteFile(path, []byte(files[name]), 0644); err != nil {
			ui.Failure(
				fmt.Sprintf("the generated file could not be written to %s: %v", path, err),
				"the output directory is not writable",
				"point --output-dir at a writable directory path",
			)
		}
		log.Infof("wrote %s", path)
	}
}
