package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezonesettingsv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezonesettings/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareZoneSettings   *cloudflarezonesettingsv1alpha1.CloudflareZoneSettings
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezonesettingsv1alpha1.CloudflareZoneSettingsIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZoneSettings = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
