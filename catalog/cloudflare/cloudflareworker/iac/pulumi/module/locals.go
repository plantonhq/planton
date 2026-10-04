package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareworkerv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareworker/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles quick references copied from the IaC input.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareWorker         *cloudflareworkerv1alpha1.CloudflareWorker
}

// initializeLocals mirrors the pattern used in existing modules.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflareworkerv1alpha1.CloudflareWorkerIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareWorker = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
