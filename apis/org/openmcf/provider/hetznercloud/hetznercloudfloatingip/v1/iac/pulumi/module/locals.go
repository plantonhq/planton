package module

import (
	"strconv"

	hetznercloudprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud"
	hetznercloudfloatingipv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud/hetznercloudfloatingip/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/hcloudlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	HetznerCloudProviderConfig *hetznercloudprovider.HetznerCloudProviderConfig
	HetznerCloudFloatingIp     *hetznercloudfloatingipv1.HetznerCloudFloatingIp
	Labels                     map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *hetznercloudfloatingipv1.HetznerCloudFloatingIpStackInput) *Locals {
	locals := &Locals{}

	locals.HetznerCloudFloatingIp = stackInput.Target
	locals.HetznerCloudProviderConfig = stackInput.ProviderConfig

	locals.Labels = map[string]string{
		hcloudlabelkeys.Resource:     strconv.FormatBool(true),
		hcloudlabelkeys.ResourceName: locals.HetznerCloudFloatingIp.Metadata.Name,
		hcloudlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_HetznerCloudFloatingIp.String(),
	}

	if locals.HetznerCloudFloatingIp.Metadata.Org != "" {
		locals.Labels[hcloudlabelkeys.Organization] = locals.HetznerCloudFloatingIp.Metadata.Org
	}

	if locals.HetznerCloudFloatingIp.Metadata.Env != "" {
		locals.Labels[hcloudlabelkeys.Environment] = locals.HetznerCloudFloatingIp.Metadata.Env
	}

	if locals.HetznerCloudFloatingIp.Metadata.Id != "" {
		locals.Labels[hcloudlabelkeys.ResourceId] = locals.HetznerCloudFloatingIp.Metadata.Id
	}

	for k, v := range locals.HetznerCloudFloatingIp.Metadata.Labels {
		if _, exists := locals.Labels[k]; !exists {
			locals.Labels[k] = v
		}
	}

	return locals
}
