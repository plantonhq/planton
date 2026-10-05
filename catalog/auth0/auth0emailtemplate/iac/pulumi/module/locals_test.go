package module

import (
	"testing"

	auth0emailtemplatev1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0emailtemplate/v1alpha1"
	"github.com/plantonhq/planton/shared"
)

func iacInput(spec *auth0emailtemplatev1alpha1.Auth0EmailTemplateSpec) *auth0emailtemplatev1alpha1.Auth0EmailTemplateIacInput {
	spec.Template = "verify_email"
	spec.From = "Planton <no-reply@planton.ai>"
	spec.Subject = "Verify your email for Planton"
	spec.Body = `<a href="{{ url }}">Verify</a>`
	return &auth0emailtemplatev1alpha1.Auth0EmailTemplateIacInput{
		Target: &auth0emailtemplatev1alpha1.Auth0EmailTemplate{
			Metadata: &shared.CatalogObjectMetadata{Name: "verify-email"},
			Spec:     spec,
		},
	}
}

func TestDefaults(t *testing.T) {
	locals := initializeLocals(iacInput(&auth0emailtemplatev1alpha1.Auth0EmailTemplateSpec{}))
	if locals.Syntax != "liquid" {
		t.Errorf("syntax: got %q, want liquid", locals.Syntax)
	}
	if !locals.Enabled {
		t.Error("enabled: got false, want true when unset")
	}
	if locals.ResultUrl != nil || locals.UrlLifetimeInSeconds != nil || locals.IncludeEmailInRedirect != nil {
		t.Errorf("unset settings must not be sent: got result_url %v, url_lifetime_in_seconds %v, include_email_in_redirect %v",
			locals.ResultUrl, locals.UrlLifetimeInSeconds, locals.IncludeEmailInRedirect)
	}
}

func TestSetValuesAreSent(t *testing.T) {
	disabled := false
	lifetime := int32(3600)
	include := false
	locals := initializeLocals(iacInput(&auth0emailtemplatev1alpha1.Auth0EmailTemplateSpec{
		ResultUrl:              "https://planton.ai",
		UrlLifetimeInSeconds:   &lifetime,
		Enabled:                &disabled,
		IncludeEmailInRedirect: &include,
	}))
	if locals.Enabled {
		t.Error("enabled: got true, want false as set")
	}
	if locals.ResultUrl == nil || *locals.ResultUrl != "https://planton.ai" {
		t.Errorf("result_url: got %v, want https://planton.ai", locals.ResultUrl)
	}
	if locals.UrlLifetimeInSeconds == nil || *locals.UrlLifetimeInSeconds != 3600 {
		t.Errorf("url_lifetime_in_seconds: got %v, want 3600", locals.UrlLifetimeInSeconds)
	}
	if locals.IncludeEmailInRedirect == nil || *locals.IncludeEmailInRedirect {
		t.Errorf("include_email_in_redirect: got %v, want false as set", locals.IncludeEmailInRedirect)
	}
}
