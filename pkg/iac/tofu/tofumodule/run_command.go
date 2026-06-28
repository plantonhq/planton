package tofumodule

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/iac/terraform"
	"github.com/plantonhq/openmcf/internal/cli/workspace"
	"github.com/plantonhq/openmcf/internal/manifest"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
	"github.com/plantonhq/openmcf/pkg/iac/stackinput"
	"github.com/plantonhq/openmcf/pkg/iac/stackinput/stackinputproviderconfig"
	"github.com/plantonhq/openmcf/pkg/iac/tofu/backendconfig"
	"github.com/plantonhq/openmcf/pkg/iac/tofu/tfbackend"
	log "github.com/sirupsen/logrus"
)

// RunCommand executes an HCL-based IaC operation (init + operation) using the specified binary.
// The binaryName parameter specifies which CLI binary to use ("tofu" or "terraform").
// The backendConfig parameter is optional - if provided, it will be used directly instead of
// extracting from manifest labels. Pass nil to fall back to manifest label extraction.
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
	providerConfig *stackinputproviderconfig.ProviderConfig,
	backendConfig *backendconfig.TofuBackendConfig,
) error {
	manifestObject, err := manifest.LoadWithOverrides(targetManifestPath, valueOverrides)
	if err != nil {
		return errors.Wrapf(err, "failed to override values in target manifest file")
	}

	// Determine backend configuration:
	// 1. If backendConfig is provided (from CLI flags), use it directly
	// 2. Otherwise, extract from manifest labels (legacy path)
	var backendType terraform.TerraformBackendType = terraform.TerraformBackendType_local
	var backendConfigArgs []string

	if backendConfig != nil {
		// Use the provided backend config (from CLI flags merged with manifest labels)
		if backendConfig.BackendType != "" {
			backendType = tfbackend.BackendTypeFromString(backendConfig.BackendType)
			if backendType == terraform.TerraformBackendType_terraform_backend_type_unspecified {
				return errors.Errorf("unsupported backend type: %s", backendConfig.BackendType)
			}
			backendConfigArgs = buildBackendConfigArgs(backendConfig)
		}
		// If BackendType is empty but config is provided, use local backend (the default)
	} else {
		// Fall back to extracting from manifest labels (legacy path for direct command usage)
		tofuBackendConfig, err := backendconfig.ExtractFromManifest(manifestObject, binaryName)
		if err != nil {
			// Log but don't fail - backend config is optional
			log.Debugf("Could not extract %s backend config from manifest labels: %v", binaryName, err)
		}

		if tofuBackendConfig != nil {
			// Convert backend type string to enum
			backendType = tfbackend.BackendTypeFromString(tofuBackendConfig.BackendType)
			if backendType == terraform.TerraformBackendType_terraform_backend_type_unspecified {
				return errors.Errorf("unsupported backend type from manifest labels: %s", tofuBackendConfig.BackendType)
			}

			// Build backend config arguments based on backend type
			backendConfigArgs = buildBackendConfigArgs(tofuBackendConfig)
		} else {
			log.Debugf("No %s backend config in manifest labels, using default local backend", binaryName)
		}
	}

	kindName, err := crkreflect.ExtractKindFromProto(manifestObject)
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

	stackInputYaml, err := stackinput.BuildStackInputYaml(manifestObject, providerConfig)
	if err != nil {
		return errors.Wrap(err, "failed to build stack input yaml")
	}

	workspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		return errors.Wrap(err, "failed to get workspace directory")
	}

	providerConfigEnvVars, err := GetProviderConfigEnvVars(stackInputYaml, workspaceDir, kubeContext)
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
