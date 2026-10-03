package module

import (
	gcpserviceaccountiammemberv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpserviceaccountiammember/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpServiceAccountIamMember *gcpserviceaccountiammemberv1alpha1.GcpServiceAccountIamMember
}

func initializeLocals(ctx *pulumi.Context, iacInput *gcpserviceaccountiammemberv1alpha1.GcpServiceAccountIamMemberIacInput) *Locals {
	return &Locals{
		GcpServiceAccountIamMember: iacInput.Target,
	}
}
