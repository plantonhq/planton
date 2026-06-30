package iacflags

import (
	"github.com/plantonhq/planton/apis/dev/planton/shared/iac/terraform"
	"github.com/plantonhq/planton/internal/cli/flag"
	"github.com/spf13/cobra"
)

// AddTofuApplyFlags adds Tofu/Terraform flags for apply and destroy commands.
func AddTofuApplyFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(string(flag.AutoApprove), false,
		"Skip interactive approval of plan before applying (Tofu/Terraform)")
}

// AddTofuPlanFlags adds Tofu/Terraform flags for the plan command.
func AddTofuPlanFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(string(flag.Destroy), false,
		"Create a destroy plan instead of apply plan (Tofu/Terraform)")
}

// AddTofuInitFlags adds Tofu/Terraform flags specific to the init command.
// These flags configure state backend settings during initialization.
func AddTofuInitFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().String(string(flag.BackendType),
		terraform.TerraformBackendType_local.String(),
		"Specifies the backend type (Tofu/Terraform) - 'local', 's3', 'gcs', 'azurerm', etc.")

	cmd.PersistentFlags().String(string(flag.BackendBucket), "",
		"State bucket name (S3/GCS) or container name (Azure)")

	cmd.PersistentFlags().String(string(flag.BackendKey), "",
		"State file path within bucket (e.g., 'env/prod/terraform.tfstate')")

	cmd.PersistentFlags().String(string(flag.BackendRegion), "",
		"Region for S3 backend (use 'auto' for S3-compatible backends like R2)")

	cmd.PersistentFlags().String(string(flag.BackendEndpoint), "",
		"Custom S3-compatible endpoint URL (required for R2, MinIO, etc.)")

	cmd.PersistentFlags().StringArray(string(flag.BackendConfig), []string{},
		"Additional backend configuration key=value pairs (Tofu/Terraform)")

	cmd.PersistentFlags().Bool(string(flag.Reconfigure), false,
		"Reconfigure backend, ignoring any saved configuration (Tofu/Terraform)")
}
