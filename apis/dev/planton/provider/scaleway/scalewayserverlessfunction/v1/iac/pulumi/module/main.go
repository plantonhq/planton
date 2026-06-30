package module

import (
	"github.com/pkg/errors"
	scalewayserverlessfunctionv1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayserverlessfunction/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the entry point for provisioning a Scaleway serverless
// function. It creates a function namespace, the function itself, and
// optional cron triggers.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewayserverlessfunctionv1.ScalewayServerlessFunctionStackInput,
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

	// 3. Provision the function namespace, function, and cron triggers.
	if err := serverlessFunction(ctx, locals, scalewayProvider); err != nil {
		return errors.Wrap(err, "failed to create serverless function")
	}

	return nil
}
