package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// applyPrompt manages the login-flow settings of the tenant the provider's
// credential belongs to -- the twin of the Terraform module's auth0_prompt. A
// tenant has exactly one set. Only the settings the spec declares are sent, so
// the tenant keeps the others as they are. The resource's delete is the
// provider's no-op: destroy leaves the last-applied values in place, as Auth0
// has no delete for the prompt settings.
func applyPrompt(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.Prompt, error) {
	prompt, err := auth0.NewPrompt(ctx, locals.ResourceName, &auth0.PromptArgs{
		UniversalLoginExperience:    pulumi.StringPtrFromPtr(locals.UniversalLoginExperience),
		IdentifierFirst:             pulumi.BoolPtrFromPtr(locals.IdentifierFirst),
		WebauthnPlatformFirstFactor: pulumi.BoolPtrFromPtr(locals.WebauthnPlatformFirstFactor),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to apply the Auth0 prompt settings for %s", locals.ResourceName)
	}
	return prompt, nil
}
