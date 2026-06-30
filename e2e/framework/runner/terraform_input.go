package runner

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/apis/dev/planton/shared/iac/terraform"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/iac/stackinput"
	"github.com/plantonhq/planton/pkg/iac/stackinput/providerenvvars"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
	"github.com/plantonhq/planton/pkg/iac/tofu/tfbackend"
)

// TerraformInput holds the prepared inputs for a Terraform E2E test run.
type TerraformInput struct {
	// TfvarsPath is the absolute path to the generated terraform.tfvars file.
	TfvarsPath string

	// EnvVars holds provider-specific environment variables (e.g., KUBECONFIG).
	EnvVars map[string]string
}

// BuildTerraformInput prepares a Terraform module working directory for E2E testing.
// It loads the manifest, generates a tfvars file, writes the backend configuration,
// and extracts provider environment variables.
//
// The workDir must already contain the TF module files (copied by PrepareWorkDir).
func BuildTerraformInput(manifestPath, workDir string) (*TerraformInput, error) {
	manifestObject, err := manifest.LoadManifest(manifestPath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load manifest from %s", manifestPath)
	}

	// Generate terraform.tfvars from the proto manifest.
	// The tfvars file is placed inside the working directory so tofu init
	// can find it alongside the module's .tf files.
	tfvarsPath := filepath.Join(workDir, "terraform.tfvars")
	if err := generators.WriteVarFile(manifestObject, tfvarsPath); err != nil {
		return nil, errors.Wrap(err, "failed to generate terraform.tfvars from manifest")
	}

	// Write backend.tf with local backend for ephemeral E2E state.
	if err := tfbackend.WriteBackendFile(workDir, terraform.TerraformBackendType_local); err != nil {
		return nil, errors.Wrap(err, "failed to write backend.tf")
	}

	// Build stack-input YAML to extract provider environment variables.
	// For Kubernetes on kind, this produces KUBECONFIG.
	// For cloud providers, this produces AWS_ACCESS_KEY_ID, GOOGLE_CREDENTIALS, etc.
	stackInputYaml, err := stackinput.BuildStackInputYaml(manifestObject, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build stack-input YAML for provider env var extraction")
	}

	providerEnvVarMap, err := providerenvvars.GetEnvVarsWithOptions(stackInputYaml, providerenvvars.Options{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract provider environment variables")
	}

	// The Terraform Kubernetes provider uses KUBE_CONFIG_PATH (not KUBECONFIG).
	// This bridges a DIFFERENT case than providerenvvars.loadKubernetesEnvVars (which now
	// sets both names for connection-derived kubeconfigs): here the kind harness exports
	// KUBECONFIG into the process for an in-cluster test kubeconfig, so we forward it to
	// KUBE_CONFIG_PATH for the TF provider. Kept distinct on purpose.
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		if _, exists := providerEnvVarMap["KUBE_CONFIG_PATH"]; !exists {
			providerEnvVarMap["KUBE_CONFIG_PATH"] = kubeconfig
		}
	}

	return &TerraformInput{
		TfvarsPath: tfvarsPath,
		EnvVars:    providerEnvVarMap,
	}, nil
}

// PrepareWorkDir creates an isolated temporary copy of a TF module directory.
// Terraform state files (.terraform/, terraform.tfstate) live in the working
// directory, so each test needs its own copy to avoid state conflicts.
//
// Returns the working directory path and a cleanup function.
func PrepareWorkDir(sourceModuleDir string) (string, func(), error) {
	workDir, err := os.MkdirTemp("", "planton-e2e-tf-*")
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to create temp directory for TF module")
	}

	cleanup := func() {
		os.RemoveAll(workDir)
	}

	entries, err := os.ReadDir(sourceModuleDir)
	if err != nil {
		cleanup()
		return "", nil, errors.Wrapf(err, "failed to read TF module directory %s", sourceModuleDir)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		srcPath := filepath.Join(sourceModuleDir, entry.Name())
		dstPath := filepath.Join(workDir, entry.Name())

		content, err := os.ReadFile(srcPath)
		if err != nil {
			cleanup()
			return "", nil, errors.Wrapf(err, "failed to read %s", srcPath)
		}
		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			cleanup()
			return "", nil, errors.Wrapf(err, "failed to write %s", dstPath)
		}
	}

	return workDir, cleanup, nil
}
