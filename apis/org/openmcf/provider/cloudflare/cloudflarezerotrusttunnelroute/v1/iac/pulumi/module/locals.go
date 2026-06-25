package module

import (
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	cloudflarezerotrusttunnelroutev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarezerotrusttunnelroute/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig       *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustTunnelRoute *cloudflarezerotrusttunnelroutev1.CloudflareZeroTrustTunnelRoute
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflarezerotrusttunnelroutev1.CloudflareZeroTrustTunnelRouteStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustTunnelRoute = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
