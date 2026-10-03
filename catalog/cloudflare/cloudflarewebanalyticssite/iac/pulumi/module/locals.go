package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarewebanalyticssitev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarewebanalyticssite/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig   *cloudflareprovider.CloudflareProviderConfig
	CloudflareWebAnalyticsSite *cloudflarewebanalyticssitev1alpha1.CloudflareWebAnalyticsSite
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarewebanalyticssitev1alpha1.CloudflareWebAnalyticsSiteIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareWebAnalyticsSite = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
