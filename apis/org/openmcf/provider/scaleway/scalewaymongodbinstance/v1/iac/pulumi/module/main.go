package module

import (
	"github.com/pkg/errors"
	scalewaymongodbinstancev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/scaleway/scalewaymongodbinstance/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that provisions the complete
// ScalewayMongodbInstance composite:
//
//  1. The MongoDB instance (managed database engine with admin user).
//  2. Additional users with role-based access control.
//
// Resources are created in dependency order:
//   - Instance first (users reference the instance ID).
//   - Users after the instance exists.
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewaymongodbinstancev1.ScalewayMongodbInstanceStackInput,
) error {
	// 1. Prepare locals (metadata, labels, resolved references).
	locals := initializeLocals(ctx, stackInput)

	// 2. Create a Scaleway provider from the supplied credential.
	scalewayProvider, err := pulumiscalewayprovider.Get(
		ctx,
		stackInput.ProviderConfig,
	)
	if err != nil {
		return errors.Wrap(err, "failed to setup scaleway provider")
	}

	// 3. Create the MongoDB instance and export core outputs.
	createdInstance, err := mongodbInstance(ctx, locals, scalewayProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create mongodb instance")
	}

	// 4. Create additional users with roles.
	if err := users(ctx, locals, scalewayProvider, createdInstance); err != nil {
		return errors.Wrap(err, "failed to create mongodb users")
	}

	return nil
}
