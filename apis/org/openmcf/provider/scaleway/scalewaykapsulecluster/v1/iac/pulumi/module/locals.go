package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway"
	scalewaykapsuleclusterv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewaykapsulecluster/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles resolved values from the stack input for use throughout
// the module. The StringValueOrRef for private_network_id is resolved to
// a plain string here -- at IaC runtime, valueFrom references have already
// been resolved by the platform middleware.
type Locals struct {
	ScalewayProviderConfig  *scalewayprovider.ScalewayProviderConfig
	ScalewayKapsuleCluster  *scalewaykapsuleclusterv1.ScalewayKapsuleCluster

	// PrivateNetworkId is resolved from the required StringValueOrRef field.
	PrivateNetworkId string

	ScalewayTags []string
}

// initializeLocals copies stack-input fields into the Locals struct, resolves
// the StringValueOrRef private_network_id, and builds a reusable tag slice.
// Tags are formatted as "key=value" strings because Scaleway tags are flat
// strings (not key-value maps).
func initializeLocals(_ *pulumi.Context, stackInput *scalewaykapsuleclusterv1.ScalewayKapsuleClusterStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayKapsuleCluster = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Resolve required Private Network ID from StringValueOrRef.
	if stackInput.Target.Spec.PrivateNetworkId != nil {
		locals.PrivateNetworkId = stackInput.Target.Spec.PrivateNetworkId.GetValue()
	}

	// Standard labels applied as Scaleway tags ("key=value" format).
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, locals.ScalewayKapsuleCluster.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayKapsuleCluster.String()),
	}

	if locals.ScalewayKapsuleCluster.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, locals.ScalewayKapsuleCluster.Metadata.Org))
	}

	if locals.ScalewayKapsuleCluster.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, locals.ScalewayKapsuleCluster.Metadata.Env))
	}

	if locals.ScalewayKapsuleCluster.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, locals.ScalewayKapsuleCluster.Metadata.Id))
	}

	return locals
}
