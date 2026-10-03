package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarelistv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarelist/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareList           *cloudflarelistv1alpha1.CloudflareList
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarelistv1alpha1.CloudflareListIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareList = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
