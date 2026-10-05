package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflaremtlscertificatev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflaremtlscertificate/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig  *cloudflareprovider.CloudflareProviderConfig
	CloudflareMtlsCertificate *cloudflaremtlscertificatev1alpha1.CloudflareMtlsCertificate
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflaremtlscertificatev1alpha1.CloudflareMtlsCertificateIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareMtlsCertificate = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
