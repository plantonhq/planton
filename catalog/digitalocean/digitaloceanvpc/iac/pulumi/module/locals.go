package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	digitaloceanvpcv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanvpc/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	DigitalOceanProviderConfig *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanVpc            *digitaloceanvpcv1alpha1.DigitalOceanVpc
	DigitalOceanLabels         map[string]string
}

// initializeLocals copies IaC input fields into the Locals struct and builds
// a reusable label map. Mirrors the style of gcp_vpc's initializeLocals().
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceanvpcv1alpha1.DigitalOceanVpcIacInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanVpc = iacInput.Target

	// Standard Planton labels for DigitalOcean resources.
	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanVpc.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: catalogkind.CatalogKind_DigitalOceanVpc.String(),
	}

	if locals.DigitalOceanVpc.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanVpc.Metadata.Org
	}

	if locals.DigitalOceanVpc.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanVpc.Metadata.Env
	}

	if locals.DigitalOceanVpc.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanVpc.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = iacInput.ProviderConfig

	return locals
}
