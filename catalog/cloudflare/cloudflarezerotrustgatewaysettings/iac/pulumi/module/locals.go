package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustgatewaysettingsv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustgatewaysettings/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig           *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustGatewaySettings *cloudflarezerotrustgatewaysettingsv1alpha1.CloudflareZeroTrustGatewaySettings
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustgatewaysettingsv1alpha1.CloudflareZeroTrustGatewaySettingsIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustGatewaySettings = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
