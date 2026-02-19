package module

import (
	"strconv"

	hetznercloudprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud"
	hetznercloudserverv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/hetznercloud/hetznercloudserver/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/hcloudlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	HetznerCloudProviderConfig *hetznercloudprovider.HetznerCloudProviderConfig
	HetznerCloudServer         *hetznercloudserverv1.HetznerCloudServer
	Labels                     map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *hetznercloudserverv1.HetznerCloudServerStackInput) *Locals {
	locals := &Locals{}

	locals.HetznerCloudServer = stackInput.Target
	locals.HetznerCloudProviderConfig = stackInput.ProviderConfig

	locals.Labels = map[string]string{
		hcloudlabelkeys.Resource:     strconv.FormatBool(true),
		hcloudlabelkeys.ResourceName: locals.HetznerCloudServer.Metadata.Name,
		hcloudlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_HetznerCloudServer.String(),
	}

	if locals.HetznerCloudServer.Metadata.Org != "" {
		locals.Labels[hcloudlabelkeys.Organization] = locals.HetznerCloudServer.Metadata.Org
	}

	if locals.HetznerCloudServer.Metadata.Env != "" {
		locals.Labels[hcloudlabelkeys.Environment] = locals.HetznerCloudServer.Metadata.Env
	}

	if locals.HetznerCloudServer.Metadata.Id != "" {
		locals.Labels[hcloudlabelkeys.ResourceId] = locals.HetznerCloudServer.Metadata.Id
	}

	for k, v := range locals.HetznerCloudServer.Metadata.Labels {
		if _, exists := locals.Labels[k]; !exists {
			locals.Labels[k] = v
		}
	}

	return locals
}
