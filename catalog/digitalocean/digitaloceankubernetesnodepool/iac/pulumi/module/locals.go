package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	digitaloceankubernetesnodepoolv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceankubernetesnodepool/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals aggregates handy references for the rest of the module.
type Locals struct {
	DigitalOceanProviderConfig     *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanKubernetesNodePool *digitaloceankubernetesnodepoolv1alpha1.DigitalOceanKubernetesNodePool
	DigitalOceanLabels             map[string]string
}

// initializeLocals mirrors the pattern used in digital_ocean_vpc.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceankubernetesnodepoolv1alpha1.DigitalOceanKubernetesNodePoolIacInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanKubernetesNodePool = iacInput.Target

	// Standard Planton labels for DigitalOcean resources.
	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanKubernetesNodePool.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: catalogkind.CatalogKind_DigitalOceanKubernetesNodePool.String(),
	}

	if locals.DigitalOceanKubernetesNodePool.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanKubernetesNodePool.Metadata.Org
	}

	if locals.DigitalOceanKubernetesNodePool.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanKubernetesNodePool.Metadata.Env
	}

	if locals.DigitalOceanKubernetesNodePool.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanKubernetesNodePool.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = iacInput.ProviderConfig

	return locals
}
