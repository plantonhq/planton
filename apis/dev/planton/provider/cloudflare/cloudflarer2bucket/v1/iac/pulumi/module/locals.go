package module

import (
	cloudflareprovider "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare"
	cloudflarer2bucketv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarer2bucket/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareR2Bucket       *cloudflarer2bucketv1.CloudflareR2Bucket
}

// initializeLocals copies stack‑input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, stackInput *cloudflarer2bucketv1.CloudflareR2BucketStackInput) *Locals {
	locals := &Locals{}

	locals.CloudflareR2Bucket = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig

	return locals
}
