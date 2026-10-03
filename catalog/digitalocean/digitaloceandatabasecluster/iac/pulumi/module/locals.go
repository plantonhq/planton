package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	digitaloceandatabaseclusterv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceandatabasecluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	DigitalOceanProviderConfig  *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanDatabaseCluster *digitaloceandatabaseclusterv1alpha1.DigitalOceanDatabaseCluster
	DigitalOceanLabels          map[string]string
}

// initializeLocals copies stack‑input fields into the Locals struct and builds
// a reusable label map. Mirrors the pattern of the DigitalOcean VPC module.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceandatabaseclusterv1alpha1.DigitalOceanDatabaseClusterIacInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanDatabaseCluster = iacInput.Target

	// Standard Planton labels for DigitalOcean resources.
	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanDatabaseCluster.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: catalogkind.CatalogKind_DigitalOceanDatabaseCluster.String(),
	}

	if locals.DigitalOceanDatabaseCluster.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanDatabaseCluster.Metadata.Org
	}

	if locals.DigitalOceanDatabaseCluster.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanDatabaseCluster.Metadata.Env
	}

	if locals.DigitalOceanDatabaseCluster.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanDatabaseCluster.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = iacInput.ProviderConfig

	return locals
}
