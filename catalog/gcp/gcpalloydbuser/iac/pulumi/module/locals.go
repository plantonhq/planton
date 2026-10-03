package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpalloydbuserv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpalloydbuser/v1alpha1"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpAlloydbUser    *gcpalloydbuserv1alpha1.GcpAlloydbUser
}

func initializeLocals(iacInput *gcpalloydbuserv1alpha1.GcpAlloydbUserIacInput) *Locals {
	return &Locals{
		GcpAlloydbUser:    iacInput.Target,
		GcpProviderConfig: iacInput.ProviderConfig,
	}
}
