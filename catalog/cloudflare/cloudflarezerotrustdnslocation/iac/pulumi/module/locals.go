package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustdnslocationv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustdnslocation/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig       *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustDnsLocation *cloudflarezerotrustdnslocationv1alpha1.CloudflareZeroTrustDnsLocation
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustdnslocationv1alpha1.CloudflareZeroTrustDnsLocationIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustDnsLocation = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
