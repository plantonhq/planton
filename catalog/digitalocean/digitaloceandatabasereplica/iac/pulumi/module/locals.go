package module

import (
	"strconv"

	digitaloceandatabasereplicav1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandatabasereplica/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	DigitalOceanDatabaseReplica *digitaloceandatabasereplicav1alpha1.DigitalOceanDatabaseReplica
	DigitalOceanLabels          map[string]string
}

// initializeLocals copies iac-input fields into the Locals struct and
// builds the standard Planton label map (rendered as "key:value" tags on
// the replica -- the identical set the Terraform module applies).
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceandatabasereplicav1alpha1.DigitalOceanDatabaseReplicaIacInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanDatabaseReplica = iacInput.Target

	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanDatabaseReplica.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: catalogkind.CatalogKind_DigitalOceanDatabaseReplica.String(),
	}

	if locals.DigitalOceanDatabaseReplica.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanDatabaseReplica.Metadata.Org
	}

	if locals.DigitalOceanDatabaseReplica.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanDatabaseReplica.Metadata.Env
	}

	if locals.DigitalOceanDatabaseReplica.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanDatabaseReplica.Metadata.Id
	}

	return locals
}
