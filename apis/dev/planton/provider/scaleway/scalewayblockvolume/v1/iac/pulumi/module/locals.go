package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	scalewayblockvolumev1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayblockvolume/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles resolved values from the stack input for use throughout
// the module.
type Locals struct {
	ScalewayProviderConfig *scalewayprovider.ScalewayProviderConfig
	ScalewayBlockVolume    *scalewayblockvolumev1.ScalewayBlockVolume
	ScalewayTags           []string
}

// initializeLocals copies stack-input fields into the Locals struct and
// builds a reusable tag slice. Tags are formatted as "key=value" strings
// because Scaleway Block Storage tags are flat strings (not key-value maps).
func initializeLocals(_ *pulumi.Context, stackInput *scalewayblockvolumev1.ScalewayBlockVolumeStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayBlockVolume = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Standard labels applied as Scaleway tags.
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, locals.ScalewayBlockVolume.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayBlockVolume.String()),
	}

	if locals.ScalewayBlockVolume.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, locals.ScalewayBlockVolume.Metadata.Org))
	}

	if locals.ScalewayBlockVolume.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, locals.ScalewayBlockVolume.Metadata.Env))
	}

	if locals.ScalewayBlockVolume.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, locals.ScalewayBlockVolume.Metadata.Id))
	}

	return locals
}
