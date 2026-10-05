package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareaccountapitokenv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareaccountapitoken/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig  *cloudflareprovider.CloudflareProviderConfig
	CloudflareAccountApiToken *cloudflareaccountapitokenv1alpha1.CloudflareAccountApiToken
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflareaccountapitokenv1alpha1.CloudflareAccountApiTokenIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareAccountApiToken = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
