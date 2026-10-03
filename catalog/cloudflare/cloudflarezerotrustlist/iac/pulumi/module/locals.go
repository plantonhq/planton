package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustlistv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustlist/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustList  *cloudflarezerotrustlistv1alpha1.CloudflareZeroTrustList
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustlistv1alpha1.CloudflareZeroTrustListIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustList = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
