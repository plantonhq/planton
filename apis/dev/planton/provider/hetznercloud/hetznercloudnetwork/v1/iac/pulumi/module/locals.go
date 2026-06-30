package module

import (
	"strconv"

	hetznercloudprovider "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud"
	hetznercloudnetworkv1 "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud/hetznercloudnetwork/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/hcloudlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	HetznerCloudProviderConfig *hetznercloudprovider.HetznerCloudProviderConfig
	HetznerCloudNetwork        *hetznercloudnetworkv1.HetznerCloudNetwork
	Labels                     map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *hetznercloudnetworkv1.HetznerCloudNetworkStackInput) *Locals {
	locals := &Locals{}

	locals.HetznerCloudNetwork = stackInput.Target
	locals.HetznerCloudProviderConfig = stackInput.ProviderConfig

	locals.Labels = map[string]string{
		hcloudlabelkeys.Resource:     strconv.FormatBool(true),
		hcloudlabelkeys.ResourceName: locals.HetznerCloudNetwork.Metadata.Name,
		hcloudlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_HetznerCloudNetwork.String(),
	}

	if locals.HetznerCloudNetwork.Metadata.Org != "" {
		locals.Labels[hcloudlabelkeys.Organization] = locals.HetznerCloudNetwork.Metadata.Org
	}

	if locals.HetznerCloudNetwork.Metadata.Env != "" {
		locals.Labels[hcloudlabelkeys.Environment] = locals.HetznerCloudNetwork.Metadata.Env
	}

	if locals.HetznerCloudNetwork.Metadata.Id != "" {
		locals.Labels[hcloudlabelkeys.ResourceId] = locals.HetznerCloudNetwork.Metadata.Id
	}

	for k, v := range locals.HetznerCloudNetwork.Metadata.Labels {
		if _, exists := locals.Labels[k]; !exists {
			locals.Labels[k] = v
		}
	}

	return locals
}
