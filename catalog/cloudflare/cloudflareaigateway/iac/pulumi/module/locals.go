package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareaigatewayv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareaigateway/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareAiGateway      *cloudflareaigatewayv1alpha1.CloudflareAiGateway
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflareaigatewayv1alpha1.CloudflareAiGatewayIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareAiGateway = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
