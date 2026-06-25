package module

import (
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	cloudflareorigincacertificatev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareorigincacertificate/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig      *cloudflareprovider.CloudflareProviderConfig
	CloudflareOriginCaCertificate *cloudflareorigincacertificatev1.CloudflareOriginCaCertificate
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflareorigincacertificatev1.CloudflareOriginCaCertificateStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareOriginCaCertificate = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
