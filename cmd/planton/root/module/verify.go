//go:build !codegen
// +build !codegen

package module

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	moduleverifyui "github.com/plantonhq/planton/internal/cli/ui/moduleverify"
	"github.com/plantonhq/planton/internal/cli/ui/validateoutputs"
	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/iac/moduleverify"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/spf13/cobra"
)

var Verify = &cobra.Command{
	Use:   "verify",
	Short: "prove a module still conforms to its kind's contract",
	Long: `Verifies that an IaC module directory honors the contract of a cloud
resource kind — run it after every meaningful change to a customized module,
and before registering one for your organization's deployments.

What is checked:
  - OpenTofu/Terraform: the declared input surface (variables.tf) against
    the kind's schema, the outputs contract, and — when tofu or terraform is
    on PATH — the engine's own validation against a temporary copy.
  - Pulumi: the project file's go runtime, the entrypoint's typed
    stack-input contract, the Go module context, and — when go is on PATH —
    a real compile.
  - Both engines: every secret home the kind declares (the field a secret
    goes in instead of one every viewer reads, such as a Cloud Run env
    entry's secret_value) is read by the module, so a secret an author puts
    where the platform asks is never silently left out of the deployment.

Every finding is reported with its deployment impact: errors fail
deployments and fail this command; warnings are worth a look but do not.`,
	Example: `  # Verify a customized OpenTofu module
  planton module verify --kind AwsS3Bucket --module-dir ./my-s3-module

  # Verify a Pulumi module and dry-run its outputs against samples
  planton module verify --kind AwsS3Bucket --module-dir ./my-s3-module \
    --sample-outputs ./sample-outputs.json

  # Name the engine explicitly when the directory shape is ambiguous
  planton module verify --kind GcpGkeCluster --module-dir . --provisioner tofu`,
	RunE: verifyHandler,
	// The handler renders the full report itself; a returned error is a
	// one-line summary that must not be buried under usage output.
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	Verify.Flags().String("kind", "", "cloud resource kind the module serves (e.g. AwsS3Bucket)")
	Verify.Flags().String("module-dir", "", "path to the module directory to verify")
	Verify.Flags().String("provisioner", "", "the module's engine: tofu, terraform, or pulumi (default: inferred from the directory)")
	Verify.Flags().String("sample-outputs", "", "path to a JSON file with sample raw outputs for a transformation dry-run")
	Verify.Flags().Bool("skip-build-checks", false, "skip the checks that run the engine's toolchain (tofu validate / go build)")

	_ = Verify.MarkFlagRequired("kind")
	_ = Verify.MarkFlagRequired("module-dir")
}

func verifyHandler(cmd *cobra.Command, args []string) error {
	kindName, _ := cmd.Flags().GetString("kind")
	moduleDir, _ := cmd.Flags().GetString("module-dir")
	provisionerName, _ := cmd.Flags().GetString("provisioner")
	samplePath, _ := cmd.Flags().GetString("sample-outputs")
	skipBuildChecks, _ := cmd.Flags().GetBool("skip-build-checks")

	kind := crkreflect.KindFromString(kindName)
	if kind == 0 {
		validateoutputs.RenderUnknownKind(kindName)
		return errors.Errorf("unknown cloud resource kind %q", kindName)
	}

	prov := provisioner.ProvisionerTypeUnspecified
	if provisionerName != "" {
		var err error
		if prov, err = provisioner.FromString(provisionerName); err != nil {
			return err
		}
	}

	var sampleOutputs map[string]interface{}
	if samplePath != "" {
		data, err := os.ReadFile(samplePath)
		if err != nil {
			validateoutputs.RenderSampleFileError(samplePath, err)
			return errors.Errorf("cannot read sample outputs at %s", samplePath)
		}
		if err := json.Unmarshal(data, &sampleOutputs); err != nil {
			validateoutputs.RenderSampleParseError(samplePath, err)
			return errors.Errorf("cannot parse sample outputs at %s", samplePath)
		}
	}

	result, err := moduleverify.Verify(moduleverify.Input{
		KindName:            kindName,
		ModuleDir:           moduleDir,
		Provisioner:         prov,
		SampleOutputs:       sampleOutputs,
		SkipToolchainChecks: skipBuildChecks,
	})
	if err != nil {
		return err
	}

	moduleverifyui.RenderResult(result)

	if result.HasErrors() {
		return errors.Errorf("the module breaks the %s contract — the errors above fail deployments", result.KindName)
	}
	return nil
}
