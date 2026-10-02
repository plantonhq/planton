package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpsccmuteconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpsccmuteconfig/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpSccMuteConfig  *gcpsccmuteconfigv1alpha1.GcpSccMuteConfig
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpsccmuteconfigv1alpha1.GcpSccMuteConfigStackInput) *Locals {
	locals := &Locals{}
	locals.GcpSccMuteConfig = stackInput.Target

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
