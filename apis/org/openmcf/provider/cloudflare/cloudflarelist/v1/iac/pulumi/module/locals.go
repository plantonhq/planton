package module

import (
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	cloudflarelistv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflarelist/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflareList           *cloudflarelistv1.CloudflareList
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflarelistv1.CloudflareListStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareList = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
