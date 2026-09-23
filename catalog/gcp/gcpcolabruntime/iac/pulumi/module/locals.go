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

func initializeLocals(_ *pulumi.Context, stackInput *gcpcolabruntimev1alpha1.GcpColabRuntimeStackInput) *Locals {
	locals := &Locals{}
	locals.GcpColabRuntime = stackInput.Target

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
