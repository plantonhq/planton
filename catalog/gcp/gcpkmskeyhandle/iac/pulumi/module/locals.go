package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpkmskeyhandlev1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpkmskeyhandle/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpKmsKeyHandle   *gcpkmskeyhandlev1alpha1.GcpKmsKeyHandle
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpkmskeyhandlev1alpha1.GcpKmsKeyHandleIacInput) *Locals {
	locals := &Locals{}
	locals.GcpKmsKeyHandle = iacInput.Target

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
