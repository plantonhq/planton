package module

import (
	"strconv"

	hetznercloudprovider "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud"
	hetznercloudloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/hetznercloud/hetznercloudloadbalancer/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/hetznercloud/hcloudlabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	HetznerCloudProviderConfig *hetznercloudprovider.HetznerCloudProviderConfig
	HetznerCloudLoadBalancer   *hetznercloudloadbalancerv1.HetznerCloudLoadBalancer
	Labels                     map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *hetznercloudloadbalancerv1.HetznerCloudLoadBalancerStackInput) *Locals {
	locals := &Locals{}

	locals.HetznerCloudLoadBalancer = stackInput.Target
	locals.HetznerCloudProviderConfig = stackInput.ProviderConfig

	locals.Labels = map[string]string{
		hcloudlabelkeys.Resource:     strconv.FormatBool(true),
		hcloudlabelkeys.ResourceName: locals.HetznerCloudLoadBalancer.Metadata.Name,
		hcloudlabelkeys.ResourceKind: cloudresourcekind.CloudResourceKind_HetznerCloudLoadBalancer.String(),
	}

	if locals.HetznerCloudLoadBalancer.Metadata.Org != "" {
		locals.Labels[hcloudlabelkeys.Organization] = locals.HetznerCloudLoadBalancer.Metadata.Org
	}

	if locals.HetznerCloudLoadBalancer.Metadata.Env != "" {
		locals.Labels[hcloudlabelkeys.Environment] = locals.HetznerCloudLoadBalancer.Metadata.Env
	}

	if locals.HetznerCloudLoadBalancer.Metadata.Id != "" {
		locals.Labels[hcloudlabelkeys.ResourceId] = locals.HetznerCloudLoadBalancer.Metadata.Id
	}

	for k, v := range locals.HetznerCloudLoadBalancer.Metadata.Labels {
		if _, exists := locals.Labels[k]; !exists {
			locals.Labels[k] = v
		}
	}

	return locals
}
