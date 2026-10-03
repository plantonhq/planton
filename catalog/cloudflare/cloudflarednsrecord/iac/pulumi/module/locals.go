package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarednsrecordv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarednsrecord/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareDnsRecord      *cloudflarednsrecordv1alpha1.CloudflareDnsRecord
}

// initializeLocals copies fields from the IaC input into Locals.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarednsrecordv1alpha1.CloudflareDnsRecordIacInput) *Locals {
	locals := &Locals{}

	locals.CloudflareDnsRecord = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig

	return locals
}
