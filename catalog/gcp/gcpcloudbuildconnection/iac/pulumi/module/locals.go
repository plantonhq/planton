package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudbuildconnectionv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudbuildconnection/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals carries the IaC input. A connection has annotations but no
// labels, so there is no attribution label set to compute.
type Locals struct {
	GcpProviderConfig       *gcpprovider.GcpProviderConfig
	GcpCloudBuildConnection *gcpcloudbuildconnectionv1alpha1.GcpCloudBuildConnection
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcloudbuildconnectionv1alpha1.GcpCloudBuildConnectionIacInput) *Locals {
	return &Locals{
		GcpProviderConfig:       iacInput.ProviderConfig,
		GcpCloudBuildConnection: iacInput.Target,
	}
}
