package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway"
	scalewayvpcv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewayvpc/v1"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	ScalewayProviderConfig *scalewayprovider.ScalewayProviderConfig
	ScalewayVpc            *scalewayvpcv1.ScalewayVpc
	ScalewayTags           []string
}

// initializeLocals copies stack-input fields into the Locals struct and builds
// a reusable tag slice. Tags are formatted as "key=value" strings because
// Scaleway tags are flat strings (not key-value maps).
func initializeLocals(_ *pulumi.Context, stackInput *scalewayvpcv1.ScalewayVpcStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayVpc = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Standard labels applied as Scaleway tags.
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, locals.ScalewayVpc.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayVpc.String()),
	}

	if locals.ScalewayVpc.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, locals.ScalewayVpc.Metadata.Org))
	}

	if locals.ScalewayVpc.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, locals.ScalewayVpc.Metadata.Env))
	}

	if locals.ScalewayVpc.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, locals.ScalewayVpc.Metadata.Id))
	}

	return locals
}
