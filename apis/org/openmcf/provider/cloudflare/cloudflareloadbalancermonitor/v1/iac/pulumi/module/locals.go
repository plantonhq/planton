package module

import (
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	cloudflareloadbalancermonitorv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareloadbalancermonitor/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig      *cloudflareprovider.CloudflareProviderConfig
	CloudflareLoadBalancerMonitor *cloudflareloadbalancermonitorv1.CloudflareLoadBalancerMonitor
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflareloadbalancermonitorv1.CloudflareLoadBalancerMonitorStackInput) *Locals {
	return &Locals{
		CloudflareProviderConfig:      stackInput.ProviderConfig,
		CloudflareLoadBalancerMonitor: stackInput.Target,
	}
}
