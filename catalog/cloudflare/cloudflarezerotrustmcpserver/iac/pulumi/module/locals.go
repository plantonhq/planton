package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarezerotrustmcpserverv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarezerotrustmcpserver/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references used across the module.
type Locals struct {
	CloudflareProviderConfig     *cloudflareprovider.CloudflareProviderConfig
	CloudflareZeroTrustMcpServer *cloudflarezerotrustmcpserverv1alpha1.CloudflareZeroTrustMcpServer
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *cloudflarezerotrustmcpserverv1alpha1.CloudflareZeroTrustMcpServerIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflareZeroTrustMcpServer = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
