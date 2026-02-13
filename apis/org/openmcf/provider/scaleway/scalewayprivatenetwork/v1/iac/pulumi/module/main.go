package module

import (
	"github.com/pkg/errors"
	scalewayprivatenetworkv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewayprivatenetwork/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewayprivatenetworkv1.ScalewayPrivateNetworkStackInput,
) error {
	// 1. Prepare locals (metadata, labels, resolved references, etc.).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Scaleway provider from the supplied credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Create the Private Network.
	if _, err := privateNetwork(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create private network")
	}

	return nil
}
