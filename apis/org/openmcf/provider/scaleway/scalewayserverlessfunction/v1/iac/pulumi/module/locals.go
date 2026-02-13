package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway"
	scalewayserverlessfunctionv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewayserverlessfunction/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles resolved values from the stack input for use
// throughout the module.
type Locals struct {
	ScalewayProviderConfig     *scalewayprovider.ScalewayProviderConfig
	ScalewayServerlessFunction *scalewayserverlessfunctionv1.ScalewayServerlessFunction
	ScalewayTags               []string
}

// initializeLocals copies stack-input fields into the Locals struct
// and builds the standard tag slice for Scaleway resources.
func initializeLocals(_ *pulumi.Context, stackInput *scalewayserverlessfunctionv1.ScalewayServerlessFunctionStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayServerlessFunction = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	target := locals.ScalewayServerlessFunction

	// Build standard Planton labels as Scaleway flat string tags
	// formatted as "key=value".
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, target.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayServerlessFunction.String()),
	}

	if target.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, target.Metadata.Org))
	}

	if target.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, target.Metadata.Env))
	}

	if target.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, target.Metadata.Id))
	}

	return locals
}
