package module

import (
	auth0tenantsettingsv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0tenantsettings/v1alpha1"
)

// Locals holds the values the module computes from the IaC input. It mirrors
// the Terraform module's locals.tf -- keep them in lockstep.
type Locals struct {
	// ResourceName is the resource's identity (the Pulumi resource name).
	ResourceName string

	// Spec is the tenant settings as declared. Every setting it leaves unset is
	// NOT MANAGED: tenantArgs never sends it, and the tenant keeps whatever value
	// it already carries.
	Spec *auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec

	// The four presentation settings. A nil pointer is not managed. Empty
	// strings are the proto's zero value for "unset", so they map to nil.
	FriendlyName *string
	PictureUrl   *string
	SupportEmail *string
	SupportUrl   *string

	// DefaultAudience and DefaultDirectory are the resolved values of their
	// references (a reference is resolved to its value before the module runs),
	// or nil when the spec leaves them unmanaged.
	DefaultAudience  *string
	DefaultDirectory *string

	// DefaultCustomDomain is the tenant's default domain, or nil when the spec
	// leaves it unmanaged (the default-domain resource is then not declared).
	DefaultCustomDomain *string
}

func initializeLocals(iacInput *auth0tenantsettingsv1alpha1.Auth0TenantSettingsIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec
	if spec == nil {
		spec = &auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec{}
	}
	return &Locals{
		ResourceName: target.Metadata.Name,
		Spec:         spec,
		FriendlyName: managed(spec.FriendlyName),
		PictureUrl:   managed(spec.PictureUrl),
		SupportEmail: managed(spec.SupportEmail),
		SupportUrl:   managed(spec.SupportUrl),

		DefaultAudience:  managed(spec.DefaultAudience.GetValue()),
		DefaultDirectory: managed(spec.DefaultDirectory.GetValue()),

		DefaultCustomDomain: managed(spec.DefaultCustomDomain.GetValue()),
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
