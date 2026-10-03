package runner

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/pkg/errors"
)

// TerraformResult captures the outcome of a Terraform CLI invocation.
type TerraformResult struct {
	Stdout   string
	Stderr   string
	Duration time.Duration
	ExitCode int
}

// TerraformDeploy runs tofu/terraform init + apply via Terratest.
// Uses the E variant to return errors instead of calling t.Fatal().
func TerraformDeploy(t testing.TB, opts *terraform.Options) (*TerraformResult, error) {
	start := time.Now()
	stdout, err := terraform.InitAndApplyE(t, opts)
	result := &TerraformResult{
		Stdout:   stdout,
		Duration: time.Since(start),
	}
	if err != nil {
		return result, errors.Wrap(err, "terraform init+apply failed")
	}
	return result, nil
}

// TerraformPlanNoChanges re-plans the just-applied configuration with
// -detailed-exitcode and fails unless the plan is empty — the Terraform arm
// of the IDEMPOTENCY phase. Exit code 0 means no changes; 2 means the module
// and the provider disagree about the applied state; anything else is a plan
// error in its own right.
func TerraformPlanNoChanges(t testing.TB, opts *terraform.Options) (*TerraformResult, error) {
	start := time.Now()
	exitCode, err := terraform.PlanExitCodeE(t, opts)
	result := &TerraformResult{
		Duration: time.Since(start),
		ExitCode: exitCode,
	}
	if err != nil {
		return result, errors.Wrap(err, "terraform re-plan failed")
	}
	if exitCode != 0 {
		return result, errors.Errorf("terraform re-plan reported pending changes after apply (idempotency violation, exit code %d)", exitCode)
	}
	return result, nil
}

// TerraformDestroy runs tofu/terraform destroy via Terratest.
func TerraformDestroy(t testing.TB, opts *terraform.Options) (*TerraformResult, error) {
	start := time.Now()
	stdout, err := terraform.DestroyE(t, opts)
	result := &TerraformResult{
		Stdout:   stdout,
		Duration: time.Since(start),
	}
	if err != nil {
		return result, errors.Wrap(err, "terraform destroy failed")
	}
	return result, nil
}

// TerraformOutputs retrieves all outputs as a map via Terratest.
func TerraformOutputs(t testing.TB, opts *terraform.Options) (map[string]interface{}, error) {
	outputs, err := terraform.OutputAllE(t, opts)
	if err != nil {
		return nil, errors.Wrap(err, "terraform output failed")
	}

	result := make(map[string]interface{}, len(outputs))
	for k, v := range outputs {
		result[k] = v
	}
	return result, nil
}

// TerraformBinary is the HCL engine binary every terraform-engine lane runs: "tofu" (matching
// Planton's CLI preference for OpenTofu), or PLANTON_E2E_TF_BINARY when set ("terraform" for
// HashiCorp Terraform). RunKindTest refuses the binary for a kind that does not declare it.
func TerraformBinary() string {
	if override := os.Getenv("PLANTON_E2E_TF_BINARY"); override != "" {
		return override
	}
	return "tofu"
}

// BuildTerratestOptions constructs Terratest Options from the prepared working
// directory, tfvars path, and provider environment variables, running TerraformBinary.
func BuildTerratestOptions(t testing.TB, workDir, tfvarsPath string, envVars map[string]string) *terraform.Options {
	binary := TerraformBinary()

	fmt.Printf("  [terraform] binary=%s workDir=%s\n", binary, workDir)

	opts := &terraform.Options{
		TerraformDir:    workDir,
		TerraformBinary: binary,
		VarFiles:        []string{tfvarsPath},
		EnvVars:         envVars,
		NoColor:         true,
	}

	return terraform.WithDefaultRetryableErrors(t, opts)
}
