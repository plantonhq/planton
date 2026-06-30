package module

import (
	cloudflareprovider "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare"
	cloudflarezerotrustaccessgroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarezerotrustaccessgroup/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig       *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustAccessGroup *cloudflarezerotrustaccessgroupv1.CloudflareZeroTrustAccessGroup
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflarezerotrustaccessgroupv1.CloudflareZeroTrustAccessGroupStackInput) *Locals {
	return &Locals{
		CloudflareProviderConfig:       stackInput.ProviderConfig,
		CloudflareZeroTrustAccessGroup: stackInput.Target,
	}
}
