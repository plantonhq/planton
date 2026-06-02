package tofu

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/plantonhq/openmcf/internal/cli/flag"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
	"github.com/plantonhq/openmcf/pkg/iac/tofu/generators"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var GenerateModule = &cobra.Command{
	Use:   "generate-module <deployment-component>",
	Short: "Generate the full thin Terraform module for a Kubernetes-CRD-projection kind",
	Long: `The "generate-module" command emits the complete iac/tf Terraform module
(variables.tf, locals.tf, main.tf, provider.tf, outputs.tf) for a deployment
component whose spec is a direct projection of a single Kubernetes custom
resource -- i.e. a kind annotated with kubernetes_manifest_projection in
CloudResourceKindMeta (Istio, Gateway API, etc.).

The module is a thin kubernetes_manifest passthrough: variable "spec" is typed
'any' and handed to the CR verbatim, because the proto->tfvars converter already
emits the manifest-shaped (camelCase, null-pruned) spec. This removes the
hand-written snake->camel / null-prune / oneOf locals.tf that CRD modules
previously carried. Kinds without the projection annotation are rejected (use
the standard 'generate-variables' + provider-resource module pattern instead).`,
	Example: `
  # Write the module into the component's iac/tf directory
  openmcf tofu generate-module KubernetesDestinationRule \
    --output-dir apis/org/openmcf/provider/kubernetes/kubernetesdestinationrule/v1/iac/tf
`,
	Args: cobra.ExactArgs(1),
	Run:  generateModuleHandler,
}

func init() {
	GenerateModule.Flags().String(string(flag.OutputDir), "", "output directory (the component's iac/tf dir); required")
}

func generateModuleHandler(cmd *cobra.Command, args []string) {
	kindName := args[0]

	outputDir, err := cmd.Flags().GetString(string(flag.OutputDir))
	flag.HandleFlagErrAndValue(err, flag.OutputDir, outputDir)

	cloudResourceKind := crkreflect.KindFromString(kindName)
	manifestObject := crkreflect.ToMessageMap[cloudResourceKind]
	if manifestObject == nil {
		log.Fatalf("proto message not found for %s cloudResourceKind", cloudResourceKind.String())
	}

	files, err := generators.GenerateManifestModule(cloudResourceKind, manifestObject)
	if err != nil {
		log.Fatalf("failed to generate terraform module for %s: %v", kindName, err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("failed to create output directory %s: %v", outputDir, err)
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
			log.Fatalf("failed to write %s: %v", path, err)
		}
		log.Infof("wrote %s", path)
	}
}
