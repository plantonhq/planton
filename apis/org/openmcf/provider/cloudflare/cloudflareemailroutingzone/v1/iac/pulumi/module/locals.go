package module

import (
	cloudflareprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare"
	cloudflareemailroutingzonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflareemailroutingzone/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig   *cloudflareprovider.CloudflareProviderConfig
	CloudflareEmailRoutingZone *cloudflareemailroutingzonev1.CloudflareEmailRoutingZone
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflareemailroutingzonev1.CloudflareEmailRoutingZoneStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflareEmailRoutingZone = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
