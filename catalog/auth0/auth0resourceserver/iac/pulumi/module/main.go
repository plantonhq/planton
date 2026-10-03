package module

import (
	"github.com/pkg/errors"
	auth0resourceserverv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0resourceserver/v1alpha1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates an Auth0 Resource Server (API) with all configured parameters
func Resources(ctx *pulumi.Context, iacInput *auth0resourceserverv1alpha1.Auth0ResourceServerIacInput) error {
	locals := initializeLocals(iacInput)

	// Setup Auth0 provider with credentials from provider config
	var provider *auth0.Provider
	var err error
	providerConfig := iacInput.ProviderConfig

	if providerConfig == nil {
		// Use default provider (assumes credentials from environment variables)
		// Environment variables: AUTH0_DOMAIN, AUTH0_CLIENT_ID, AUTH0_CLIENT_SECRET
		provider, err = auth0.NewProvider(ctx, "auth0-provider", &auth0.ProviderArgs{})
		if err != nil {
			return errors.Wrap(err, "failed to create default Auth0 provider")
		}
	} else {
		// Create provider with explicit credentials
		provider, err = auth0.NewProvider(ctx, "auth0-provider", &auth0.ProviderArgs{
			Domain:       pulumi.String(providerConfig.Domain),
			ClientId:     pulumi.String(providerConfig.ClientId),
			ClientSecret: pulumi.String(providerConfig.ClientSecret),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create Auth0 provider with credentials")
		}
	}

	// Create the resource server
	resourceServer, err := createResourceServer(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Auth0 resource server")
	}

	// Create scopes if defined
	var scopes *auth0.ResourceServerScopes
	if len(locals.Scopes) > 0 {
		scopes, err = createResourceServerScopes(ctx, locals, provider, resourceServer)
		if err != nil {
			return errors.Wrap(err, "failed to create Auth0 resource server scopes")
		}
	}

	// Default grants for third-party applications, one per subject type
	defaultGrantIds, err := createDefaultGrants(ctx, locals, provider, resourceServer, scopes)
	if err != nil {
		return errors.Wrap(err, "failed to create the default grants for third-party applications")
	}

	// Export outputs
	return exportOutputs(ctx, resourceServer, defaultGrantIds)
}
