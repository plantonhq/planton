package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
	digitaloceanappplatformservicev1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceanappplatformservice/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals gathers convenient references for the rest of the module.
type Locals struct {
	DigitalOceanProviderConfig     *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanAppPlatformService *digitaloceanappplatformservicev1.DigitalOceanAppPlatformService
	DigitalOceanLabels             map[string]string
}

// initializeLocals mirrors the pattern used in digital_ocean_vpc's initializeLocals().
func initializeLocals(_ *pulumi.Context, stackInput *digitaloceanappplatformservicev1.DigitalOceanAppPlatformServiceStackInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanAppPlatformService = stackInput.Target

	// Standard Planton labels for DigitalOcean resources.
	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanAppPlatformService.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_DigitalOceanAppPlatformService.String(),
	}

	if locals.DigitalOceanAppPlatformService.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanAppPlatformService.Metadata.Org
	}

	if locals.DigitalOceanAppPlatformService.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanAppPlatformService.Metadata.Env
	}

	if locals.DigitalOceanAppPlatformService.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanAppPlatformService.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = stackInput.ProviderConfig

	return locals
}
