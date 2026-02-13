package module

import (
	"github.com/pkg/errors"
	scalewayserverlesscontainerv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewayserverlesscontainer/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point for provisioning a Scaleway serverless
// container. It creates a container namespace, the container itself,
// and optional cron triggers.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewayserverlesscontainerv1.ScalewayServerlessContainerStackInput,
) error {
	// 1. Initialize locals.
	locals := initializeLocals(ctx, stackInput)

	// 2. Create Scaleway provider from credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Provision the container namespace, container, and cron triggers.
	if err := serverlessContainer(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create serverless container")
	}

	return nil
}
