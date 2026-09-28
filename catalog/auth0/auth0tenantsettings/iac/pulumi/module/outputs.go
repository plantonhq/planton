package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// exportOutputs exports the tenant settings as the tenant carries them after the
// apply, managed or not, and the default domain when this resource sets it. The
// Terraform module's outputs.tf exports the same names -- keep them in lockstep.
func exportOutputs(ctx *pulumi.Context, tenant *auth0.Tenant, defaultDomain *auth0.CustomDomainDefault) error {
	ctx.Export("friendly_name", tenant.FriendlyName)
	ctx.Export("picture_url", tenant.PictureUrl)
	ctx.Export("support_email", tenant.SupportEmail)
	ctx.Export("support_url", tenant.SupportUrl)

	if defaultDomain != nil {
		ctx.Export("default_custom_domain", defaultDomain.Domain)
	} else {
		ctx.Export("default_custom_domain", pulumi.String(""))
	}

	ctx.Export("default_audience", tenant.DefaultAudience)
	ctx.Export("default_directory", tenant.DefaultDirectory)
	ctx.Export("client_id_metadata_document_supported", tenant.ClientIdMetadataDocumentSupported)
	ctx.Export("resource_parameter_profile", tenant.ResourceParameterProfile)
	ctx.Export("enable_dynamic_client_registration", tenant.Flags.EnableDynamicClientRegistration())
	ctx.Export("dynamic_client_registration_security_mode", tenant.DynamicClientRegistrationSecurityMode)
	ctx.Export("session_lifetime", tenant.SessionLifetime)
	ctx.Export("idle_session_lifetime", tenant.IdleSessionLifetime)
	ctx.Export("session_cookie_mode", tenant.SessionCookie.Mode())
	ctx.Export("enabled_locales", tenant.EnabledLocales)

	return nil
}
