package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarenotificationpolicyv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarenotificationpolicy/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig     *cloudflareprovider.CloudflareProviderConfig
	CloudflareNotificationPolicy *cloudflarenotificationpolicyv1alpha1.CloudflareNotificationPolicy
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarenotificationpolicyv1alpha1.CloudflareNotificationPolicyIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareNotificationPolicy = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
