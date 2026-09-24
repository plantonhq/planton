package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	// generatedPasswordLength: letters and digits only, no symbols, so the
	// minted value satisfies Auth0's every built-in password policy (the
	// strictest, "excellent", wants 10+ characters from three classes) and
	// never needs quoting wherever a person pastes it. Twin of the Terraform
	// module's random_password arguments.
	generatedPasswordLength = 24
)

// createUser creates the Auth0 user, minting its initial password when the
// spec declares none on a database connection. Returns the user and, when a
// password was generated, the secret output carrying it (nil otherwise) so
// exportOutputs can report the minted value and only the minted value.
func createUser(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.User, pulumi.StringOutput, bool, error) {
	spec := locals.Auth0User.Spec

	userArgs := &auth0.UserArgs{
		ConnectionName: pulumi.String(locals.ConnectionName),
	}

	// Optional strings are set only when declared: an empty argument would
	// still reach the provider as a change from "unset" and, for fields Auth0
	// derives (name, nickname, picture), overwrite the derived value.
	if spec.Email != "" {
		userArgs.Email = pulumi.String(spec.Email)
	}
	if spec.Username != "" {
		userArgs.Username = pulumi.String(spec.Username)
	}
	if spec.Name != "" {
		userArgs.Name = pulumi.String(spec.Name)
	}
	if spec.GivenName != "" {
		userArgs.GivenName = pulumi.String(spec.GivenName)
	}
	if spec.FamilyName != "" {
		userArgs.FamilyName = pulumi.String(spec.FamilyName)
	}
	if spec.Nickname != "" {
		userArgs.Nickname = pulumi.String(spec.Nickname)
	}
	if spec.Picture != "" {
		userArgs.Picture = pulumi.String(spec.Picture)
	}
	if spec.PhoneNumber != "" {
		userArgs.PhoneNumber = pulumi.String(spec.PhoneNumber)
	}
	if spec.UserId != "" {
		userArgs.UserId = pulumi.String(spec.UserId)
	}
	if spec.CustomDomainHeader != "" {
		userArgs.CustomDomainHeader = pulumi.String(spec.CustomDomainHeader)
	}
	if locals.UserMetadataJSON != "" {
		userArgs.UserMetadata = pulumi.String(locals.UserMetadataJSON)
	}
	if locals.AppMetadataJSON != "" {
		userArgs.AppMetadata = pulumi.String(locals.AppMetadataJSON)
	}

	// Plain bools whose false equals the provider's own omitted behavior are
	// sent as declared; verify_email is presence-tracked because its unset
	// state means "Auth0 decides", a third state a plain bool cannot carry.
	userArgs.EmailVerified = pulumi.Bool(spec.EmailVerified)
	userArgs.PhoneVerified = pulumi.Bool(spec.PhoneVerified)
	userArgs.Blocked = pulumi.Bool(spec.Blocked)
	if spec.VerifyEmail != nil {
		userArgs.VerifyEmail = pulumi.Bool(*spec.VerifyEmail)
	}

	// The password: declared (used as given, never echoed), minted (a
	// database user with none declared), or none at all (passwordless).
	var mintedPassword pulumi.StringOutput
	minted := false
	switch {
	case spec.Passwordless:
		// Auth0 refuses a password on an email or SMS connection.
	case locals.GeneratePassword:
		// The generation-shape arguments are ignored after creation so an
		// IMPORTED credential never silently regenerates: rotation stays an
		// explicit act in Auth0, never plan fallout. Twin: the Terraform
		// module's lifecycle.ignore_changes on the same argument set.
		generationShapeIgnores := pulumi.IgnoreChanges([]string{
			"length", "special", "upper", "lower", "numeric",
			"minLower", "minNumeric", "minSpecial", "minUpper", "overrideSpecial",
		})
		generated, err := random.NewRandomPassword(ctx, locals.ResourceName+"-password",
			&random.RandomPasswordArgs{
				Length:     pulumi.Int(generatedPasswordLength),
				Special:    pulumi.Bool(false),
				MinUpper:   pulumi.Int(2),
				MinLower:   pulumi.Int(2),
				MinNumeric: pulumi.Int(2),
			},
			generationShapeIgnores)
		if err != nil {
			return nil, mintedPassword, false, errors.Wrap(err, "failed to generate the initial password")
		}
		mintedPassword = generated.Result
		minted = true
		userArgs.Password = mintedPassword
	default:
		// Marked secret so the value is encrypted in the Pulumi state -- twin
		// of the Terraform module's sensitive variable.
		userArgs.Password = pulumi.ToSecret(pulumi.String(spec.Password)).(pulumi.StringOutput)
	}

	user, err := auth0.NewUser(ctx, locals.ResourceName, userArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, mintedPassword, false, errors.Wrapf(err, "failed to create Auth0 user %s", locals.ResourceName)
	}

	return user, mintedPassword, minted, nil
}

// createUserRoles sets the authoritative role set for the user.
// auth0.UserRoles manages the complete list of roles assigned to the user, so a
// role omitted here is removed from the user on the next apply. No-op when the
// spec declares no roles.
func createUserRoles(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider, user *auth0.User) error {
	if len(locals.RoleIds) == 0 {
		return nil
	}

	_, err := auth0.NewUserRoles(ctx, locals.ResourceName+"-roles", &auth0.UserRolesArgs{
		UserId: user.ID(),
		Roles:  pulumi.ToStringArray(locals.RoleIds),
	}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{user}))
	if err != nil {
		return errors.Wrapf(err, "failed to set roles for Auth0 user %s", locals.ResourceName)
	}

	return nil
}

// createUserPermissions sets the authoritative set of direct API permissions
// for the user. auth0.UserPermissions manages the complete list, so a
// permission omitted here is removed from the user on the next apply. No-op
// when the spec declares no permissions.
func createUserPermissions(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider, user *auth0.User) error {
	if len(locals.Permissions) == 0 {
		return nil
	}

	permissionArray := auth0.UserPermissionsPermissionArray{}
	for _, permission := range locals.Permissions {
		permissionArray = append(permissionArray, &auth0.UserPermissionsPermissionArgs{
			Name:                     pulumi.String(permission.Name),
			ResourceServerIdentifier: pulumi.String(permission.ResourceServerIdentifier),
		})
	}

	_, err := auth0.NewUserPermissions(ctx, locals.ResourceName+"-permissions", &auth0.UserPermissionsArgs{
		UserId:      user.ID(),
		Permissions: permissionArray,
	}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{user}))
	if err != nil {
		return errors.Wrapf(err, "failed to set permissions for Auth0 user %s", locals.ResourceName)
	}

	return nil
}
