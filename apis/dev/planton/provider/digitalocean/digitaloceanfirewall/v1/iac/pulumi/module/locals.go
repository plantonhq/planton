package module

import (
	"strconv"

	digitaloceanprovider "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
	digitaloceanfirewallv1 "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean/digitaloceanfirewall/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/digitalocean/digitaloceanlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors the pattern used in digital_ocean_vpc.
type Locals struct {
	DigitalOceanProviderConfig *digitaloceanprovider.DigitalOceanProviderConfig
	DigitalOceanFirewall       *digitaloceanfirewallv1.DigitalOceanFirewall
	DigitalOceanLabels         map[string]string
}

// initializeLocals builds the label set and copies references we need elsewhere.
func initializeLocals(_ *pulumi.Context, stackInput *digitaloceanfirewallv1.DigitalOceanFirewallStackInput) *Locals {
	locals := &Locals{}

	locals.DigitalOceanFirewall = stackInput.Target

	locals.DigitalOceanLabels = map[string]string{
		digitaloceanlabelkeys.Resource:     strconv.FormatBool(true),
		digitaloceanlabelkeys.ResourceName: locals.DigitalOceanFirewall.Metadata.Name,
		digitaloceanlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_DigitalOceanFirewall.String(),
	}

	if locals.DigitalOceanFirewall.Metadata.Org != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Organization] = locals.DigitalOceanFirewall.Metadata.Org
	}

	if locals.DigitalOceanFirewall.Metadata.Env != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.Environment] = locals.DigitalOceanFirewall.Metadata.Env
	}

	if locals.DigitalOceanFirewall.Metadata.Id != "" {
		locals.DigitalOceanLabels[digitaloceanlabelkeys.ResourceId] = locals.DigitalOceanFirewall.Metadata.Id
	}

	locals.DigitalOceanProviderConfig = stackInput.ProviderConfig

	return locals
}
