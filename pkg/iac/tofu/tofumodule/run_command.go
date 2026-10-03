package tofumodule

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/internal/manifest"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/iacinput"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/tofu/backendconfig"
	"github.com/plantonhq/planton/pkg/iac/tofu/tfbackend"
	"github.com/plantonhq/planton/pkg/iac/tofu/tfoverride"
	"github.com/plantonhq/planton/shared/iac/terraform"
	log "github.com/sirupsen/logrus"
)

// RunCommand executes an HCL-based IaC operation (init + operation) using the specified binary.
// The binaryName parameter specifies which CLI binary to use ("tofu" or "terraform").
// The backendConfig parameter is optional - if provided, it will be used directly instead of
// extracting from manifest annotations. Pass nil to fall back to manifest annotation extraction.
func RunCommand(
	binaryName string,
	inputModuleDir string,
	targetManifestPath string,
	terraformOperation terraform.TerraformOperationType,
	valueOverrides map[string]string,
	isAutoApprove bool,
	isDestroyPlan bool,
	isReconfigure bool,
	moduleVersion string,
	noCleanup bool,
	kubeContext string,
	providerConfig *iacinputproviderconfig.ProviderConfig,
	backendConfig *backendconfig.TofuBackendConfig,
	opts ...RunOption,
) error {
	var cfg runConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	manifestObject, err := manifest.LoadWithOverrides(targetManifestPath, valueOverrides)
	if err != nil {
		return errors.Wrapf(err, "failed to override values in target manifest file")
	}

	// Determine backend configuration:
	// 1. If backendConfig is provided (from CLI flags), use it directly
	// 2. Otherwise, extract from manifest annotations (legacy path)
	var backendType terraform.TerraformBackendType = terraform.TerraformBackendType_local
	var backendConfigArgs []string

	if backendConfig != nil {
		// Use the provided backend config (from CLI flags merged with manifest annotations)
		if backendConfig.BackendType != "" {
			backendType = tfbackend.BackendTypeFromString(backendConfig.BackendType)
			if backendType == terraform.TerraformBackendType_terraform_backend_type_unspecified {
				return errors.Errorf("unsupported backend type: %s", backendConfig.BackendType)
			}
			backendConfigArgs = buildBackendConfigArgs(backendConfig)
		}
		// If BackendType is empty but config is provided, use local backend (the default)
	} else {
		// Fall back to extracting from manifest annotations (legacy path for direct command usage)
		tofuBackendConfig, err := backendconfig.ExtractFromManifest(manifestObject, binaryName)
		if err != nil {
			// Log but don't fail - backend config is optional
			log.Debugf("Could not extract %s backend config from manifest annotations: %v", binaryName, err)
		}

		if tofuBackendConfig != nil {
			// Convert backend type string to enum
			backendType = tfbackend.BackendTypeFromString(tofuBackendConfig.BackendType)
			if backendType == terraform.TerraformBackendType_terraform_backend_type_unspecified {
				return errors.Errorf("unsupported backend type from manifest annotations: %s", tofuBackendConfig.BackendType)
			}

			// Build backend config arguments based on backend type
			backendConfigArgs = buildBackendConfigArgs(tofuBackendConfig)
		} else {
			log.Debugf("No %s backend config in manifest annotations, using default local backend", binaryName)
		}
	}

	kindName, err := catalogkindreflect.ExtractKindFromProto(manifestObject)
	if err != nil {
		return errors.Wrapf(err, "failed to extract kind name from manifest proto")
	}

	// Get module path using staging-based approach
	pathResult, err := GetModulePath(inputModuleDir, kindName, moduleVersion, noCleanup)
	if err != nil {
		return errors.Wrapf(err, "failed to get %s module directory", binaryName)
	}

	// Setup cleanup to run after execution
	if pathResult.ShouldCleanup {
		defer func() {
			if cleanupErr := pathResult.CleanupFunc(); cleanupErr != nil {
				fmt.Printf("Warning: failed to cleanup workspace copy: %v\n", cleanupErr)
			}
		}()
	}

	modulePath := pathResult.ModulePath

	iacInputYaml, err := iacinput.BuildIacInputYaml(manifestObject, providerConfig)
	if err != nil {
		return errors.Wrap(err, "failed to build IaC input yaml")
	}

	// Write (or remove) the provider-override file carrying the provider-block
	// arguments that cannot ride env vars (assume-role chain, endpoints, ...).
	// Deferred removal keeps module directories that outlive this run clean --
	// two of the three GetModulePath modes return a non-disposable directory
	// (the user's own checkout, the zip cache), where a leftover override would
	// silently apply this run's provider settings to a later run.
	wroteOverride, err := tfoverride.WriteProviderOverrideFile(modulePath, iacInputYaml)
	if err != nil {
		return errors.Wrap(err, "failed to write provider override file")
	}
	if wroteOverride {
		defer func() {
			if removeErr := os.Remove(filepath.Join(modulePath, tfoverride.OverrideFileName)); removeErr != nil && !os.IsNotExist(removeErr) {
				fmt.Printf("Warning: failed to remove %s: %v\n", tfoverride.OverrideFileName, removeErr)
			}
		}()
	}

	workspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		return errors.Wrap(err, "failed to get workspace directory")
	}

	providerConfigEnvVars, err := GetProviderConfigEnvVars(iacInputYaml, workspaceDir, kubeContext)
	if err != nil {
		return errors.Wrap(err, "failed to get provider config env vars")
	}

	// Initialize with backend configuration before any operation.
	// CLI usage has no cancellation context to thread, so use context.Background();
	// the runner passes the real activity ctx through its own RunOperation/Init calls.
	err = Init(context.Background(), binaryName, modulePath, manifestObject, backendType, backendConfigArgs,
		providerConfigEnvVars, isReconfigure, false, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to initialize %s module", binaryName)
	}

	err = RunOperation(context.Background(), binaryName, modulePath, terraformOperation,
		isAutoApprove, isDestroyPlan, manifestObject,
		providerConfigEnvVars, false, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to run %s operation", binaryName)
	}

	// Capture must run here, before the deferred workspace cleanup and
	// provider-override removal fire: `output -json` re-initializes the
	// backend and providers, so it needs the workspace exactly as the apply
	// left it. Only apply captures — plan writes no state, and destroy and
	// refresh have no fresh outputs to read.
	if cfg.captureSink != nil && terraformOperation == terraform.TerraformOperationType_apply {
		if captureErr := captureOutputs(context.Background(), binaryName, modulePath, kindName,
			providerConfigEnvVars, cfg.captureSink); captureErr != nil {
			// The apply already succeeded; a capture failure must not turn a
			// deployed stack into a failed command. Report and move on.
			log.Warnf("outputs could not be captured after apply: %v", captureErr)
		}
	}

	return nil
}

