package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareloadbalancermonitorv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareloadbalancermonitor/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig      *cloudflareprovider.CloudflareProviderConfig
	CloudflareLoadBalancerMonitor *cloudflareloadbalancermonitorv1alpha1.CloudflareLoadBalancerMonitor
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflareloadbalancermonitorv1alpha1.CloudflareLoadBalancerMonitorIacInput) *Locals {
	return &Locals{
		CloudflareProviderConfig:      iacInput.ProviderConfig,
		CloudflareLoadBalancerMonitor: iacInput.Target,
	}
}
