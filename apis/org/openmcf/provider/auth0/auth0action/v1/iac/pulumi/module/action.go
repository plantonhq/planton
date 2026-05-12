package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createAction(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.Action, error) {
	actionArgs := &auth0.ActionArgs{
		Name:   pulumi.String(locals.ActionName),
		Code:   pulumi.String(locals.Code),
		Deploy: pulumi.Bool(locals.Deploy),
		SupportedTriggers: &auth0.ActionSupportedTriggersArgs{
			Id:      pulumi.String(locals.SupportedTrigger.Id),
			Version: pulumi.String(locals.SupportedTrigger.Version),
		},
	}

	if locals.Runtime != "" {
		actionArgs.Runtime = pulumi.String(locals.Runtime)
	}

	if len(locals.Dependencies) > 0 {
		deps := auth0.ActionDependencyArray{}
		for _, dep := range locals.Dependencies {
			deps = append(deps, &auth0.ActionDependencyArgs{
				Name:    pulumi.String(dep.Name),
				Version: pulumi.String(dep.Version),
			})
		}
		actionArgs.Dependencies = deps
	}

	if len(locals.Secrets) > 0 {
		secs := auth0.ActionSecretArray{}
		for _, sec := range locals.Secrets {
			secs = append(secs, &auth0.ActionSecretArgs{
				Name:  pulumi.String(sec.Name),
				Value: pulumi.String(sec.Value),
			})
		}
		actionArgs.Secrets = secs
	}

	action, err := auth0.NewAction(ctx, locals.ActionName, actionArgs, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Auth0 action %s", locals.ActionName)
	}

	return action, nil
}
