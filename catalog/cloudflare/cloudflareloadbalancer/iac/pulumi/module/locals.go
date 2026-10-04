package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareloadbalancerv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareloadbalancer/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals stores quick references for metadata, spec & credentials.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareLoadBalancer   *cloudflareloadbalancerv1alpha1.CloudflareLoadBalancer
}

// initializeLocals copies relevant IaC input fields into Locals.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflareloadbalancerv1alpha1.CloudflareLoadBalancerIacInput) *Locals {
	return &Locals{
		CloudflareProviderConfig: iacInput.ProviderConfig,
		CloudflareLoadBalancer:   iacInput.Target,
	}
}
