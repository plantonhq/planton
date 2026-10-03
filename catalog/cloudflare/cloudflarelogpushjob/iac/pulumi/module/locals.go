package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarelogpushjobv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarelogpushjob/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareLogpushJob     *cloudflarelogpushjobv1alpha1.CloudflareLogpushJob
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarelogpushjobv1alpha1.CloudflareLogpushJobIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareLogpushJob = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
