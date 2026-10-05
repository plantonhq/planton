package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflaresnippetv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflaresnippet/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareSnippet        *cloudflaresnippetv1alpha1.CloudflareSnippet
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflaresnippetv1alpha1.CloudflareSnippetIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareSnippet = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
