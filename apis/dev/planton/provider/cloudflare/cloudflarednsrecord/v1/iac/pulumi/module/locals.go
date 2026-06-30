package module

import (
	cloudflareprovider "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare"
	cloudflarednsrecordv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarednsrecord/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles the data we need throughout the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareDnsRecord      *cloudflarednsrecordv1.CloudflareDnsRecord
}

// initializeLocals copies fields from the stack input into Locals.
func initializeLocals(_ *pulumi.Context, stackInput *cloudflarednsrecordv1.CloudflareDnsRecordStackInput) *Locals {
	locals := &Locals{}

	locals.CloudflareDnsRecord = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig

	return locals
}
