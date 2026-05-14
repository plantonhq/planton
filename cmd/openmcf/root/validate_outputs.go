//go:build !codegen
// +build !codegen

package root

import (
	"encoding/json"
	"os"

	"github.com/plantonhq/openmcf/internal/cli/ui/validateoutputs"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
	"github.com/plantonhq/openmcf/pkg/outputs"
	"github.com/spf13/cobra"
)

var ValidateOutputs = &cobra.Command{
	Use:   "validate-outputs",
	Short: "Validate output transformation overrides for a custom IaC module",
	Long: `Checks that a module's output_transform.yaml or transform-outputs executable
is correct and, optionally, runs a dry-run transformation against sample outputs.

Schema validation (always runs):
  - Discovers which override mechanism exists in --module-dir
  - For YAML mappings: validates structure, version, and proto field targets
  - For executables: checks that the file has execute permission

Dry-run validation (when --sample-outputs is provided):
  - Runs the full Flatten → override → Transform pipeline
  - Reports which proto fields were populated and which outputs were unmatched`,
	Example: `  # Schema-only check
  openmcf validate-outputs --kind AwsVpc --module-dir ./my-custom-module

  # Full dry-run with sample outputs
  openmcf validate-outputs --kind AwsVpc --module-dir ./my-custom-module \
    --sample-outputs ./test-outputs.json`,
	Run: validateOutputsHandler,
}

func init() {
	ValidateOutputs.Flags().String("kind", "", "CloudResourceKind name (e.g., AwsVpc, Auth0ResourceServer)")
	ValidateOutputs.Flags().String("module-dir", "", "Path to the IaC module directory containing overrides")
	ValidateOutputs.Flags().String("sample-outputs", "", "Path to a JSON file with sample raw outputs for dry-run")

	_ = ValidateOutputs.MarkFlagRequired("kind")
	_ = ValidateOutputs.MarkFlagRequired("module-dir")
}

func validateOutputsHandler(cmd *cobra.Command, args []string) {
	kindName, _ := cmd.Flags().GetString("kind")
	moduleDir, _ := cmd.Flags().GetString("module-dir")
	samplePath, _ := cmd.Flags().GetString("sample-outputs")

	kind := crkreflect.KindFromString(kindName)
	if kind == 0 {
		validateoutputs.RenderUnknownKind(kindName)
		os.Exit(1)
	}

	if info, err := os.Stat(moduleDir); err != nil || !info.IsDir() {
		validateoutputs.RenderModuleDirNotFound(moduleDir)
		os.Exit(1)
	}

	var sampleOutputs map[string]interface{}
	if samplePath != "" {
		data, err := os.ReadFile(samplePath)
		if err != nil {
			validateoutputs.RenderSampleFileError(samplePath, err)
			os.Exit(1)
		}
		if err := json.Unmarshal(data, &sampleOutputs); err != nil {
			validateoutputs.RenderSampleParseError(samplePath, err)
			os.Exit(1)
		}
	}

	result, err := outputs.ValidateOverride(kind, moduleDir, sampleOutputs)
	if err != nil {
		validateoutputs.RenderValidationInternalError(err)
		os.Exit(1)
	}

	hasSchemaErrors := len(result.SchemaErrors) > 0
	hasDryRunErrors := result.DryRun != nil && len(result.DryRun.Errors) > 0

	if hasSchemaErrors || hasDryRunErrors {
		validateoutputs.RenderValidationFailure(kindName, moduleDir, result)
		os.Exit(1)
	}

	validateoutputs.RenderValidationSuccess(kindName, moduleDir, result)
}
