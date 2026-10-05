package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustgatewaypolicyv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustgatewaypolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig         *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustGatewayPolicy *cloudflarezerotrustgatewaypolicyv1alpha1.CloudflareZeroTrustGatewayPolicy
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustgatewaypolicyv1alpha1.CloudflareZeroTrustGatewayPolicyIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustGatewayPolicy = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
