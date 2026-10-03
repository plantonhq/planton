package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustaccessidentityproviderv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustaccessidentityprovider/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig                  *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustAccessIdentityProvider *cloudflarezerotrustaccessidentityproviderv1alpha1.CloudflareZeroTrustAccessIdentityProvider
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustaccessidentityproviderv1alpha1.CloudflareZeroTrustAccessIdentityProviderIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustAccessIdentityProvider = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
