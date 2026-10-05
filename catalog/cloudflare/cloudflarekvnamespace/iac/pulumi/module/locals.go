package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarekvnamespacev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarekvnamespace/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareKvNamespace    *cloudflarekvnamespacev1alpha1.CloudflareKvNamespace
}

// initializeLocals copies IaC input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarekvnamespacev1alpha1.CloudflareKvNamespaceIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareKvNamespace = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
