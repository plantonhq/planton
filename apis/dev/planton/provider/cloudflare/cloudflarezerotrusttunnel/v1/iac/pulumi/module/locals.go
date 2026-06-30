package module

import (
	cloudflareprovider "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare"
	cloudflarezerotrusttunnelv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarezerotrusttunnel/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig  *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustTunnel *cloudflarezerotrusttunnelv1.CloudflareZeroTrustTunnel
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflarezerotrusttunnelv1.CloudflareZeroTrustTunnelStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustTunnel = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
