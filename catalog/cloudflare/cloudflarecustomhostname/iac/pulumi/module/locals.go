package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarecustomhostnamev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarecustomhostname/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareCustomHostname *cloudflarecustomhostnamev1alpha1.CloudflareCustomHostname
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarecustomhostnamev1alpha1.CloudflareCustomHostnameIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareCustomHostname = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
