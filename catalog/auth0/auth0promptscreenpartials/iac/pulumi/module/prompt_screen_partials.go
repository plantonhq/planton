package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// applyPromptScreenPartials sets every partial of one prompt -- the twin of the
// Terraform module's auth0_prompt_screen_partials. Auth0 replaces the prompt's
// whole set on every write, so a screen or insertion point the spec leaves out
// renders nothing. The resource's delete writes an empty set, removing every
// partial of the prompt. The single-screen auth0.PromptScreenPartial is never
// declared: this resource owns the whole set, and Auth0 warns against mixing
// the two on one prompt.
func applyPromptScreenPartials(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.PromptScreenPartials, error) {
	screenPartials := make(auth0.PromptScreenPartialsScreenPartialArray, 0, len(locals.ScreenPartials))
	for _, partial := range locals.ScreenPartials {
		screenPartials = append(screenPartials, auth0.PromptScreenPartialsScreenPartialArgs{
			ScreenName: pulumi.String(partial.ScreenName),
			InsertionPoints: auth0.PromptScreenPartialsScreenPartialInsertionPointsArgs{
				FormContent:           pulumi.StringPtrFromPtr(partial.FormContent),
				FormContentStart:      pulumi.StringPtrFromPtr(partial.FormContentStart),
				FormContentEnd:        pulumi.StringPtrFromPtr(partial.FormContentEnd),
				FormFooterStart:       pulumi.StringPtrFromPtr(partial.FormFooterStart),
				FormFooterEnd:         pulumi.StringPtrFromPtr(partial.FormFooterEnd),
				SecondaryActionsStart: pulumi.StringPtrFromPtr(partial.SecondaryActionsStart),
				SecondaryActionsEnd:   pulumi.StringPtrFromPtr(partial.SecondaryActionsEnd),
			},
		})
	}

	partials, err := auth0.NewPromptScreenPartials(ctx, locals.ResourceName, &auth0.PromptScreenPartialsArgs{
		PromptType:     pulumi.String(locals.PromptType),
		ScreenPartials: screenPartials,
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to set the screen partials of the %s prompt for %s", locals.PromptType, locals.ResourceName)
	}
	return partials, nil
}
