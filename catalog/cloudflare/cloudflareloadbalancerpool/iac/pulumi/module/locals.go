package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareloadbalancerpoolv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareloadbalancerpool/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig   *cloudflareprovider.CloudflareProviderConfig
	CloudflareLoadBalancerPool *cloudflareloadbalancerpoolv1alpha1.CloudflareLoadBalancerPool
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflareloadbalancerpoolv1alpha1.CloudflareLoadBalancerPoolIacInput) *Locals {
	return &Locals{
		CloudflareProviderConfig:   iacInput.ProviderConfig,
		CloudflareLoadBalancerPool: iacInput.Target,
	}
}
