package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrusttunnelroutev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrusttunnelroute/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig       *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustTunnelRoute *cloudflarezerotrusttunnelroutev1alpha1.CloudflareZeroTrustTunnelRoute
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrusttunnelroutev1alpha1.CloudflareZeroTrustTunnelRouteIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustTunnelRoute = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
