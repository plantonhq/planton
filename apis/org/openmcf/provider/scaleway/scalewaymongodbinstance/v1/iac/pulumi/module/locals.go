package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway"
	scalewaymongodbinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewaymongodbinstance/v1"
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
	ScalewayMongodbInstance *scalewaymongodbinstancev1.ScalewayMongodbInstance

	// PrivateNetworkId is resolved from the optional StringValueOrRef field.
	// Empty string if no Private Network is configured.
	PrivateNetworkId string

	ScalewayTags []string
}

// initializeLocals copies stack-input fields into the Locals struct, resolves
// the StringValueOrRef private_network_id, and builds a reusable tag slice.
// Tags are formatted as "key=value" strings because Scaleway tags are flat
// strings (not key-value maps).
func initializeLocals(_ *pulumi.Context, stackInput *scalewaymongodbinstancev1.ScalewayMongodbInstanceStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayMongodbInstance = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Resolve optional Private Network ID from StringValueOrRef.
	if stackInput.Target.Spec.PrivateNetworkId != nil {
		locals.PrivateNetworkId = stackInput.Target.Spec.PrivateNetworkId.GetValue()
	}

	// Standard labels applied as Scaleway tags ("key=value" format).
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, locals.ScalewayMongodbInstance.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayMongodbInstance.String()),
	}

	if locals.ScalewayMongodbInstance.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, locals.ScalewayMongodbInstance.Metadata.Org))
	}

	if locals.ScalewayMongodbInstance.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, locals.ScalewayMongodbInstance.Metadata.Env))
	}

	if locals.ScalewayMongodbInstance.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, locals.ScalewayMongodbInstance.Metadata.Id))
	}

	return locals
}
