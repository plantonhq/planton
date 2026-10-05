package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustdevicecustomprofilev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustdevicecustomprofile/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig               *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustDeviceCustomProfile *cloudflarezerotrustdevicecustomprofilev1alpha1.CloudflareZeroTrustDeviceCustomProfile
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustdevicecustomprofilev1alpha1.CloudflareZeroTrustDeviceCustomProfileIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustDeviceCustomProfile = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
