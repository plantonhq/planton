package module

import (
	"github.com/pkg/errors"
	auth0resourceserverv1 "github.com/plantonhq/planton/apis/dev/planton/provider/auth0/auth0resourceserver/v1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates an Auth0 Resource Server (API) with all configured parameters
func Resources(ctx *pulumi.Context, stackInput *auth0resourceserverv1.Auth0ResourceServerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Setup Auth0 provider with credentials from provider config
	var provider *auth0.Provider
	var err error
	providerConfig := stackInput.ProviderConfig

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
	if len(locals.Scopes) > 0 {
		_, err = createResourceServerScopes(ctx, locals, provider, resourceServer)
		if err != nil {
			return errors.Wrap(err, "failed to create Auth0 resource server scopes")
		}
	}

	// Export stack outputs
	return exportOutputs(ctx, resourceServer, locals)
}
