package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustaccesspolicyv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustaccesspolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig        *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustAccessPolicy *cloudflarezerotrustaccesspolicyv1alpha1.CloudflareZeroTrustAccessPolicy
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustaccesspolicyv1alpha1.CloudflareZeroTrustAccessPolicyIacInput) *Locals {
	return &Locals{
		CloudflareProviderConfig:        iacInput.ProviderConfig,
		CloudflareZeroTrustAccessPolicy: iacInput.Target,
	}
}
