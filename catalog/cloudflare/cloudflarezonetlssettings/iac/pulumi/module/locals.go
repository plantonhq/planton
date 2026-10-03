package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezonetlssettingsv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezonetlssettings/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig  *cloudflareprovider.CloudflareProviderConfig
	CloudflareZoneTlsSettings *cloudflarezonetlssettingsv1alpha1.CloudflareZoneTlsSettings
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezonetlssettingsv1alpha1.CloudflareZoneTlsSettingsIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZoneTlsSettings = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
