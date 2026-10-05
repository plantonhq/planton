package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarecachesettingsv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarecachesettings/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareCacheSettings  *cloudflarecachesettingsv1alpha1.CloudflareCacheSettings
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarecachesettingsv1alpha1.CloudflareCacheSettingsIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareCacheSettings = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
