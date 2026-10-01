package module

import (
	"encoding/json"

	"github.com/pkg/errors"
	auth0promptcustomtextv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0promptcustomtext/v1alpha1"
)

// Locals holds the values the module computes from the stack input. It mirrors
// the Terraform module's locals.tf -- keep them in lockstep.
type Locals struct {
	// ResourceName is the resource's identity (the Pulumi resource name).
	ResourceName string

	// Prompt and Language name the one custom text Auth0 stores.
	Prompt   string
	Language string

	// Body is the prompt's words as the provider takes them: one JSON document,
	// {"<screen>": {"<key>": "<text>"}}.
	Body string
}

func initializeLocals(stackInput *auth0promptcustomtextv1alpha1.Auth0PromptCustomTextStackInput) (*Locals, error) {
	target := stackInput.Target
	spec := target.Spec
	body, err := renderBody(spec.Screens)
	if err != nil {
		return nil, err
	}
	return &Locals{
		ResourceName: target.Metadata.Name,
		Prompt:       spec.Prompt,
		Language:     spec.Language,
		Body:         body,
	}, nil
}

// renderBody renders the screens into the provider's body document. json.Marshal
// writes map keys in sorted order and escapes <, > and & as \u003c, \u003e and
// \u0026, exactly as Terraform's jsonencode does, so both engines send the same
// bytes for the same spec.
func renderBody(screens map[string]*auth0promptcustomtextv1alpha1.Auth0PromptScreenText) (string, error) {
	document := make(map[string]map[string]string, len(screens))
	for screen, text := range screens {
		texts := text.GetTexts()
		if texts == nil {
			texts = map[string]string{}
		}
		document[screen] = texts
	}
	body, err := json.Marshal(document)
	if err != nil {
		return "", errors.Wrap(err, "failed to encode the screens as the custom text document")
	}
	return string(body), nil
}
