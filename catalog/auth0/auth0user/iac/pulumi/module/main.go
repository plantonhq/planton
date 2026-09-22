package module

import (
	"github.com/pkg/errors"
	auth0userv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0user/v1alpha1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates an Auth0 User, sets its authoritative roles and direct
// permissions, and exports the outputs, from the stack input.
func Resources(ctx *pulumi.Context, stackInput *auth0userv1alpha1.Auth0UserStackInput) error {
	locals, err := initializeLocals(ctx, stackInput)
	if err != nil {
		return errors.Wrap(err, "failed to initialize locals")
	}

	// Setup Auth0 provider with credentials from provider config.
	var provider *auth0.Provider
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

	// Create the user, minting its password when the spec declares none.
	user, mintedPassword, minted, err := createUser(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Auth0 user")
	}

	// Set the user's roles (no-op when none are declared).
	if err := createUserRoles(ctx, locals, provider, user); err != nil {
		return errors.Wrap(err, "failed to set Auth0 user roles")
	}

	// Set the user's direct permissions (no-op when none are declared).
	if err := createUserPermissions(ctx, locals, provider, user); err != nil {
		return errors.Wrap(err, "failed to set Auth0 user permissions")
	}

	// Export stack outputs.
	return exportOutputs(ctx, user, mintedPassword, minted)
}
