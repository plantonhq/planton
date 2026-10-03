package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareauthenticatedoriginpullscertificatev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareauthenticatedoriginpullscertificate/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig                      *cloudflareprovider.CloudflareProviderConfig
	CloudflareAuthenticatedOriginPullsCertificate *cloudflareauthenticatedoriginpullscertificatev1alpha1.CloudflareAuthenticatedOriginPullsCertificate
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflareauthenticatedoriginpullscertificatev1alpha1.CloudflareAuthenticatedOriginPullsCertificateIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareAuthenticatedOriginPullsCertificate = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
