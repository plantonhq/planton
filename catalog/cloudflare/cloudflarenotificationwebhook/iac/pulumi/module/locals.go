package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarenotificationwebhookv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarenotificationwebhook/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig      *cloudflareprovider.CloudflareProviderConfig
	CloudflareNotificationWebhook *cloudflarenotificationwebhookv1alpha1.CloudflareNotificationWebhook
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarenotificationwebhookv1alpha1.CloudflareNotificationWebhookIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareNotificationWebhook = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
