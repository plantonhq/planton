package module

import (
	cloudflareprovider "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare"
	cloudflarehyperdriveconfigv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarehyperdriveconfig/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig   *cloudflareprovider.CloudflareProviderConfig
	CloudflareHyperdriveConfig *cloudflarehyperdriveconfigv1.CloudflareHyperdriveConfig
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflarehyperdriveconfigv1.CloudflareHyperdriveConfigStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareHyperdriveConfig = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
