package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrusttunnelv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrusttunnel/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig  *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustTunnel *cloudflarezerotrusttunnelv1alpha1.CloudflareZeroTrustTunnel
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrusttunnelv1alpha1.CloudflareZeroTrustTunnelIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustTunnel = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
