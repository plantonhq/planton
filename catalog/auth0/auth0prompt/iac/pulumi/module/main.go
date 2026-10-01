package module

import (
	"github.com/pkg/errors"
	auth0promptv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0prompt/v1alpha1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources applies the login-flow settings from the stack input to the tenant
// the provider's credential belongs to.
func Resources(ctx *pulumi.Context, stackInput *auth0promptv1alpha1.Auth0PromptStackInput) error {
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

	prompt, err := applyPrompt(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to apply the Auth0 prompt settings")
	}

	return exportOutputs(ctx, prompt)
}
