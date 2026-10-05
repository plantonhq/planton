package module

import (
	gcpprojectiammemberv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpprojectiammember/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpProjectIamMember *gcpprojectiammemberv1alpha1.GcpProjectIamMember
}

func initializeLocals(ctx *pulumi.Context, iacInput *gcpprojectiammemberv1alpha1.GcpProjectIamMemberIacInput) *Locals {
	return &Locals{
		GcpProjectIamMember: iacInput.Target,
	}
}
