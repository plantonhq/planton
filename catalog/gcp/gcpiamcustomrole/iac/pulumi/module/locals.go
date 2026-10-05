package module

import (
	gcpiamcustomrolev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpiamcustomrole/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpIamCustomRole *gcpiamcustomrolev1alpha1.GcpIamCustomRole
}

func initializeLocals(ctx *pulumi.Context, iacInput *gcpiamcustomrolev1alpha1.GcpIamCustomRoleIacInput) *Locals {
	return &Locals{
		GcpIamCustomRole: iacInput.Target,
	}
}
