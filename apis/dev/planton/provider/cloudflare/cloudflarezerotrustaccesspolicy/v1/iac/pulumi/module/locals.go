package module

import (
	cloudflareprovider "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare"
	cloudflarezerotrustaccesspolicyv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarezerotrustaccesspolicy/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig        *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustAccessPolicy *cloudflarezerotrustaccesspolicyv1.CloudflareZeroTrustAccessPolicy
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflarezerotrustaccesspolicyv1.CloudflareZeroTrustAccessPolicyStackInput) *Locals {
	return &Locals{
		CloudflareProviderConfig:        stackInput.ProviderConfig,
		CloudflareZeroTrustAccessPolicy: stackInput.Target,
	}
}
