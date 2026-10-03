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

func initializeLocals(_ *pulumi.Context, iacInput *gcpbinaryauthorizationpolicyv1alpha1.GcpBinaryAuthorizationPolicyIacInput) *Locals {
	locals := &Locals{}
	locals.GcpBinaryAuthorizationPolicy = iacInput.Target

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}
