package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	scalewayloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayloadbalancer/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles resolved values from the stack input for use throughout the module.
type Locals struct {
	ScalewayProviderConfig *scalewayprovider.ScalewayProviderConfig
	ScalewayLoadBalancer   *scalewayloadbalancerv1.ScalewayLoadBalancer

	// PrivateNetworkId is resolved from the StringValueOrRef field in the spec.
	// Empty string if no Private Network is configured.
	PrivateNetworkId string

	ScalewayTags []string
}

// initializeLocals copies stack-input fields into the Locals struct, resolves
// the StringValueOrRef private_network_id, and builds a reusable tag slice.
func initializeLocals(_ *pulumi.Context, stackInput *scalewayloadbalancerv1.ScalewayLoadBalancerStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayLoadBalancer = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Resolve optional Private Network ID from StringValueOrRef.
	if stackInput.Target.Spec.PrivateNetworkId != nil {
		locals.PrivateNetworkId = stackInput.Target.Spec.PrivateNetworkId.GetValue()
	}

	// Standard labels applied as Scaleway tags ("key=value" format).
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, locals.ScalewayLoadBalancer.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayLoadBalancer.String()),
	}

	if locals.ScalewayLoadBalancer.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, locals.ScalewayLoadBalancer.Metadata.Org))
	}

	if locals.ScalewayLoadBalancer.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, locals.ScalewayLoadBalancer.Metadata.Env))
	}

	if locals.ScalewayLoadBalancer.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, locals.ScalewayLoadBalancer.Metadata.Id))
	}

	return locals
}
