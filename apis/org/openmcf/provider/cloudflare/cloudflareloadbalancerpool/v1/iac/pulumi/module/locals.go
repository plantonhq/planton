package module

import (
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	cloudflareloadbalancerpoolv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareloadbalancerpool/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig   *cloudflareprovider.CloudflareProviderConfig
	CloudflareLoadBalancerPool *cloudflareloadbalancerpoolv1.CloudflareLoadBalancerPool
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflareloadbalancerpoolv1.CloudflareLoadBalancerPoolStackInput) *Locals {
	return &Locals{
		CloudflareProviderConfig:   stackInput.ProviderConfig,
		CloudflareLoadBalancerPool: stackInput.Target,
	}
}
