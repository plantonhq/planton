package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareipaccessrulev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareipaccessrule/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareIpAccessRule   *cloudflareipaccessrulev1alpha1.CloudflareIpAccessRule
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflareipaccessrulev1alpha1.CloudflareIpAccessRuleIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareIpAccessRule = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
