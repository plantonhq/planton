package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
	digitaloceankubernetesnodepoolv1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceankubernetesnodepool/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals aggregates handy references for the rest of the module.
type Locals struct {
	DigitalOceanProviderConfig     *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanKubernetesNodePool *digitaloceankubernetesnodepoolv1.DigitalOceanKubernetesNodePool
	DigitalOceanLabels             map[string]string
}

// initializeLocals mirrors the pattern used in digital_ocean_vpc.
func initializeLocals(_ *pulumi.Context, stackInput *digitaloceankubernetesnodepoolv1.DigitalOceanKubernetesNodePoolStackInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanKubernetesNodePool = stackInput.Target

	// Standard Planton labels for DigitalOcean resources.
	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanKubernetesNodePool.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_DigitalOceanKubernetesNodePool.String(),
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

	locals.DigitalOceanProviderConfig = stackInput.ProviderConfig

	return locals
}
