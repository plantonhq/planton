package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudbuildtriggerv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudbuildtrigger/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals carries the IaC input. A trigger has tags but no
// labels, so there is no attribution label set to compute.
type Locals struct {
	GcpProviderConfig    *gcpprovider.GcpProviderConfig
	GcpCloudBuildTrigger *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTrigger
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcloudbuildtriggerv1alpha1.GcpCloudBuildTriggerIacInput) *Locals {
	return &Locals{
		GcpProviderConfig:    iacInput.ProviderConfig,
		GcpCloudBuildTrigger: iacInput.Target,
	}
}
