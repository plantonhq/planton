package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareemailroutingrulev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareemailroutingrule/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig   *cloudflareprovider.CloudflareProviderConfig
	CloudflareEmailRoutingRule *cloudflareemailroutingrulev1alpha1.CloudflareEmailRoutingRule
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflareemailroutingrulev1alpha1.CloudflareEmailRoutingRuleIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareEmailRoutingRule = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
