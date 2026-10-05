package module

import (
	digitaloceandatabasedbv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandatabasedb/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module. The logical
// database resource has no tag surface, so no Planton label set applies.
type Locals struct {
	DigitalOceanDatabaseDb *digitaloceandatabasedbv1alpha1.DigitalOceanDatabaseDb
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceandatabasedbv1alpha1.DigitalOceanDatabaseDbIacInput) *Locals {
	return &Locals{
		DigitalOceanDatabaseDb: iacInput.Target,
	}
}
