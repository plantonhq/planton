package module

import (
	"sort"

	auth0promptscreenpartialsv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0promptscreenpartials/v1alpha1"
)

// Locals holds the values the module computes from the IaC input. It mirrors
// the Terraform module's locals.tf -- keep them in lockstep.
type Locals struct {
	// ResourceName is the resource's identity (the Pulumi resource name).
	ResourceName string

	// PromptType is the prompt whose screens the partials extend, and the
	// partials set's identity in Auth0.
	PromptType string

	// ScreenPartials are the partials of each screen, ordered by screen name.
	ScreenPartials []ScreenPartial
}

// ScreenPartial is the partials of one screen. A nil insertion point is never
// sent, and the screen renders nothing there.
type ScreenPartial struct {
	ScreenName            string
	FormContent           *string
	FormContentStart      *string
	FormContentEnd        *string
	FormFooterStart       *string
	FormFooterEnd         *string
	SecondaryActionsStart *string
	SecondaryActionsEnd   *string
}

func initializeLocals(iacInput *auth0promptscreenpartialsv1alpha1.Auth0PromptScreenPartialsIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec
	return &Locals{
		ResourceName:   target.Metadata.Name,
		PromptType:     spec.PromptType,
		ScreenPartials: screenPartials(spec.ScreenPartials),
	}
}

// screenPartials orders the screens by name -- the order the provider reads the
// prompt's partials back in, so a spec that lists them otherwise plans no
// change -- and maps each empty insertion point to nil, so it is never sent.
func screenPartials(partials []*auth0promptscreenpartialsv1alpha1.Auth0PromptScreenPartial) []ScreenPartial {
	result := make([]ScreenPartial, 0, len(partials))
	for _, partial := range partials {
		points := partial.GetInsertionPoints()
		result = append(result, ScreenPartial{
			ScreenName:            partial.GetScreenName(),
			FormContent:           managed(points.GetFormContent()),
			FormContentStart:      managed(points.GetFormContentStart()),
			FormContentEnd:        managed(points.GetFormContentEnd()),
			FormFooterStart:       managed(points.GetFormFooterStart()),
			FormFooterEnd:         managed(points.GetFormFooterEnd()),
			SecondaryActionsStart: managed(points.GetSecondaryActionsStart()),
			SecondaryActionsEnd:   managed(points.GetSecondaryActionsEnd()),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ScreenName < result[j].ScreenName })
	return result
}

// managed maps the proto's "unset" (the empty string) to nil, so the insertion
// point is never sent.
func managed(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
