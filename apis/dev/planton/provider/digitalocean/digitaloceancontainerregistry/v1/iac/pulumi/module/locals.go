package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
	digitaloceancontainerregistryv1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceancontainerregistry/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	DigitalOceanProviderConfig    *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanContainerRegistry *digitaloceancontainerregistryv1.DigitalOceanContainerRegistry
	DigitalOceanLabels            map[string]string
}

// initializeLocals copies stack‑input fields into the Locals struct and builds
// a reusable label map.
func initializeLocals(_ *pulumi.Context, stackInput *digitaloceancontainerregistryv1.DigitalOceanContainerRegistryStackInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanContainerRegistry = stackInput.Target

	// Standard Planton labels for DigitalOcean resources.
	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanContainerRegistry.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_DigitalOceanContainerRegistry.String(),
	}

	if locals.DigitalOceanContainerRegistry.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanContainerRegistry.Metadata.Org
	}

	if locals.DigitalOceanContainerRegistry.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanContainerRegistry.Metadata.Env
	}

	if locals.DigitalOceanContainerRegistry.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanContainerRegistry.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = stackInput.ProviderConfig

	return locals
}
