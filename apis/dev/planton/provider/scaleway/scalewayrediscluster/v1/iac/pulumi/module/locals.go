package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	scalewayredisclusterv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayrediscluster/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles resolved values from the stack input for use throughout
// the module. The StringValueOrRef for private_network_id is resolved to
// a plain string here -- at IaC runtime, valueFrom references have already
// been resolved by the platform middleware.
type Locals struct {
	ScalewayProviderConfig *scalewayprovider.ScalewayProviderConfig
	ScalewayRedisCluster   *scalewayredisclusterv1.ScalewayRedisCluster

	// PrivateNetworkId is resolved from the optional StringValueOrRef field.
	// Empty string if no Private Network is configured.
	PrivateNetworkId string

	ScalewayTags []string
}

// initializeLocals copies stack-input fields into the Locals struct, resolves
// the StringValueOrRef private_network_id, and builds a reusable tag slice.
// Tags are formatted as "key=value" strings because Scaleway tags are flat
// strings (not key-value maps).
func initializeLocals(_ *pulumi.Context, stackInput *scalewayredisclusterv1.ScalewayRedisClusterStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayRedisCluster = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Resolve optional Private Network ID from StringValueOrRef.
	if stackInput.Target.Spec.PrivateNetworkId != nil {
		locals.PrivateNetworkId = stackInput.Target.Spec.PrivateNetworkId.GetValue()
	}

	// Standard labels applied as Scaleway tags ("key=value" format).
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, locals.ScalewayRedisCluster.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayRedisCluster.String()),
	}

	if locals.ScalewayRedisCluster.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, locals.ScalewayRedisCluster.Metadata.Org))
	}

	if locals.ScalewayRedisCluster.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, locals.ScalewayRedisCluster.Metadata.Env))
	}

	if locals.ScalewayRedisCluster.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, locals.ScalewayRedisCluster.Metadata.Id))
	}

	return locals
}
