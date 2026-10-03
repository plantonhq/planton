package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflareturnstilewidgetv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflareturnstilewidget/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig  *cloudflareprovider.CloudflareProviderConfig
	CloudflareTurnstileWidget *cloudflareturnstilewidgetv1alpha1.CloudflareTurnstileWidget
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflareturnstilewidgetv1alpha1.CloudflareTurnstileWidgetIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareTurnstileWidget = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
