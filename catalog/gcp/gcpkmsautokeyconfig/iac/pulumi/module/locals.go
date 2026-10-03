package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpkmsautokeyconfigv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpkmsautokeyconfig/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig   *gcpprovider.GcpProviderConfig
	GcpKmsAutokeyConfig *gcpkmsautokeyconfigv1alpha1.GcpKmsAutokeyConfig
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpkmsautokeyconfigv1alpha1.GcpKmsAutokeyConfigIacInput) *Locals {
	locals := &Locals{}
	locals.GcpKmsAutokeyConfig = iacInput.Target

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
