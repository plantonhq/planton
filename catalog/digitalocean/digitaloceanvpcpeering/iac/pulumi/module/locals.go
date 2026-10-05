package module

import (
	digitaloceanvpcpeeringv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanvpcpeering/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module. The peering
// resource has no tag surface, so no Planton label set applies.
type Locals struct {
	DigitalOceanVpcPeering *digitaloceanvpcpeeringv1alpha1.DigitalOceanVpcPeering
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceanvpcpeeringv1alpha1.DigitalOceanVpcPeeringIacInput) *Locals {
	return &Locals{
		DigitalOceanVpcPeering: iacInput.Target,
	}
}
