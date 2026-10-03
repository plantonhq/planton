package module

import (
	gcpworkloadidentitypoolproviderv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpworkloadidentitypoolprovider/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs.
type Locals struct {
	GcpWorkloadIdentityPoolProvider *gcpworkloadidentitypoolproviderv1alpha1.GcpWorkloadIdentityPoolProvider
}

func initializeLocals(ctx *pulumi.Context, iacInput *gcpworkloadidentitypoolproviderv1alpha1.GcpWorkloadIdentityPoolProviderIacInput) *Locals {
	return &Locals{
		GcpWorkloadIdentityPoolProvider: iacInput.Target,
	}
}
