package iacrunner

import (
	"fmt"
	"os"

	"github.com/plantonhq/planton/apis/dev/planton/shared/iac/terraform"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/plantonhq/planton/internal/cli/prompt"
	"github.com/plantonhq/planton/internal/cli/ui"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/pkg/iac/tofu/backendconfig"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofumodule"
	"github.com/spf13/cobra"
)

// RunTofu executes an OpenTofu operation using the resolved context.
func RunTofu(ctx *Context, cmd *cobra.Command, operation terraform.TerraformOperationType) error {
	return runHcl(ctx, cmd, operation, provisioner.HclBinaryTofu)
}

// runHcl executes an HCL-based IaC operation (tofu or terraform) using the resolved context.
func runHcl(ctx *Context, cmd *cobra.Command, operation terraform.TerraformOperationType, binary provisioner.HclBinary) error {
	// Get auto-approve flag if defined (ignore error for commands that don't register it)
	isAutoApprove, _ := cmd.Flags().GetBool(string(flag.AutoApprove))

	// Get reconfigure flag if defined (ignore error for commands that don't register it)
	isReconfigure, _ := cmd.Flags().GetBool(string(flag.Reconfigure))

	// For plan operation, check if it's a destroy plan
	isDestroyPlan := false
	if operation == terraform.TerraformOperationType_plan {
		isDestroyPlan, _ = cmd.Flags().GetBool(string(flag.Destroy))
		// Plan is always auto-approve (non-interactive)
		isAutoApprove = true
	}

	// For refresh operation, no approval needed (read-only state sync)
	if operation == terraform.TerraformOperationType_refresh {
		isAutoApprove = true
	}

	// Check if binary is available before proceeding
	if err := binary.CheckAvailable(); err != nil {
		cliprint.PrintError(err.Error())
		os.Exit(1)
	}

	// Build and validate backend configuration before handoff
	backendCfg, err := buildAndValidateBackendConfig(ctx, cmd, string(binary))
	if err != nil {
		cliprint.PrintError(err.Error())
		os.Exit(1)
	}

	// ALWAYS display backend configuration before handoff
	if backendCfg != nil {
		ui.BackendConfigSummary(backendCfg)
	}

	// Display module path
	cliprint.PrintModulePath(ctx.ModuleDir)

	cliprint.PrintHandoff(binary.DisplayName())

	err = tofumodule.RunCommand(
		string(binary),
		ctx.ModuleDir,
		ctx.ManifestPath,
		operation,
		ctx.ValueOverrides,
		isAutoApprove,
		isDestroyPlan,
		isReconfigure,
		ctx.ModuleVersion,
		ctx.NoCleanup,
		ctx.KubeContext,
		ctx.ProviderConfig,
		backendCfg,
	)
	if err != nil {
		printHclExecutionError(binary, err)
		os.Exit(1)
	}

	printHclSuccess(binary)
	return nil
}

// buildAndValidateBackendConfig builds backend config from CLI flags and manifest labels,
// validates it, and prompts for missing values if in interactive mode.
func buildAndValidateBackendConfig(ctx *Context, cmd *cobra.Command, provisionerType string) (*backendconfig.TofuBackendConfig, error) {
	// Extract CLI flags for backend configuration
	cliFlags := extractCLIBackendFlags(cmd)

	// Build merged configuration (CLI flags override manifest labels)
	config, err := backendconfig.BuildBackendConfig(ctx.ManifestObject, provisionerType, cliFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to build backend configuration: %w", err)
	}

	// Check for incomplete configuration (has fields but no type)
	if config.BackendType == "" && hasAnyBackendFields(config) {
		ui.IncompleteBackendConfigWarning()
	}

	// Skip validation for local backend or when no backend is configured
	if config.BackendType == "" || config.BackendType == "local" {
		return config, nil
	}

	// Detect and announce S3-compatible backend
	if config.S3Compatible {
		if config.BackendRegion == "auto" {
			ui.S3CompatibleDetected("Region is set to 'auto', indicating an S3-compatible backend")
		} else if config.BackendEndpoint != "" {
			ui.S3CompatibleDetected("Custom endpoint detected")
		}
	}

	// Validate configuration completeness
	validation := backendconfig.Validate(config)

	// If configuration is incomplete, handle based on interactivity
	if !validation.Valid {
		if !prompt.IsInteractive() {
			// Non-interactive mode: show error and fail
			ui.MissingBackendConfigError(validation.MissingFields, config.BackendType)
			return nil, fmt.Errorf("incomplete backend configuration - provide missing values via CLI flags or manifest labels")
		}

		// Interactive mode: prompt for missing values
		ui.MissingBackendConfigError(validation.MissingFields, config.BackendType)
		config, err = prompt.PromptForMissingBackendConfig(config, validation.MissingFields)
		if err != nil {
			return nil, fmt.Errorf("failed to get backend configuration: %w", err)
		}
	}

	return config, nil
}

// extractCLIBackendFlags extracts backend configuration from CLI flags.
func extractCLIBackendFlags(cmd *cobra.Command) backendconfig.CLIBackendFlags {
	// Get values from flags, ignoring errors for flags that aren't registered
	backendType, _ := cmd.Flags().GetString(string(flag.BackendType))
	backendBucket, _ := cmd.Flags().GetString(string(flag.BackendBucket))
	backendKey, _ := cmd.Flags().GetString(string(flag.BackendKey))
	backendRegion, _ := cmd.Flags().GetString(string(flag.BackendRegion))
	backendEndpoint, _ := cmd.Flags().GetString(string(flag.BackendEndpoint))

	return backendconfig.CLIBackendFlags{
		BackendType:     backendType,
		BackendBucket:   backendBucket,
		BackendKey:      backendKey,
		BackendRegion:   backendRegion,
		BackendEndpoint: backendEndpoint,
	}
}

// hasAnyBackendFields returns true if any backend field is configured.
func hasAnyBackendFields(config *backendconfig.TofuBackendConfig) bool {
	return config.BackendBucket != "" ||
		config.BackendKey != "" ||
		config.BackendRegion != "" ||
		config.BackendEndpoint != ""
}

// printHclSuccess prints a success message for the appropriate binary.
func printHclSuccess(binary provisioner.HclBinary) {
	switch binary {
	case provisioner.HclBinaryTofu:
		cliprint.PrintTofuSuccess()
	case provisioner.HclBinaryTerraform:
		cliprint.PrintTerraformSuccess()
	}
}

// printHclFailure prints a failure message for the appropriate binary.
func printHclFailure(binary provisioner.HclBinary) {
	switch binary {
	case provisioner.HclBinaryTofu:
		cliprint.PrintTofuFailure()
	case provisioner.HclBinaryTerraform:
		cliprint.PrintTerraformFailure()
	}
}

// printHclExecutionError prints a beautiful error message when HCL execution fails.
// It displays the actual error message along with helpful hints for troubleshooting.
func printHclExecutionError(binary provisioner.HclBinary, err error) {
	title := fmt.Sprintf("%s Execution Failed", binary.DisplayName())
	ui.ErrorWithoutExit(title, err.Error(),
		"Check the module configuration for syntax errors",
		"Ensure all required provider credentials are configured")
}
