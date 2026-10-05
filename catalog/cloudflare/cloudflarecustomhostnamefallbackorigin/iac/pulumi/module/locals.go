package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarecustomhostnamefallbackoriginv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarecustomhostnamefallbackorigin/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig               *cloudflareprovider.CloudflareProviderConfig
	CloudflareCustomHostnameFallbackOrigin *cloudflarecustomhostnamefallbackoriginv1alpha1.CloudflareCustomHostnameFallbackOrigin
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarecustomhostnamefallbackoriginv1alpha1.CloudflareCustomHostnameFallbackOriginIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareCustomHostnameFallbackOrigin = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
