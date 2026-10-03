package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarequeuev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarequeue/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareQueue          *cloudflarequeuev1alpha1.CloudflareQueue
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarequeuev1alpha1.CloudflareQueueIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareQueue = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
