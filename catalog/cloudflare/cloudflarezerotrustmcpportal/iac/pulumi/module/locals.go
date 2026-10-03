package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustmcpportalv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustmcpportal/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig     *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustMcpPortal *cloudflarezerotrustmcpportalv1alpha1.CloudflareZeroTrustMcpPortal
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustmcpportalv1alpha1.CloudflareZeroTrustMcpPortalIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustMcpPortal = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
