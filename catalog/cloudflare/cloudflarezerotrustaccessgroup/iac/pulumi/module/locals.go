package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustaccessgroupv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustaccessgroup/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig       *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustAccessGroup *cloudflarezerotrustaccessgroupv1alpha1.CloudflareZeroTrustAccessGroup
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustaccessgroupv1alpha1.CloudflareZeroTrustAccessGroupIacInput) *Locals {
	return &Locals{
		CloudflareProviderConfig:       iacInput.ProviderConfig,
		CloudflareZeroTrustAccessGroup: iacInput.Target,
	}
}
