package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpcloudbuildworkerpoolv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpcloudbuildworkerpool/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals carries the stack input. A worker pool has annotations but no
// labels, so there is no attribution label set to compute.
type Locals struct {
	GcpProviderConfig       *gcpprovider.GcpProviderConfig
	GcpCloudBuildWorkerPool *gcpcloudbuildworkerpoolv1alpha1.GcpCloudBuildWorkerPool
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpcloudbuildworkerpoolv1alpha1.GcpCloudBuildWorkerPoolStackInput) *Locals {
	return &Locals{
		GcpProviderConfig:       stackInput.ProviderConfig,
		GcpCloudBuildWorkerPool: stackInput.Target,
	}
}
