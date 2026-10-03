package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflaresecretsstorev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflaresecretsstore/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareSecretsStore   *cloudflaresecretsstorev1alpha1.CloudflareSecretsStore
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflaresecretsstorev1alpha1.CloudflareSecretsStoreIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareSecretsStore = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
