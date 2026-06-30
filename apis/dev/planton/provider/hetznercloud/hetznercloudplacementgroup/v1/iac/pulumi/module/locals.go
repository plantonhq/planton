package module

import (
	"strconv"

	hetznercloudprovider "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud"
	hetznercloudplacementgroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud/hetznercloudplacementgroup/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/hcloudlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	HetznerCloudProviderConfig *hetznercloudprovider.HetznerCloudProviderConfig
	HetznerCloudPlacementGroup *hetznercloudplacementgroupv1.HetznerCloudPlacementGroup
	Labels                     map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *hetznercloudplacementgroupv1.HetznerCloudPlacementGroupStackInput) *Locals {
	locals := &Locals{}

	locals.HetznerCloudPlacementGroup = stackInput.Target
	locals.HetznerCloudProviderConfig = stackInput.ProviderConfig

	locals.Labels = map[string]string{
		hcloudlabelkeys.Resource:     strconv.FormatBool(true),
		hcloudlabelkeys.ResourceName: locals.HetznerCloudPlacementGroup.Metadata.Name,
		hcloudlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_HetznerCloudPlacementGroup.String(),
	}

	if locals.HetznerCloudPlacementGroup.Metadata.Org != "" {
		locals.Labels[hcloudlabelkeys.Organization] = locals.HetznerCloudPlacementGroup.Metadata.Org
	}

	if locals.HetznerCloudPlacementGroup.Metadata.Env != "" {
		locals.Labels[hcloudlabelkeys.Environment] = locals.HetznerCloudPlacementGroup.Metadata.Env
	}

	if locals.HetznerCloudPlacementGroup.Metadata.Id != "" {
		locals.Labels[hcloudlabelkeys.ResourceId] = locals.HetznerCloudPlacementGroup.Metadata.Id
	}

	for k, v := range locals.HetznerCloudPlacementGroup.Metadata.Labels {
		if _, exists := locals.Labels[k]; !exists {
			locals.Labels[k] = v
		}
	}

	return locals
}
