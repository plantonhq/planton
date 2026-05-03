package runner

import "github.com/pkg/errors"

// TerraformResult captures the outcome of a Terraform CLI invocation.
// This is a placeholder for T01 -- Terraform E2E support will be implemented in T02.
type TerraformResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// TerraformDeploy runs `tofu apply` for the given module directory.
// Not implemented in T01 -- returns an explicit error.
func TerraformDeploy(moduleDir, stackInputFilePath string) (*TerraformResult, error) {
	return nil, errors.New("terraform E2E execution not yet implemented (planned for T02)")
}

// TerraformDestroy runs `tofu destroy` for the given module directory.
// Not implemented in T01 -- returns an explicit error.
func TerraformDestroy(moduleDir string) (*TerraformResult, error) {
	return nil, errors.New("terraform E2E execution not yet implemented (planned for T02)")
}
