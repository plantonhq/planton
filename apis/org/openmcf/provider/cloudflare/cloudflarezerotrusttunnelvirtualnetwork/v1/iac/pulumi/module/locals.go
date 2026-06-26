package module

import (
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	cloudflarezerotrusttunnelvirtualnetworkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarezerotrusttunnelvirtualnetwork/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig                *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustTunnelVirtualNetwork *cloudflarezerotrusttunnelvirtualnetworkv1.CloudflareZeroTrustTunnelVirtualNetwork
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflarezerotrusttunnelvirtualnetworkv1.CloudflareZeroTrustTunnelVirtualNetworkStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustTunnelVirtualNetwork = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
