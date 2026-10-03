package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarecertificatepackv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarecertificatepack/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig  *cloudflareprovider.CloudflareProviderConfig
	CloudflareCertificatePack *cloudflarecertificatepackv1alpha1.CloudflareCertificatePack
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarecertificatepackv1alpha1.CloudflareCertificatePackIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareCertificatePack = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
