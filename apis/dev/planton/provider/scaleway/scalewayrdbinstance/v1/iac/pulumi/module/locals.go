package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	scalewayrdbinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayrdbinstance/v1"
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
	ScalewayRdbInstance    *scalewayrdbinstancev1.ScalewayRdbInstance

	// PrivateNetworkId is resolved from the optional StringValueOrRef field.
	// Empty string if no Private Network is configured.
	PrivateNetworkId string

	ScalewayTags []string
}

// initializeLocals copies stack-input fields into the Locals struct, resolves
// the StringValueOrRef private_network_id, and builds a reusable tag slice.
// Tags are formatted as "key=value" strings because Scaleway tags are flat
// strings (not key-value maps).
func initializeLocals(_ *pulumi.Context, stackInput *scalewayrdbinstancev1.ScalewayRdbInstanceStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayRdbInstance = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Resolve optional Private Network ID from StringValueOrRef.
	if stackInput.Target.Spec.PrivateNetworkId != nil {
		locals.PrivateNetworkId = stackInput.Target.Spec.PrivateNetworkId.GetValue()
	}

	// Standard labels applied as Scaleway tags ("key=value" format).
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, locals.ScalewayRdbInstance.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayRdbInstance.String()),
	}

	if locals.ScalewayRdbInstance.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, locals.ScalewayRdbInstance.Metadata.Org))
	}

	if locals.ScalewayRdbInstance.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, locals.ScalewayRdbInstance.Metadata.Env))
	}

	if locals.ScalewayRdbInstance.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, locals.ScalewayRdbInstance.Metadata.Id))
	}

	return locals
}
