package module

import (
	auth0tenantsettingsv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0tenantsettings/v1alpha1"
)

// Locals holds the values the module computes from the stack input. It mirrors
// the Terraform module's locals.tf -- keep them in lockstep.
type Locals struct {
	// ResourceName is the resource's identity (the Pulumi resource name).
	ResourceName string

	// The four managed settings. A nil pointer is NOT MANAGED: the provider
	// never sends it, and the tenant keeps whatever value it already carries.
	// Empty strings are the proto's zero value for "unset", so they map to nil.
	FriendlyName *string
	PictureUrl   *string
	SupportEmail *string
	SupportUrl   *string
}

func initializeLocals(stackInput *auth0tenantsettingsv1alpha1.Auth0TenantSettingsStackInput) *Locals {
	target := stackInput.Target
	spec := target.Spec
	return &Locals{
		ResourceName: target.Metadata.Name,
		FriendlyName: managed(spec.FriendlyName),
		PictureUrl:   managed(spec.PictureUrl),
		SupportEmail: managed(spec.SupportEmail),
		SupportUrl:   managed(spec.SupportUrl),
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