// buildBackendConfigArgs builds backend configuration arguments based on backend type.
// For S3-compatible backends (R2, MinIO, etc.), it adds the endpoint and skip flags.
// Both Terraform and OpenTofu support the same S3 backend configuration format.
func buildBackendConfigArgs(config *backendconfig.TofuBackendConfig) []string {
	var args []string

	switch config.BackendType {
	case "s3":
		// S3 backend: bucket, key, and region
		if config.BackendBucket != "" {
			args = append(args, fmt.Sprintf("bucket=%s", config.BackendBucket))
		}
		if config.BackendKey != "" {
			args = append(args, fmt.Sprintf("key=%s", config.BackendKey))
		}
		if config.BackendRegion != "" {
			args = append(args, fmt.Sprintf("region=%s", config.BackendRegion))
		}

		// S3-compatible endpoint (R2, MinIO, etc.)
		// Both Terraform and OpenTofu use the endpoints={s3="..."} format
		if config.BackendEndpoint != "" {
			args = append(args, fmt.Sprintf("endpoints={s3=\"%s\"}", config.BackendEndpoint))
		}

		// S3-compatible skip flags (auto-enabled when S3Compatible is true)
		// These are required for non-AWS S3 implementations like Cloudflare R2 or MinIO
		// All flags are supported by both Terraform and OpenTofu
		if config.S3Compatible {
			args = append(args, "skip_credentials_validation=true")
			args = append(args, "skip_region_validation=true")
			args = append(args, "skip_metadata_api_check=true")
			args = append(args, "skip_requesting_account_id=true")
			args = append(args, "skip_s3_checksum=true")
			args = append(args, "use_path_style=true")
		}

	case "gcs":
		// GCS backend: bucket and prefix (key is called prefix in GCS)
		if config.BackendBucket != "" {
			args = append(args, fmt.Sprintf("bucket=%s", config.BackendBucket))
		}
		if config.BackendKey != "" {
			args = append(args, fmt.Sprintf("prefix=%s", config.BackendKey))
		}

	case "azurerm":
		// Azure backend: container_name and key
		if config.BackendBucket != "" {
			args = append(args, fmt.Sprintf("container_name=%s", config.BackendBucket))
		}
		if config.BackendKey != "" {
			args = append(args, fmt.Sprintf("key=%s", config.BackendKey))
		}

	case "local":
		// Local backend doesn't need config args
		// The path is handled by terraform itself
	}

	return args
}
