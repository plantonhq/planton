package module

import (
	auth0customdomainv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0customdomain/v1alpha1"
)

// Locals holds the values the module computes from the IaC input. It mirrors
// the Terraform module's locals.tf -- keep them in lockstep.
type Locals struct {
	// ResourceName is the resource's identity (the Pulumi resource name).
	ResourceName string

	Domain string
	Type   string

	// The optional settings. A nil pointer is left to Auth0: the provider never
	// sends it. Empty strings are the proto's zero value for "unset", so they map
	// to nil.
	CustomClientIpHeader   *string
	TlsPolicy              *string
	RelyingPartyIdentifier *string

	// DomainMetadata is sent only when the spec sets at least one pair.
	DomainMetadata map[string]string
}

func initializeLocals(iacInput *auth0customdomainv1alpha1.Auth0CustomDomainIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec
	return &Locals{
		ResourceName:           target.Metadata.Name,
		Domain:                 spec.Domain,
		Type:                   spec.Type,
		CustomClientIpHeader:   optional(spec.CustomClientIpHeader),
		TlsPolicy:              optional(spec.TlsPolicy),
		RelyingPartyIdentifier: optional(spec.RelyingPartyIdentifier),
		DomainMetadata:         spec.DomainMetadata,
	}
}

// optional maps the proto's "unset" (the empty string) to nil, so the setting
// is never sent.
func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
