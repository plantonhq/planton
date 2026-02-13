package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway"
	scalewayprivatenetworkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewayprivatenetwork/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	ScalewayProviderConfig *scalewayprovider.ScalewayProviderConfig
	ScalewayPrivateNetwork *scalewayprivatenetworkv1.ScalewayPrivateNetwork
	// VpcId is resolved from the StringValueOrRef field in the spec.
	// The platform middleware resolves valueFrom references before IaC
	// modules run, so GetValue() always returns the resolved literal string.
	VpcId        string
	ScalewayTags []string
}

// initializeLocals copies stack-input fields into the Locals struct, resolves
// the StringValueOrRef vpc_id, and builds a reusable tag slice.
// Tags are formatted as "key=value" strings because Scaleway tags are flat
// strings (not key-value maps).
func initializeLocals(_ *pulumi.Context, stackInput *scalewayprivatenetworkv1.ScalewayPrivateNetworkStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayPrivateNetwork = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Resolve the VPC ID from StringValueOrRef.
	// At IaC runtime, value_from references have already been resolved by the
	// platform, so GetValue() returns the final literal VPC UUID.
	locals.VpcId = stackInput.Target.Spec.VpcId.GetValue()

	// Standard labels applied as Scaleway tags.
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, locals.ScalewayPrivateNetwork.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayPrivateNetwork.String()),
	}

	if locals.ScalewayPrivateNetwork.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, locals.ScalewayPrivateNetwork.Metadata.Org))
	}

	if locals.ScalewayPrivateNetwork.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, locals.ScalewayPrivateNetwork.Metadata.Env))
	}

	if locals.ScalewayPrivateNetwork.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, locals.ScalewayPrivateNetwork.Metadata.Id))
	}

	return locals
}
