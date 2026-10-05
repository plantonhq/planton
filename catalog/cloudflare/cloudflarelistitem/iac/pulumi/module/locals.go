package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarelistitemv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarelistitem/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareListItem       *cloudflarelistitemv1alpha1.CloudflareListItem
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarelistitemv1alpha1.CloudflareListItemIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareListItem = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
