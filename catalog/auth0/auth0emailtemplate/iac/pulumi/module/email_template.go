package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// applyEmailTemplate customizes one of the emails the tenant the provider's
// credential belongs to sends -- the twin of iac/tf/main.tf. The template name
// is its identity. It sends through the tenant's email provider
// (Auth0EmailProvider): Auth0 refuses custom templates without one. When the
// template already exists, the create takes it over and rewrites it. Auth0
// cannot delete a template: destroy disables it (a PATCH of enabled = false),
// and the tenant sends Auth0's default email again.
func applyEmailTemplate(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.EmailTemplate, error) {
	emailTemplate, err := auth0.NewEmailTemplate(ctx, locals.ResourceName, &auth0.EmailTemplateArgs{
		Template:               pulumi.String(locals.Template),
		From:                   pulumi.String(locals.From),
		Subject:                pulumi.String(locals.Subject),
		Body:                   pulumi.String(locals.Body),
		Syntax:                 pulumi.String(locals.Syntax),
		ResultUrl:              pulumi.StringPtrFromPtr(locals.ResultUrl),
		UrlLifetimeInSeconds:   pulumi.IntPtrFromPtr(locals.UrlLifetimeInSeconds),
		Enabled:                pulumi.Bool(locals.Enabled),
		IncludeEmailInRedirect: pulumi.BoolPtrFromPtr(locals.IncludeEmailInRedirect),
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to apply the %s email template of the Auth0 tenant for %s", locals.Template, locals.ResourceName)
	}
	return emailTemplate, nil
}
