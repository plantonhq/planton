package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustaccessservicetokenv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustaccessservicetoken/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig              *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustAccessServiceToken *cloudflarezerotrustaccessservicetokenv1alpha1.CloudflareZeroTrustAccessServiceToken
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustaccessservicetokenv1alpha1.CloudflareZeroTrustAccessServiceTokenIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustAccessServiceToken = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
