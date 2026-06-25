package module

import (
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	cloudflarecustomhostnamev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarecustomhostname/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareCustomHostname *cloudflarecustomhostnamev1.CloudflareCustomHostname
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflarecustomhostnamev1.CloudflareCustomHostnameStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareCustomHostname = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
