package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpgkefleetv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpgkefleet/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the IaC input. The fleet carries no labels: the pinned
// SDK has no fleet labels argument, and both engines send the same
// arguments.
type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpGkeFleet       *gcpgkefleetv1alpha1.GcpGkeFleet
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpgkefleetv1alpha1.GcpGkeFleetIacInput) *Locals {
	locals := &Locals{}
	locals.GcpGkeFleet = iacInput.Target

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
