package module

import (
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// tenantOutputs are the tenant's settings as it carries them after the apply,
// managed or not: read from the tenant resource when this resource manages any
// setting, and from the provider's read-only tenant lookup when it manages only
// the default domain. The Terraform module's outputs.tf exports the same names
// -- keep them in lockstep.
type tenantOutputs struct {
	FriendlyName                          pulumi.StringOutput
	PictureUrl                            pulumi.StringOutput
	SupportEmail                          pulumi.StringOutput
	SupportUrl                            pulumi.StringOutput
	DefaultAudience                       pulumi.StringOutput
	DefaultDirectory                      pulumi.StringOutput
	ClientIdMetadataDocumentSupported     pulumi.BoolOutput
	ResourceParameterProfile              pulumi.StringOutput
	EnableDynamicClientRegistration       pulumi.BoolPtrOutput
	DynamicClientRegistrationSecurityMode pulumi.StringOutput
	SessionLifetime                       pulumi.Float64Output
	IdleSessionLifetime                   pulumi.Float64Output
	SessionCookieMode                     pulumi.StringPtrOutput
	EnabledLocales                        pulumi.StringArrayOutput
}

// managedTenantOutputs reads the outputs from the tenant resource.
func managedTenantOutputs(tenant *auth0.Tenant) tenantOutputs {
	return tenantOutputs{
		FriendlyName:                          tenant.FriendlyName,
		PictureUrl:                            tenant.PictureUrl,
		SupportEmail:                          tenant.SupportEmail,
		SupportUrl:                            tenant.SupportUrl,
		DefaultAudience:                       tenant.DefaultAudience,
		DefaultDirectory:                      tenant.DefaultDirectory,
		ClientIdMetadataDocumentSupported:     tenant.ClientIdMetadataDocumentSupported,
		ResourceParameterProfile:              tenant.ResourceParameterProfile,
		EnableDynamicClientRegistration:       tenant.Flags.EnableDynamicClientRegistration(),
		DynamicClientRegistrationSecurityMode: tenant.DynamicClientRegistrationSecurityMode,
		SessionLifetime:                       tenant.SessionLifetime,
		IdleSessionLifetime:                   tenant.IdleSessionLifetime,
		SessionCookieMode:                     tenant.SessionCookie.Mode(),
		EnabledLocales:                        tenant.EnabledLocales,
	}
}

// lookedUpTenantOutputs reads the outputs from the provider's tenant lookup,
// which writes nothing.
func lookedUpTenantOutputs(current auth0.LookupTenantResultOutput) tenantOutputs {
	return tenantOutputs{
		FriendlyName:                          current.FriendlyName(),
		PictureUrl:                            current.PictureUrl(),
		SupportEmail:                          current.SupportEmail(),
		SupportUrl:                            current.SupportUrl(),
		DefaultAudience:                       current.DefaultAudience(),
		DefaultDirectory:                      current.DefaultDirectory(),
		ClientIdMetadataDocumentSupported:     current.ClientIdMetadataDocumentSupported(),
		ResourceParameterProfile:              current.ResourceParameterProfile(),
		EnableDynamicClientRegistration:       lookedUpDynamicClientRegistration(current),
		DynamicClientRegistrationSecurityMode: current.DynamicClientRegistrationSecurityMode(),
		SessionLifetime:                       current.SessionLifetime(),
		IdleSessionLifetime:                   current.IdleSessionLifetime(),
		SessionCookieMode:                     lookedUpSessionCookieMode(current),
		EnabledLocales:                        current.EnabledLocales(),
	}
}

// The lookup reports its blocks as lists, and a provider-generated list's
// Index does not guard its bounds (it panics on an empty list), so each block's
// value is read from the result itself.
func lookedUpDynamicClientRegistration(current auth0.LookupTenantResultOutput) pulumi.BoolPtrOutput {
	return current.ApplyT(func(r auth0.LookupTenantResult) *bool {
		if len(r.Flags) == 0 {
			return nil
		}
		enabled := r.Flags[0].EnableDynamicClientRegistration
		return &enabled
	}).(pulumi.BoolPtrOutput)
}

func lookedUpSessionCookieMode(current auth0.LookupTenantResultOutput) pulumi.StringPtrOutput {
	return current.ApplyT(func(r auth0.LookupTenantResult) *string {
		if len(r.SessionCookies) == 0 {
			return nil
		}
		mode := r.SessionCookies[0].Mode
		return &mode
	}).(pulumi.StringPtrOutput)
}

// exportOutputs exports the tenant's settings and the default domain when this
// resource sets it.
func exportOutputs(ctx *pulumi.Context, tenant tenantOutputs, defaultDomain *auth0.CustomDomainDefault) error {
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
	ctx.Export("enable_dynamic_client_registration", tenant.EnableDynamicClientRegistration)
	ctx.Export("dynamic_client_registration_security_mode", tenant.DynamicClientRegistrationSecurityMode)
	ctx.Export("session_lifetime", tenant.SessionLifetime)
	ctx.Export("idle_session_lifetime", tenant.IdleSessionLifetime)
	ctx.Export("session_cookie_mode", tenant.SessionCookieMode)
	ctx.Export("enabled_locales", tenant.EnabledLocales)

	return nil
}
