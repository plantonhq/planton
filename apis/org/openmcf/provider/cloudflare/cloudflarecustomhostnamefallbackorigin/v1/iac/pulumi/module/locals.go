package module

import (
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	cloudflarecustomhostnamefallbackoriginv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarecustomhostnamefallbackorigin/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig               *cloudflareprovider.CloudflareProviderConfig
	CloudflareCustomHostnameFallbackOrigin *cloudflarecustomhostnamefallbackoriginv1.CloudflareCustomHostnameFallbackOrigin
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflarecustomhostnamefallbackoriginv1.CloudflareCustomHostnameFallbackOriginStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareCustomHostnameFallbackOrigin = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
