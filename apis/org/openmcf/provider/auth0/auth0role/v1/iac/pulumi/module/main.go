package module

import (
	"github.com/pkg/errors"
	auth0rolev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/auth0/auth0role/v1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates an Auth0 Role and sets its permissions from the stack input.
func Resources(ctx *pulumi.Context, stackInput *auth0rolev1.Auth0RoleStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Setup Auth0 provider with credentials from provider config.
	var provider *auth0.Provider
	var err error
	providerConfig := stackInput.ProviderConfig

	if providerConfig == nil {
		// Use default provider (assumes credentials from environment variables).
		// Environment variables: AUTH0_DOMAIN, AUTH0_CLIENT_ID, AUTH0_CLIENT_SECRET
		provider, err = auth0.NewProvider(ctx, "auth0-provider", &auth0.ProviderArgs{})
		if err != nil {
			return errors.Wrap(err, "failed to create default Auth0 provider")
		}
	} else {
		// Create provider with explicit credentials.
		provider, err = auth0.NewProvider(ctx, "auth0-provider", &auth0.ProviderArgs{
			Domain:       pulumi.String(providerConfig.Domain),
			ClientId:     pulumi.String(providerConfig.ClientId),
			ClientSecret: pulumi.String(providerConfig.ClientSecret),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create Auth0 provider with credentials")
		}
	}

	// Create the role.
	role, err := createRole(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Auth0 role")
	}

	// Set the role's permissions (no-op when none are defined).
	if err := createRolePermissions(ctx, locals, provider, role); err != nil {
		return errors.Wrap(err, "failed to set Auth0 role permissions")
	}

	// Export stack outputs.
	return exportOutputs(ctx, role, locals)
}
