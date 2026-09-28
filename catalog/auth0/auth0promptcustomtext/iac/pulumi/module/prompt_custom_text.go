package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// applyPromptCustomText sets the words of one prompt in one language -- the twin
// of the Terraform module's auth0_prompt_custom_text. Auth0 stores one custom
// text per prompt and language and replaces it whole on every write, so this
// resource owns all of it: a screen or key the spec leaves out shows Auth0's
// default words. The resource's delete writes an empty document, which returns
// the prompt to Auth0's defaults in that language.
func applyPromptCustomText(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.PromptCustomText, error) {
	customText, err := auth0.NewPromptCustomText(ctx, locals.ResourceName, &auth0.PromptCustomTextArgs{
		Prompt:   pulumi.String(locals.Prompt),
		Language: pulumi.String(locals.Language),
		Body:     pulumi.String(locals.Body),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to set the %s custom text of the %s prompt for %s", locals.Language, locals.Prompt, locals.ResourceName)
	}
	return customText, nil
}
