package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	scalewayinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayinstance/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles resolved values from the stack input for use throughout
// the module. All StringValueOrRef fields are resolved to plain strings
// here -- at IaC runtime, valueFrom references have already been resolved
// by the platform middleware.
type Locals struct {
	ScalewayProviderConfig *scalewayprovider.ScalewayProviderConfig
	ScalewayInstance       *scalewayinstancev1.ScalewayInstance

	// SecurityGroupId is resolved from the optional StringValueOrRef field.
	// Empty string if not configured.
	SecurityGroupId string

	// PrivateNetworkId is resolved from the optional StringValueOrRef field.
	// Empty string if not configured.
	PrivateNetworkId string

	ScalewayTags []string
}

// initializeLocals copies stack-input fields into the Locals struct, resolves
// optional StringValueOrRef references, and builds a reusable tag slice.
// Tags are formatted as "key=value" strings because Scaleway tags are flat
// strings (not key-value maps).
func initializeLocals(_ *pulumi.Context, stackInput *scalewayinstancev1.ScalewayInstanceStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayInstance = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Resolve optional Security Group ID from StringValueOrRef.
	if stackInput.Target.Spec.SecurityGroupId != nil {
		locals.SecurityGroupId = stackInput.Target.Spec.SecurityGroupId.GetValue()
	}

	// Resolve optional Private Network ID from StringValueOrRef.
	if stackInput.Target.Spec.PrivateNetworkId != nil {
		locals.PrivateNetworkId = stackInput.Target.Spec.PrivateNetworkId.GetValue()
	}

	// Standard labels applied as Scaleway tags ("key=value" format).
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, locals.ScalewayInstance.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayInstance.String()),
	}

	if locals.ScalewayInstance.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, locals.ScalewayInstance.Metadata.Org))
	}

	if locals.ScalewayInstance.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, locals.ScalewayInstance.Metadata.Env))
	}

	if locals.ScalewayInstance.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, locals.ScalewayInstance.Metadata.Id))
	}

	return locals
}
