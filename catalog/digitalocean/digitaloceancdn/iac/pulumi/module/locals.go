package module

import (
	digitaloceancdnv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceancdn/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module. A CDN
// endpoint has no tag surface, so no Planton label set applies.
type Locals struct {
	DigitalOceanCdn *digitaloceancdnv1alpha1.DigitalOceanCdn
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceancdnv1alpha1.DigitalOceanCdnIacInput) *Locals {
	return &Locals{
		DigitalOceanCdn: iacInput.Target,
	}
}
