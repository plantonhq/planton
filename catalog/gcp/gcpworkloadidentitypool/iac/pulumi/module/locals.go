package module

import (
	gcpworkloadidentitypoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpworkloadidentitypool/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpWorkloadIdentityPool *gcpworkloadidentitypoolv1alpha1.GcpWorkloadIdentityPool
}

func initializeLocals(ctx *pulumi.Context, iacInput *gcpworkloadidentitypoolv1alpha1.GcpWorkloadIdentityPoolIacInput) *Locals {
	return &Locals{
		GcpWorkloadIdentityPool: iacInput.Target,
	}
}
