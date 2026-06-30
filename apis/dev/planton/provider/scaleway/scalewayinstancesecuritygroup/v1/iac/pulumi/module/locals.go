package module

import (
	"fmt"
	"strconv"

	scalewayprovider "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway"
	scalewayinstancesecuritygroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayinstancesecuritygroup/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/scalewaylabelkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals bundles handy references for the rest of the module.
type Locals struct {
	ScalewayProviderConfig        *scalewayprovider.ScalewayProviderConfig
	ScalewayInstanceSecurityGroup *scalewayinstancesecuritygroupv1.ScalewayInstanceSecurityGroup
	ScalewayTags                  []string
}

// initializeLocals copies stack-input fields into the Locals struct and builds
// a reusable tag slice. Tags are formatted as "key=value" strings because
// Scaleway tags are flat strings (not key-value maps).
func initializeLocals(_ *pulumi.Context, stackInput *scalewayinstancesecuritygroupv1.ScalewayInstanceSecurityGroupStackInput) *Locals {
	locals := &Locals{}

	locals.ScalewayInstanceSecurityGroup = stackInput.Target
	locals.ScalewayProviderConfig = stackInput.ProviderConfig

	// Standard labels applied as Scaleway tags.
	locals.ScalewayTags = []string{
		fmt.Sprintf("%s=%s", scalewaylabelkeys.Resource, strconv.FormatBool(true)),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceName, locals.ScalewayInstanceSecurityGroup.Metadata.Name),
		fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceKind, cloudresourcekind.CloudResourceKind_ScalewayInstanceSecurityGroup.String()),
	}

	if locals.ScalewayInstanceSecurityGroup.Metadata.Org != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Organization, locals.ScalewayInstanceSecurityGroup.Metadata.Org))
	}

	if locals.ScalewayInstanceSecurityGroup.Metadata.Env != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.Environment, locals.ScalewayInstanceSecurityGroup.Metadata.Env))
	}

	if locals.ScalewayInstanceSecurityGroup.Metadata.Id != "" {
		locals.ScalewayTags = append(locals.ScalewayTags,
			fmt.Sprintf("%s=%s", scalewaylabelkeys.ResourceId, locals.ScalewayInstanceSecurityGroup.Metadata.Id))
	}

	return locals
}
