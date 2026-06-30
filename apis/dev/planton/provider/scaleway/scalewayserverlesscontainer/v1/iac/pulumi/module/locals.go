package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	scalewayserverlesscontainerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayserverlesscontainer/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles resolved values from the stack input for use
// throughout the module.
type Locals struct {
	ScalewayProviderConfig      *scalewayprovider.ScalewayProviderConfig
	ScalewayServerlessContainer *scalewayserverlesscontainerv1.ScalewayServerlessContainer
	ScalewayTags                []string
}

// initializeLocals copies stack-input fields into the Locals struct
// and builds the standard tag slice for Scaleway resources.
func initializeLocals(_ *pulumi.Context, stackInput *scalewayserverlesscontainerv1.ScalewayServerlessContainerStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayServerlessContainer = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	target := locals.ScalewayServerlessContainer

	// Build standard Planton labels as Scaleway flat string tags
	// formatted as "key=value".
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, target.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayServerlessContainer.String()),
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
