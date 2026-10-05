package module

import (
	auth0emailtemplatev1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0emailtemplate/v1alpha1"
)

// Locals holds the values the module computes from the IaC input. It mirrors
// the Terraform module's locals.tf -- keep them in lockstep.
type Locals struct {
	// ResourceName is the resource's identity (the Pulumi resource name).
	ResourceName string

	Template string
	From     string
	Subject  string
	Body     string

	// Syntax and Enabled are required by the provider, so the spec's defaults
	// fill them when unset: "liquid", the only language Auth0 renders, and on.
	Syntax  string
	Enabled bool

	// The optional settings. A nil pointer is left to Auth0: the provider never
	// sends it. An empty result_url is the proto's zero value for "unset", so it
	// maps to nil.
	ResultUrl              *string
	UrlLifetimeInSeconds   *int
	IncludeEmailInRedirect *bool
}

func initializeLocals(iacInput *auth0emailtemplatev1alpha1.Auth0EmailTemplateIacInput) *Locals {
	target := iacInput.Target
	spec := target.Spec

	locals := &Locals{
		ResourceName: target.Metadata.Name,
		Template:     spec.Template,
		From:         spec.From,
		Subject:      spec.Subject,
		Body:         spec.Body,
		Syntax:       "liquid",
		Enabled:      spec.Enabled == nil || spec.GetEnabled(),
		ResultUrl:    optional(spec.ResultUrl),
	}
	if spec.Syntax != nil && spec.GetSyntax() != "" {
		locals.Syntax = spec.GetSyntax()
	}
	if spec.UrlLifetimeInSeconds != nil {
		lifetime := int(spec.GetUrlLifetimeInSeconds())
		locals.UrlLifetimeInSeconds = &lifetime
	}
	if spec.IncludeEmailInRedirect != nil {
		include := spec.GetIncludeEmailInRedirect()
		locals.IncludeEmailInRedirect = &include
	}
	return locals
}

// optional maps the proto's "unset" (the empty string) to nil, so the setting
// is never sent.
func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
