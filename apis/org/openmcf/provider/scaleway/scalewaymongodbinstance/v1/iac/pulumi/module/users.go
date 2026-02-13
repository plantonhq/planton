package module

import (
	"fmt"

	"github.com/pkg/errors"
	scaleway "github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumiverse/pulumi-scaleway/sdk/go/scaleway/mongodb"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// users creates additional MongoDB users with their role assignments.
//
// Each user defined in the spec creates a `mongodb.User` resource with
// inline role assignments. Unlike RDB (which has separate database and
// privilege resources), MongoDB combines roles directly on the user.
//
// Role scope is determined by the spec's role definition:
//   - database_name set: role scoped to that specific database.
//   - any_database set: role applies to all databases.
//
// The admin user is NOT created here -- it is created inline on the
// MongoDB instance resource via the user_name/password fields.
func users(
	ctx *pulumi.Context,
	locals *Locals,
	scalewayProvider *scaleway.Provider,
	instance *mongodb.Instance,
) error {
	spec := locals.ScalewayMongodbInstance.Spec

	for _, user := range spec.Users {
		resourceName := fmt.Sprintf("user-%s", user.Name)

		// Build role assignments.
		roles := make(mongodb.UserRoleArray, 0, len(user.Roles))
		for _, role := range user.Roles {
			roleArgs := &mongodb.UserRoleArgs{
				Role: pulumi.String(role.Role),
			}

			if role.DatabaseName != "" {
				roleArgs.DatabaseName = pulumi.StringPtr(role.DatabaseName)
			}
			if role.AnyDatabase {
				roleArgs.AnyDatabase = pulumi.BoolPtr(true)
			}

			roles = append(roles, roleArgs)
		}

		// Build user arguments.
		userArgs := &mongodb.UserArgs{
			InstanceId: instance.ID(),
			Name:       pulumi.StringPtr(user.Name),
			Password:   pulumi.String(user.Password),
			Region:     pulumi.StringPtr(locals.ScalewayMongodbInstance.Spec.Region),
		}

		if len(roles) > 0 {
			userArgs.Roles = roles
		}

		// Create the user.
		_, err := mongodb.NewUser(
			ctx,
			resourceName,
			userArgs,
			pulumi.Provider(scalewayProvider),
			pulumi.DependsOn([]pulumi.Resource{instance}),
		)
		if err != nil {
			return errors.Wrapf(err, "failed to create mongodb user %q", user.Name)
		}
	}

	return nil
}
