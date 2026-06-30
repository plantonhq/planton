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

// TerraformOutputs retrieves all stack outputs as a map via Terratest.
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

// BuildTerratestOptions constructs Terratest Options from the prepared working
// directory, tfvars path, and provider environment variables.
//
// The binary defaults to "tofu" (matching Planton's CLI preference for OpenTofu).
// Set PLANTON_E2E_TF_BINARY="terraform" to use HashiCorp Terraform instead.
func BuildTerratestOptions(t testing.TB, workDir, tfvarsPath string, envVars map[string]string) *terraform.Options {
	binary := "tofu"
	if override := os.Getenv("PLANTON_E2E_TF_BINARY"); override != "" {
		binary = override
	}

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
