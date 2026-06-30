package module

import (
	cloudflareprovider "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare"
	cloudflarepagesprojectv1 "github.com/plantonhq/planton/apis/dev/planton/provider/cloudflare/cloudflarepagesproject/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflarePagesProject   *cloudflarepagesprojectv1.CloudflarePagesProject
}

func initializeLocals(_ *pulumi.Context, stackInput *cloudflarepagesprojectv1.CloudflarePagesProjectStackInput) *Locals {
	locals := &Locals{}
	locals.CloudflarePagesProject = stackInput.Target
	locals.CloudflareProviderConfig = stackInput.ProviderConfig
	return locals
}
