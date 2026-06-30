package module

import (
	ocinetworkloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocinetworkloadbalancer/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	OciNetworkLoadBalancer *ocinetworkloadbalancerv1.OciNetworkLoadBalancer
	DisplayName            string
	FreeformTags           map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *ocinetworkloadbalancerv1.OciNetworkLoadBalancerStackInput) *Locals {
	locals := &Locals{}
	locals.OciNetworkLoadBalancer = stackInput.Target

	if stackInput.Target.Spec.DisplayName != "" {
		locals.DisplayName = stackInput.Target.Spec.DisplayName
	} else {
		locals.DisplayName = stackInput.Target.Metadata.Name
	}

	locals.FreeformTags = map[string]string{
		"resource":      "true",
		"resource_kind": cloudresourcekind.CloudResourceKind_OciNetworkLoadBalancer.String(),
		"resource_id":   stackInput.Target.Metadata.Id,
	}
	if stackInput.Target.Metadata.Org != "" {
		locals.FreeformTags["organization"] = stackInput.Target.Metadata.Org
	}
	if stackInput.Target.Metadata.Env != "" {
		locals.FreeformTags["environment"] = stackInput.Target.Metadata.Env
	}
	for k, v := range stackInput.Target.Metadata.Labels {
		locals.FreeformTags[k] = v
	}

	return locals
}
