package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway"
	scalewaypublicgatewayv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewaypublicgateway/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	ScalewayProviderConfig *scalewayprovider.ScalewayProviderConfig
	ScalewayPublicGateway  *scalewaypublicgatewayv1.ScalewayPublicGateway

	// PrivateNetworkId is resolved from the StringValueOrRef field in the spec.
	// The platform middleware resolves valueFrom references before IaC
	// modules run, so GetValue() always returns the resolved literal string.
	PrivateNetworkId string

	ScalewayTags []string
}

// initializeLocals copies stack-input fields into the Locals struct, resolves
// the StringValueOrRef private_network_id, and builds a reusable tag slice.
// Tags are formatted as "key=value" strings because Scaleway tags are flat
// strings (not key-value maps).
func initializeLocals(_ *pulumi.Context, stackInput *scalewaypublicgatewayv1.ScalewayPublicGatewayStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayPublicGateway = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Resolve the Private Network ID from StringValueOrRef.
	// At IaC runtime, value_from references have already been resolved by the
	// platform, so GetValue() returns the final literal UUID.
	locals.PrivateNetworkId = stackInput.Target.Spec.PrivateNetworkId.GetValue()

	// Standard labels applied as Scaleway tags.
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, locals.ScalewayPublicGateway.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayPublicGateway.String()),
	}

	if locals.ScalewayPublicGateway.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, locals.ScalewayPublicGateway.Metadata.Org))
	}

	if locals.ScalewayPublicGateway.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, locals.ScalewayPublicGateway.Metadata.Env))
	}

	if locals.ScalewayPublicGateway.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, locals.ScalewayPublicGateway.Metadata.Id))
	}

	return locals
}
