package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcptpuqueuedresourcev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcptpuqueuedresource/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig    *gcpprovider.GcpProviderConfig
	GcpTpuQueuedResource *gcptpuqueuedresourcev1alpha1.GcpTpuQueuedResource
}

func initializeLocals(_ *pulumi.Context, stackInput *gcptpuqueuedresourcev1alpha1.GcpTpuQueuedResourceStackInput) *Locals {
	locals := &Locals{}
	locals.GcpTpuQueuedResource = stackInput.Target

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
