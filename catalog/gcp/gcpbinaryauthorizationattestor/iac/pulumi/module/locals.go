package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpbinaryauthorizationattestorv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbinaryauthorizationattestor/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig              *gcpprovider.GcpProviderConfig
	GcpBinaryAuthorizationAttestor *gcpbinaryauthorizationattestorv1alpha1.GcpBinaryAuthorizationAttestor
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpbinaryauthorizationattestorv1alpha1.GcpBinaryAuthorizationAttestorIacInput) *Locals {
	locals := &Locals{}
	locals.GcpBinaryAuthorizationAttestor = iacInput.Target

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
