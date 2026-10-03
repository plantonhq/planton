package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrusttunnelvirtualnetworkv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrusttunnelvirtualnetwork/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig                *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustTunnelVirtualNetwork *cloudflarezerotrusttunnelvirtualnetworkv1alpha1.CloudflareZeroTrustTunnelVirtualNetwork
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrusttunnelvirtualnetworkv1alpha1.CloudflareZeroTrustTunnelVirtualNetworkIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustTunnelVirtualNetwork = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
