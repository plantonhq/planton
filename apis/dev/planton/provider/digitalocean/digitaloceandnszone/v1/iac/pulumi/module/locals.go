package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
	digitaloceandnszonev1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceandnszone/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds quick references used by other files.
type Locals struct {
	DigitalOceanProviderConfig *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanDnsZone        *digitaloceandnszonev1.DigitalOceanDnsZone
	DigitalOceanLabels         map[string]string
}

// initializeLocals mirrors the pattern from the VPC module.
func initializeLocals(_ *pulumi.Context, stackInput *digitaloceandnszonev1.DigitalOceanDnsZoneStackInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanDnsZone = stackInput.Target

	// Standard Planton labels for DigitalOcean resources.
	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanDnsZone.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_DigitalOceanDnsZone.String(),
	}

	if locals.DigitalOceanDnsZone.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanDnsZone.Metadata.Org
	}
	if locals.DigitalOceanDnsZone.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanDnsZone.Metadata.Env
	}
	if locals.DigitalOceanDnsZone.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanDnsZone.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = stackInput.ProviderConfig

	return locals
}
