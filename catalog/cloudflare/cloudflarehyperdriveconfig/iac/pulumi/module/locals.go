package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarehyperdriveconfigv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarehyperdriveconfig/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig   *cloudflareprovider.CloudflareProviderConfig
	CloudflareHyperdriveConfig *cloudflarehyperdriveconfigv1alpha1.CloudflareHyperdriveConfig
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarehyperdriveconfigv1alpha1.CloudflareHyperdriveConfigIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareHyperdriveConfig = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
