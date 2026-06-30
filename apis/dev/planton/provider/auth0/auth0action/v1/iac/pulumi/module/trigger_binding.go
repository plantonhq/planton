package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func createTriggerBinding(ctx *pulumi.Context, locals *Locals, action *auth0.Action, provider *auth0.Provider) error {
	if locals.TriggerBinding == nil {
		return nil
	}

	resourceName := fmt.Sprintf("%s-trigger-binding", locals.ActionName)

	_, err := auth0.NewTriggerAction(ctx, resourceName, &auth0.TriggerActionArgs{
		Trigger:  pulumi.String(locals.SupportedTrigger.Id),
		ActionId: action.ID(),
	}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{action}))
	if err != nil {
		return errors.Wrapf(err, "failed to bind Auth0 action %s to trigger %s", locals.ActionName, locals.SupportedTrigger.Id)
	}

	return nil
}
