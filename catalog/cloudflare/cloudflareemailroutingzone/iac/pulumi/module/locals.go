package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareemailroutingzonev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareemailroutingzone/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig   *cloudflareprovider.CloudflareProviderConfig
	CloudflareEmailRoutingZone *cloudflareemailroutingzonev1alpha1.CloudflareEmailRoutingZone
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflareemailroutingzonev1alpha1.CloudflareEmailRoutingZoneIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareEmailRoutingZone = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
