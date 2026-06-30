package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
	digitaloceankubernetesclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceankubernetescluster/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	DigitalOceanProviderConfig    *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanKubernetesCluster *digitaloceankubernetesclusterv1.DigitalOceanKubernetesCluster
	DigitalOceanLabels            map[string]string
}

// initializeLocals copies stack‑input fields into the Locals struct and builds
// a reusable label map—mirrors the style used in digital_ocean_vpc.
func initializeLocals(_ *pulumi.Context, stackInput *digitaloceankubernetesclusterv1.DigitalOceanKubernetesClusterStackInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanKubernetesCluster = stackInput.Target

	// Standard Planton labels for DigitalOcean resources.
	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanKubernetesCluster.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_DigitalOceanKubernetesCluster.String(),
	}

	if locals.DigitalOceanKubernetesCluster.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanKubernetesCluster.Metadata.Org
	}

	if locals.DigitalOceanKubernetesCluster.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanKubernetesCluster.Metadata.Env
	}

	if locals.DigitalOceanKubernetesCluster.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanKubernetesCluster.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = stackInput.ProviderConfig

	return locals
}
