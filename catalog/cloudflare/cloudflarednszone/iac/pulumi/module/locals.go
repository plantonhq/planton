package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarednszonev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarednszone/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the bits we need everywhere else.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareDnsZone        *cloudflarednszonev1alpha1.CloudflareDnsZone
}

// initializeLocals copies fields from the stack‑input into Locals.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarednszonev1alpha1.CloudflareDnsZoneIacInput) *Locals {
	locals := &Locals{}

	locals.CloudflareDnsZone = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig

	return locals
}
