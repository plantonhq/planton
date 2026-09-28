package module

import (
	auth0promptv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0prompt/v1alpha1"
)

// Locals holds the values the module computes from the stack input. It mirrors
// the Terraform module's locals.tf -- keep them in lockstep.
type Locals struct {
	// ResourceName is the resource's identity (the Pulumi resource name).
	ResourceName string

	// The three managed settings. A nil pointer is NOT MANAGED: the provider
	// never sends it, and the tenant keeps whatever value it already carries.
	// The empty string is the proto's zero value for an unset experience, so it
	// maps to nil; the optional bools carry their presence already.
	UniversalLoginExperience    *string
	IdentifierFirst             *bool
	WebauthnPlatformFirstFactor *bool
}

func initializeLocals(stackInput *auth0promptv1alpha1.Auth0PromptStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec
	return &Locals{
		ResourceName:                target.Metadata.Name,
		UniversalLoginExperience:    managed(spec.UniversalLoginExperience),
		IdentifierFirst:             spec.IdentifierFirst,
		WebauthnPlatformFirstFactor: spec.WebauthnPlatformFirstFactor,
	}
}

// managed maps the proto's "unset" (the empty string) to nil, so the setting is
// never sent.
func managed(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
