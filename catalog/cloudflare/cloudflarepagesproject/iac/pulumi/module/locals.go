package module

import (
	cloudflareprovider "github.com/plantonhq/planton/catalog/cloudflare"
	cloudflarepagesprojectv1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflarepagesproject/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	CloudflareProviderConfig *cloudflareprovider.CloudflareProviderConfig
	CloudflarePagesProject   *cloudflarepagesprojectv1alpha1.CloudflarePagesProject
}

func initializeLocals(_ *pulumi.Context, iacInput *cloudflarepagesprojectv1alpha1.CloudflarePagesProjectIacInput) *Locals {
	locals := &Locals{}
	locals.CloudflarePagesProject = iacInput.Target
	locals.CloudflareProviderConfig = iacInput.ProviderConfig
	return locals
}
