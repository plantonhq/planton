package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarebotmanagementv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarebotmanagement/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareBotManagement  *cloudflarebotmanagementv1alpha1.CloudflareBotManagement
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarebotmanagementv1alpha1.CloudflareBotManagementIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareBotManagement = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
