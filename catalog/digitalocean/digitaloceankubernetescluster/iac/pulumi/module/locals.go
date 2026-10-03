package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	digitaloceankubernetesclusterv1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceankubernetescluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	DigitalOceanProviderConfig    *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanKubernetesCluster *digitaloceankubernetesclusterv1alpha1.DigitalOceanKubernetesCluster
	DigitalOceanLabels            map[string]string
}

// initializeLocals copies stack‑input fields into the Locals struct and builds
// a reusable label map—mirrors the style used in digital_ocean_vpc.
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceankubernetesclusterv1alpha1.DigitalOceanKubernetesClusterIacInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanKubernetesCluster = iacInput.Target

	// Standard Planton labels for DigitalOcean resources.
	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanKubernetesCluster.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: catalogkind.CatalogKind_DigitalOceanKubernetesCluster.String(),
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

	locals.DigitalOceanProviderConfig = iacInput.ProviderConfig

	return locals
}
