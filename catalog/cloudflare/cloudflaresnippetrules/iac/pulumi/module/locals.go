package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflaresnippetrulesv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflaresnippetrules/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareSnippetRules   *cloudflaresnippetrulesv1alpha1.CloudflareSnippetRules
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflaresnippetrulesv1alpha1.CloudflareSnippetRulesIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareSnippetRules = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
