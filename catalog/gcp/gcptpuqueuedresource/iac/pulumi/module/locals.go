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

func initializeLocals(_ *pulumi.Context, iacInput *gcptpuqueuedresourcev1alpha1.GcpTpuQueuedResourceIacInput) *Locals {
	locals := &Locals{}
	locals.GcpTpuQueuedResource = iacInput.Target

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
