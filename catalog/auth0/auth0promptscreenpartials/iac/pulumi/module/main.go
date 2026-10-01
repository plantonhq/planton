package module

import (
	"github.com/pkg/errors"
	auth0promptscreenpartialsv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0promptscreenpartials/v1alpha1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources sets the partials of one prompt, from the stack input, on the tenant
// the provider's credential belongs to.
func Resources(ctx *pulumi.Context, stackInput *auth0promptscreenpartialsv1alpha1.Auth0PromptScreenPartialsStackInput) error {
	locals := initializeLocals(stackInput)

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

	partials, err := applyPromptScreenPartials(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to set the Auth0 prompt's screen partials")
	}

	return exportOutputs(ctx, partials)
}
