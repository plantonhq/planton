package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcolabruntimev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcolabruntime/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpColabRuntime   *gcpcolabruntimev1alpha1.GcpColabRuntime
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpcolabruntimev1alpha1.GcpColabRuntimeIacInput) *Locals {
	locals := &Locals{}
	locals.GcpColabRuntime = iacInput.Target

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
