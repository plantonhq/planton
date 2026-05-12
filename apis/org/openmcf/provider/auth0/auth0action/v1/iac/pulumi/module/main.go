package module

import (
	"github.com/pkg/errors"
	auth0actionv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/auth0/auth0action/v1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *auth0actionv1.Auth0ActionStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *auth0.Provider
	var err error
	providerConfig := stackInput.ProviderConfig

	if providerConfig == nil {
		provider, err = auth0.NewProvider(ctx, "auth0-provider", &auth0.ProviderArgs{})
		if err != nil {
			return errors.Wrap(err, "failed to create default Auth0 provider")
		}
	} else {
		provider, err = auth0.NewProvider(ctx, "auth0-provider", &auth0.ProviderArgs{
			Domain:       pulumi.String(providerConfig.Domain),
			ClientId:     pulumi.String(providerConfig.ClientId),
			ClientSecret: pulumi.String(providerConfig.ClientSecret),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create Auth0 provider with credentials")
		}
	}

	createdAction, err := createAction(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Auth0 action")
	}

	if err := createTriggerBinding(ctx, locals, createdAction, provider); err != nil {
		return errors.Wrap(err, "failed to create trigger binding")
	}

	return exportOutputs(ctx, createdAction, locals)
}
