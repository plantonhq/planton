package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflaresecretsstoresecretv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflaresecretsstoresecret/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig     *cloudflareprovider.CloudflareProviderConfig
	CloudflareSecretsStoreSecret *cloudflaresecretsstoresecretv1alpha1.CloudflareSecretsStoreSecret
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflaresecretsstoresecretv1alpha1.CloudflareSecretsStoreSecretIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareSecretsStoreSecret = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
