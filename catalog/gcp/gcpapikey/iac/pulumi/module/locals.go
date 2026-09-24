package module

import (
	gcpapikeyv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpapikey/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the Terraform module's locals {} convention. An API key
// carries no labels and its only name is the spec's key_id, so the only
// local is the resolved target.
type Locals struct {
	GcpApiKey *gcpapikeyv1alpha1.GcpApiKey
}

func initializeLocals(_ *pulumi.Context, stackInput *gcpapikeyv1alpha1.GcpApiKeyStackInput) *Locals {
	return &Locals{
		GcpApiKey: stackInput.Target,
	}
}
