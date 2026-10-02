package module

import (
	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpbinaryauthorizationpolicyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpbinaryauthorizationpolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig            *gcpprovider.GcpProviderConfig
	GcpBinaryAuthorizationPolicy *gcpbinaryauthorizationpolicyv1alpha1.GcpBinaryAuthorizationPolicy
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpbinaryauthorizationpolicyv1alpha1.GcpBinaryAuthorizationPolicyStackInput) *Locals {
	locals := &Locals{}
	locals.GcpBinaryAuthorizationPolicy = stackInput.Target

	locals.GcpProviderConfig = stackInput.ProviderConfig
	return locals
}
