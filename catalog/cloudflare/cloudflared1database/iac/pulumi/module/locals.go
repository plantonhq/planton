package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflared1databasev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflared1database/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareD1Database     *cloudflared1databasev1alpha1.CloudflareD1Database
}

// initializeLocals copies IaC input fields into the Locals struct.
// Mirrors the style used in other Planton modules.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflared1databasev1alpha1.CloudflareD1DatabaseIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareD1Database = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
