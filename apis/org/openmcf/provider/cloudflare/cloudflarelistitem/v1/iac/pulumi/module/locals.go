package module

import (
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	cloudflarelistitemv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarelistitem/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareListItem       *cloudflarelistitemv1.CloudflareListItem
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflarelistitemv1.CloudflareListItemStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareListItem = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
