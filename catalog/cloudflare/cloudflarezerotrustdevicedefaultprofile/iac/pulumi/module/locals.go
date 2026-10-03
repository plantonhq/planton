package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustdevicedefaultprofilev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustdevicedefaultprofile/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig                *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustDeviceDefaultProfile *cloudflarezerotrustdevicedefaultprofilev1alpha1.CloudflareZeroTrustDeviceDefaultProfile
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustdevicedefaultprofilev1alpha1.CloudflareZeroTrustDeviceDefaultProfileIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustDeviceDefaultProfile = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
