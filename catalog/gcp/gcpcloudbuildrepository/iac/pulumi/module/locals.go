package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudbuildrepositoryv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudbuildrepository/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals carries the IaC input. A repository link has annotations but no
// labels, so there is no attribution label set to compute.
type Locals struct {
	GcpProviderConfig       *gcpprovider.GcpProviderConfig
	GcpCloudBuildRepository *gcpcloudbuildrepositoryv1alpha1.GcpCloudBuildRepository
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcloudbuildrepositoryv1alpha1.GcpCloudBuildRepositoryIacInput) *Locals {
	return &Locals{
		GcpProviderConfig:       iacInput.ProviderConfig,
		GcpCloudBuildRepository: iacInput.Target,
	}
}
