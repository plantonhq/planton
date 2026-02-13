package module

import (
	"github.com/pkg/errors"
	scalewayinstancesecuritygroupv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewayinstancesecuritygroup/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewayinstancesecuritygroupv1.ScalewayInstanceSecurityGroupStackInput,
) error {
	// 1. Prepare locals (metadata, labels, credentials, etc.).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Scaleway provider from the supplied credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Create the security group.
	if err := securityGroup(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create security group")
	}

	return nil
}
