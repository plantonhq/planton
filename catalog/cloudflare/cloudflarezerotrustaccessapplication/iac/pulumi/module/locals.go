package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustaccessapplicationv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustaccessapplication/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles convenient shortcuts for the rest of the module.
type Locals struct {
	CloudflareProviderConfig             *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustAccessApplication *cloudflarezerotrustaccessapplicationv1alpha1.CloudflareZeroTrustAccessApplication
}

// initializeLocals copies IaC input fields into Locals.
func initializeLocals(
	_ *pulumi.Context,
	iacInput *cloudflarezerotrustaccessapplicationv1alpha1.CloudflareZeroTrustAccessApplicationIacInput,
) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustAccessApplication = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
