package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// createRole creates the Auth0 role.
func createRole(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.Role, error) {
	roleArgs := &auth0.RoleArgs{
		Name: pulumi.String(locals.RoleName),
	}

	if locals.Description != "" {
		roleArgs.Description = pulumi.String(locals.Description)
	}

	role, err := auth0.NewRole(ctx, locals.ResourceName, roleArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Auth0 role %s", locals.ResourceName)
	}

	return role, nil
}

// createRolePermissions sets the authoritative permission set for the role.
// auth0.RolePermissions manages the complete list of permissions assigned to the
// role, so a permission omitted here is removed from the role on the next apply.
func createRolePermissions(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider, role *auth0.Role) error {
	if len(locals.Permissions) == 0 {
		return nil
	}

	permissionArray := auth0.RolePermissionsPermissionArray{}
	for _, permission := range locals.Permissions {
		permissionArray = append(permissionArray, &auth0.RolePermissionsPermissionArgs{
			Name:                     pulumi.String(permission.Name),
			ResourceServerIdentifier: pulumi.String(permission.ResourceServerIdentifier),
		})
	}

	_, err := auth0.NewRolePermissions(ctx, locals.ResourceName+"-permissions", &auth0.RolePermissionsArgs{
		RoleId:      role.ID(),
		Permissions: permissionArray,
	}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{role}))
	if err != nil {
		return errors.Wrapf(err, "failed to set permissions for Auth0 role %s", locals.ResourceName)
	}

	return nil
}
