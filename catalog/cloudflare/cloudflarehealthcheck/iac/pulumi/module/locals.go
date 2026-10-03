package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarehealthcheckv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarehealthcheck/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareHealthcheck    *cloudflarehealthcheckv1alpha1.CloudflareHealthcheck
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarehealthcheckv1alpha1.CloudflareHealthcheckIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareHealthcheck = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
