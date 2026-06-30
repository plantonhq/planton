package module

import (
	"strconv"

	hetznercloudprovider "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud"
	hetznercloudprimaryipv1 "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud/hetznercloudprimaryip/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/hcloudlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	HetznerCloudProviderConfig *hetznercloudprovider.HetznerCloudProviderConfig
	HetznerCloudPrimaryIp      *hetznercloudprimaryipv1.HetznerCloudPrimaryIp
	Labels                     map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *hetznercloudprimaryipv1.HetznerCloudPrimaryIpStackInput) *Locals {
	locals := &Locals{}

	locals.HetznerCloudPrimaryIp = stackInput.Target
	locals.HetznerCloudProviderConfig = stackInput.ProviderConfig

	locals.Labels = map[string]string{
		hcloudlabelkeys.Resource:     strconv.FormatBool(true),
		hcloudlabelkeys.ResourceName: locals.HetznerCloudPrimaryIp.Metadata.Name,
		hcloudlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_HetznerCloudPrimaryIp.String(),
	}

	if locals.HetznerCloudPrimaryIp.Metadata.Org != "" {
		locals.Labels[hcloudlabelkeys.Organization] = locals.HetznerCloudPrimaryIp.Metadata.Org
	}

	if locals.HetznerCloudPrimaryIp.Metadata.Env != "" {
		locals.Labels[hcloudlabelkeys.Environment] = locals.HetznerCloudPrimaryIp.Metadata.Env
	}

	if locals.HetznerCloudPrimaryIp.Metadata.Id != "" {
		locals.Labels[hcloudlabelkeys.ResourceId] = locals.HetznerCloudPrimaryIp.Metadata.Id
	}

	for k, v := range locals.HetznerCloudPrimaryIp.Metadata.Labels {
		if _, exists := locals.Labels[k]; !exists {
			locals.Labels[k] = v
		}
	}

	return locals
}
