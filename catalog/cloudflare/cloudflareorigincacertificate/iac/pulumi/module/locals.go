package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareorigincacertificatev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareorigincacertificate/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig      *cloudflareprovider.CloudflareProviderConfig
	CloudflareOriginCaCertificate *cloudflareorigincacertificatev1alpha1.CloudflareOriginCaCertificate
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflareorigincacertificatev1alpha1.CloudflareOriginCaCertificateIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareOriginCaCertificate = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
