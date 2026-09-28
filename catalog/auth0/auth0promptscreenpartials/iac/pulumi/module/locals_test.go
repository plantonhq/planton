package module

import (
	"testing"

	auth0promptscreenpartialsv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0promptscreenpartials/v1alpha1"
)

func TestScreenPartials(t *testing.T) {
	partials := screenPartials([]*auth0promptscreenpartialsv1alpha1.Auth0PromptScreenPartial{
		{
			ScreenName: "signup-password",
			InsertionPoints: &auth0promptscreenpartialsv1alpha1.Auth0PromptInsertionPoints{
				FormFooterEnd: "<p>Need help? support@planton.ai</p>",
			},
		},
		{
			ScreenName: "signup-id",
			InsertionPoints: &auth0promptscreenpartialsv1alpha1.Auth0PromptInsertionPoints{
				FormContentEnd: "<div class=\"ulp-field\">terms</div>",
			},
		},
	})

	t.Run("screens are ordered by name, as the provider reads them back", func(t *testing.T) {
		if len(partials) != 2 || partials[0].ScreenName != "signup-id" || partials[1].ScreenName != "signup-password" {
			t.Fatalf("got %+v, want signup-id then signup-password", partials)
		}
	})

	t.Run("a declared insertion point is sent as written", func(t *testing.T) {
		if got := partials[0].FormContentEnd; got == nil || *got != "<div class=\"ulp-field\">terms</div>" {
			t.Errorf("form_content_end: got %v", got)
		}
	})

	t.Run("an empty insertion point is never sent", func(t *testing.T) {
		p := partials[0]
		for name, point := range map[string]*string{
			"form_content":            p.FormContent,
			"form_content_start":      p.FormContentStart,
			"form_footer_start":       p.FormFooterStart,
			"form_footer_end":         p.FormFooterEnd,
			"secondary_actions_start": p.SecondaryActionsStart,
			"secondary_actions_end":   p.SecondaryActionsEnd,
		} {
			if point != nil {
				t.Errorf("%s: got %q, want nil", name, *point)
			}
		}
	})
}
