package module

import (
	"strconv"

	hetznercloudprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud"
	hetznercloudsnapshotv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud/hetznercloudsnapshot/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/hcloudlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	HetznerCloudProviderConfig *hetznercloudprovider.HetznerCloudProviderConfig
	HetznerCloudSnapshot       *hetznercloudsnapshotv1.HetznerCloudSnapshot
	Labels                     map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *hetznercloudsnapshotv1.HetznerCloudSnapshotStackInput) *Locals {
	locals := &Locals{}

	locals.HetznerCloudSnapshot = stackInput.Target
	locals.HetznerCloudProviderConfig = stackInput.ProviderConfig

	locals.Labels = map[string]string{
		hcloudlabelkeys.Resource:     strconv.FormatBool(true),
		hcloudlabelkeys.ResourceName: locals.HetznerCloudSnapshot.Metadata.Name,
		hcloudlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_HetznerCloudSnapshot.String(),
	}

	if locals.HetznerCloudSnapshot.Metadata.Org != "" {
		locals.Labels[hcloudlabelkeys.Organization] = locals.HetznerCloudSnapshot.Metadata.Org
	}

	if locals.HetznerCloudSnapshot.Metadata.Env != "" {
		locals.Labels[hcloudlabelkeys.Environment] = locals.HetznerCloudSnapshot.Metadata.Env
	}

	if locals.HetznerCloudSnapshot.Metadata.Id != "" {
		locals.Labels[hcloudlabelkeys.ResourceId] = locals.HetznerCloudSnapshot.Metadata.Id
	}

	for k, v := range locals.HetznerCloudSnapshot.Metadata.Labels {
		if _, exists := locals.Labels[k]; !exists {
			locals.Labels[k] = v
		}
	}

	return locals
}
