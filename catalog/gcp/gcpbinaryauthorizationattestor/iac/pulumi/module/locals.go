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

func initializeLocals(_ *pulumi.Context, stackInput *gcpbinaryauthorizationattestorv1alpha1.GcpBinaryAuthorizationAttestorStackInput) *Locals {
	locals := &Locals{}
	locals.GcpBinaryAuthorizationAttestor = stackInput.Target

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
