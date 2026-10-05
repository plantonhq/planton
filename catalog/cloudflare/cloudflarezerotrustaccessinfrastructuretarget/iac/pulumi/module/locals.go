package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustaccessinfrastructuretargetv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustaccessinfrastructuretarget/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig                      *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustAccessInfrastructureTarget *cloudflarezerotrustaccessinfrastructuretargetv1alpha1.CloudflareZeroTrustAccessInfrastructureTarget
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustaccessinfrastructuretargetv1alpha1.CloudflareZeroTrustAccessInfrastructureTargetIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustAccessInfrastructureTarget = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
