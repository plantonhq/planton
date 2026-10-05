package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustdeviceposturerulev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustdeviceposturerule/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig             *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustDevicePostureRule *cloudflarezerotrustdeviceposturerulev1alpha1.CloudflareZeroTrustDevicePostureRule
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustdeviceposturerulev1alpha1.CloudflareZeroTrustDevicePostureRuleIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustDevicePostureRule = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
