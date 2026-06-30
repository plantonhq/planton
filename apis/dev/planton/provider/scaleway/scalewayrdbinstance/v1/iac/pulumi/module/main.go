package module

import (
	"github.com/pkg/errors"
	scalewayrdbinstancev1 "github.com/plantonhq/planton/apis/dev/planton/provider/scaleway/scalewayrdbinstance/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/scaleway/pulumiscalewayprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the module entry point that provisions the complete
// ScalewayRdbInstance composite:
//
//  1. The RDB instance (managed database engine with admin user).
//  2. Logical databases within the instance.
//  3. Additional users with per-database privileges.
//  4. Network ACL rules restricting access to the public endpoint.
//
// Resources are created in dependency order:
//   - Instance first (all other resources reference the instance ID).
//   - Databases and users in parallel (no dependencies between them).
//   - Privileges after both databases and users exist (references both).
//   - ACL independently (only depends on instance).
func Resources(
	ctx *pulumi.Context,
	stackInput *scalewayrdbinstancev1.ScalewayRdbInstanceStackInput,
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

	// 3. Create the RDB instance and export core outputs.
	createdInstance, err := rdbInstance(ctx, locals, scalewayProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create rdb instance")
	}

	// 4. Create databases, users, privileges, and ACL.
	if err := databasesAndUsers(ctx, locals, scalewayProvider, createdInstance); err != nil {
		return errors.Wrap(err, "failed to create databases and users")
	}

	return nil
}
