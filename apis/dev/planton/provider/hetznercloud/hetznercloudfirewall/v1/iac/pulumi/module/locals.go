package module

import (
	"strconv"

	hetznercloudprovider "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud"
	hetznercloudfirewallv1 "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud/hetznercloudfirewall/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/hcloudlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	HetznerCloudProviderConfig *hetznercloudprovider.HetznerCloudProviderConfig
	HetznerCloudFirewall       *hetznercloudfirewallv1.HetznerCloudFirewall
	Labels                     map[string]string
}

// initializeLocals copies stack-input fields into the Locals struct and builds
// a reusable label map.
func initializeLocals(_ *pulumi.Context, stackInput *hetznercloudfirewallv1.HetznerCloudFirewallStackInput) *Locals {
	locals := &Locals{}

	locals.HetznerCloudFirewall = stackInput.Target
	locals.HetznerCloudProviderConfig = stackInput.ProviderConfig

	locals.Labels = map[string]string{
		hcloudlabelkeys.Resource:     strconv.FormatBool(true),
		hcloudlabelkeys.ResourceName: locals.HetznerCloudFirewall.Metadata.Name,
		hcloudlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_HetznerCloudFirewall.String(),
	}

	if locals.HetznerCloudFirewall.Metadata.Org != "" {
		locals.Labels[hcloudlabelkeys.Organization] = locals.HetznerCloudFirewall.Metadata.Org
	}

	if locals.HetznerCloudFirewall.Metadata.Env != "" {
		locals.Labels[hcloudlabelkeys.Environment] = locals.HetznerCloudFirewall.Metadata.Env
	}

	if locals.HetznerCloudFirewall.Metadata.Id != "" {
		locals.Labels[hcloudlabelkeys.ResourceId] = locals.HetznerCloudFirewall.Metadata.Id
	}

	// Merge user-specified metadata labels; standard labels take precedence.
	for k, v := range locals.HetznerCloudFirewall.Metadata.Labels {
		if _, exists := locals.Labels[k]; !exists {
			locals.Labels[k] = v
		}
	}

	return locals
}
