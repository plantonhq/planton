package module

import (
	digitaloceandatabasekafkaschemav1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandatabasekafkaschema/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module. The schema
// registry resource has no tag surface, so no Planton label set applies.
type Locals struct {
	DigitalOceanDatabaseKafkaSchema *digitaloceandatabasekafkaschemav1alpha1.DigitalOceanDatabaseKafkaSchema
}

// initializeLocals copies iac-input fields into the Locals struct.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceandatabasekafkaschemav1alpha1.DigitalOceanDatabaseKafkaSchemaIacInput) *Locals {
	return &Locals{
		DigitalOceanDatabaseKafkaSchema: iacInput.Target,
	}
}
