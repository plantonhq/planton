package module

import (
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	cloudflareemailroutingrulev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareemailroutingrule/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig   *cloudflareprovider.CloudflareProviderConfig
	CloudflareEmailRoutingRule *cloudflareemailroutingrulev1.CloudflareEmailRoutingRule
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflareemailroutingrulev1.CloudflareEmailRoutingRuleStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareEmailRoutingRule = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
