package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarer2bucketv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarer2bucket/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareR2Bucket       *cloudflarer2bucketv1alpha1.CloudflareR2Bucket
}

// initializeLocals copies stack‑input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarer2bucketv1alpha1.CloudflareR2BucketIacInput) *Locals {
	locals := &Locals{}

	locals.CloudflareR2Bucket = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig

	return locals
}
