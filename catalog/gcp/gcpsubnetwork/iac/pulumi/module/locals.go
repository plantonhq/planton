package module

import (
	gcpsubnetworkv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsubnetwork/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention: the resolved
// resource plus any derived values the module needs. Subnetworks accept no
// labels in GCP, so none are derived here.
type Locals struct {
	GcpSubnetwork *gcpsubnetworkv1alpha1.GcpSubnetwork
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpsubnetworkv1alpha1.GcpSubnetworkIacInput) *Locals {
	return &Locals{
		GcpSubnetwork: iacInput.Target,
	}
}
