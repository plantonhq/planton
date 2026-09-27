package runner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/iac/stackinput"
	"github.com/plantonhq/planton/pkg/iac/stackinput/providerenvvars"
	"github.com/plantonhq/planton/pkg/iac/stackinput/stackinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/tofu/generators"
	"github.com/plantonhq/planton/pkg/iac/tofu/tfbackend"
	"github.com/plantonhq/planton/pkg/iac/tofu/tfoverride"
	"github.com/plantonhq/planton/shared/iac/terraform"
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
// writes the provider-override file when the component's provider-config fixture
// carries provider-block arguments, and extracts provider environment variables.
//
// The workDir must already contain the TF module files (copied by PrepareWorkDir).
// providerConfig is nil for the harness's default posture (ambient credentials,
// empty provider block); see LoadProviderConfigFixture.
func BuildTerraformInput(manifestPath, workDir string,
	providerConfig *stackinputproviderconfig.ProviderConfig) (*TerraformInput, error) {
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
	stackInputYaml, err := stackinput.BuildStackInputYaml(manifestObject, providerConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build stack-input YAML for provider env var extraction")
	}

	// Provider-block arguments (assume-role chain, default tags, ...) cannot
	// ride env vars; they reach the module through the generated override
	// file -- the same seam the CLI and the platform runner use. A no-op when
	// the fixture carries none (or there is no fixture); workDir is a
	// disposable per-test copy, so no cleanup is needed.
	if _, err := tfoverride.WriteProviderOverrideFile(workDir, stackInputYaml); err != nil {
		return nil, errors.Wrap(err, "failed to write provider override file")
	}

	// The kind harness exports KUBECONFIG into the process and passes no
	// provider config, which is exactly the operator's local workflow; the
	// loader's ambient kubernetes branch hands the Terraform providers that
	// kubeconfig under the names they read. The harness adds nothing of its
	// own, so a lane here proves what a laptop gets.
	// A provider configuration that carries a connection (a lane identity's
	// kubeconfig) is written to a file the engine reads by path, so the file
	// lives in the lane's own working directory: absolute for the providers,
	// and gone with the directory when the lane ends.
	providerEnvVarMap, err := providerenvvars.GetEnvVarsWithOptions(stackInputYaml, providerenvvars.Options{FileCacheLoc: workDir, Engine: providerenvvars.EngineReadsEnvironment})
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract provider environment variables")
	}

	return &TerraformInput{
		TfvarsPath: tfvarsPath,
		EnvVars:    providerEnvVarMap,
	}, nil
}

// PrepareWorkDir creates an isolated temporary copy of a TF module
// directory. Terraform state files (.terraform/, terraform.tfstate) live in
// the working directory, so each test needs its own copy to avoid state
// conflicts.
//
// The copy is the module directory and NOTHING beside it -- the same file
// set release packaging zips into module.zip and the OpenTofu runtime
// extracts into its cache. A module is its directory and reads nothing
// outside it (the anatomy gate and
// hack/guards/ensure_modules_are_self_contained.sh enforce this), so a
// lane that passes here proves what a published release ships; copying
// any wider tree would let a lane pass on a layout no release has.
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

	// Never copied: engine-local state (a shared checkout may carry a
	// developer's .terraform plugin tree and state files).
	skip := map[string]bool{
		".terraform": true,
	}

	var copyTree func(src, dst string) error
	copyTree = func(src, dst string) error {
		entries, err := os.ReadDir(src)
		if err != nil {
			return errors.Wrapf(err, "failed to read directory %s", src)
		}
		if err := os.MkdirAll(dst, 0755); err != nil {
			return errors.Wrapf(err, "failed to create directory %s", dst)
		}
		for _, entry := range entries {
			name := entry.Name()
			if skip[name] || strings.HasPrefix(name, "terraform.tfstate") || name == "terraform.tfplan" {
				continue
			}
			srcPath := filepath.Join(src, name)
			dstPath := filepath.Join(dst, name)
			if entry.IsDir() {
				if err := copyTree(srcPath, dstPath); err != nil {
					return err
				}
				continue
			}
			content, err := os.ReadFile(srcPath)
			if err != nil {
				return errors.Wrapf(err, "failed to read %s", srcPath)
			}
			if err := os.WriteFile(dstPath, content, 0644); err != nil {
				return errors.Wrapf(err, "failed to write %s", dstPath)
			}
		}
		return nil
	}

	if err := copyTree(filepath.Clean(sourceModuleDir), workDir); err != nil {
		cleanup()
		return "", nil, err
	}

	return workDir, cleanup, nil
}
