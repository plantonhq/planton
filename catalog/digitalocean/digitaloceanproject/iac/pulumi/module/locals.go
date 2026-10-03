package module

import (
	digitaloceanprojectv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanproject/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module. A project has
// no tag surface, so no Planton label set applies here.
type Locals struct {
	DigitalOceanProject *digitaloceanprojectv1alpha1.DigitalOceanProject
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceanprojectv1alpha1.DigitalOceanProjectIacInput) *Locals {
	return &Locals{
		DigitalOceanProject: iacInput.Target,
	}
}
