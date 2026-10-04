package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/catalog/digitalocean"
	digitaloceanvolumev1alpha1 "github.com/plantonhq/planton/catalog/digitalocean/digitaloceanvolume/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	DigitalOceanProviderConfig *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanVolume         *digitaloceanvolumev1alpha1.DigitalOceanVolume
	DigitalOceanLabels         map[string]string
}

// initializeLocals copies IaC input fields into the Locals struct and builds
// a reusable label map. Mirrors the style of digital_ocean_vpc's initializeLocals().
func initializeLocals(_ *pulumi.Context, iacInput *digitaloceanvolumev1alpha1.DigitalOceanVolumeIacInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanVolume = iacInput.Target

	// Standard Planton labels for DigitalOcean resources.
	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanVolume.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: catalogkind.CatalogKind_DigitalOceanVolume.String(),
	}

	if locals.DigitalOceanVolume.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanVolume.Metadata.Org
	}

	if locals.DigitalOceanVolume.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanVolume.Metadata.Env
	}

	if locals.DigitalOceanVolume.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanVolume.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = iacInput.ProviderConfig

	return locals
}
