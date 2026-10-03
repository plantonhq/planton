package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareworkerskvpairv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareworkerskvpair/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareWorkersKvPair  *cloudflareworkerskvpairv1alpha1.CloudflareWorkersKvPair
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflareworkerskvpairv1alpha1.CloudflareWorkersKvPairIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareWorkersKvPair = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
