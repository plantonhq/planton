package module

import (
	digitaloceanspaceskeyv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanspaceskey/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module. A Spaces key
// has no tag surface, so no Planton label set applies.
type Locals struct {
	DigitalOceanSpacesKey *digitaloceanspaceskeyv1alpha1.DigitalOceanSpacesKey
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceanspaceskeyv1alpha1.DigitalOceanSpacesKeyIacInput) *Locals {
	return &Locals{
		DigitalOceanSpacesKey: iacInput.Target,
	}
}
