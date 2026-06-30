package iacrunner

import (
	"github.com/plantonhq/planton/apis/dev/planton/shared/iac/terraform"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/spf13/cobra"
)

// RunTerraform executes a Terraform operation using the resolved context.
func RunTerraform(ctx *Context, cmd *cobra.Command, operation terraform.TerraformOperationType) error {
	return runHcl(ctx, cmd, operation, provisioner.HclBinaryTerraform)
}
