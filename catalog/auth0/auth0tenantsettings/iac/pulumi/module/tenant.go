package module

import (
	"strconv"

	"github.com/pkg/errors"
	auth0tenantsettingsv1alpha1 "github.com/plantonhq/planton/catalog/auth0/auth0tenantsettings/v1alpha1"
	"github.com/pulumi/pulumi-auth0/sdk/v3/go/auth0"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"google.golang.org/protobuf/proto"
)

// applyTenantSettings manages the settings of the EXISTING tenant the provider's
// credential belongs to; the Management API cannot create or delete a tenant.
// The resource's delete is the provider's no-op: destroy leaves the last-applied
// values in place, as Auth0 has no delete for tenant settings. The Terraform
// module's auth0_tenant (iac/tf/main.tf) is its twin.
func applyTenantSettings(ctx *pulumi.Context, locals *Locals, provider *auth0.Provider) (*auth0.Tenant, error) {
	tenant, err := auth0.NewTenant(ctx, locals.ResourceName, tenantArgs(locals), pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to apply the settings of the Auth0 tenant for %s", locals.ResourceName)
	}
	return tenant, nil
}

// managesTenantSettings reports whether the spec declares any tenant setting
// beyond the default domain, which has a resource of its own. Every tenant
// setting has presence, so a set field (a false toggle included) is a
// declared one.
func managesTenantSettings(spec *auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec) bool {
	settings := proto.Clone(spec).(*auth0tenantsettingsv1alpha1.Auth0TenantSettingsSpec)
	settings.DefaultCustomDomain = nil
	return proto.Size(settings) > 0
}

// tenantArgs maps the spec onto the tenant's arguments, setting each argument
// and block ONLY when the spec sets it: an unset argument is left nil, so the
// provider never sends it and the tenant keeps its value (a tenant adopted
// into this kind previews no change for what the spec never declared). A
// declared block carries only the fields the spec sets inside it.
//
// The provider itself still writes six settings the spec leaves unset on the
// first deploy (and after an import): default_redirection_uri,
// skip_non_verifiable_callback_uri_confirmation_prompt, mtls, error_page,
// default_token_quota and country_codes -- see the spec's comments. Nothing
// here can prevent that; the manifest declares the live values instead.
func tenantArgs(locals *Locals) *auth0.TenantArgs {
	spec := locals.Spec
	return &auth0.TenantArgs{
		// Identity.
		FriendlyName:   pulumi.StringPtrFromPtr(locals.FriendlyName),
		PictureUrl:     pulumi.StringPtrFromPtr(locals.PictureUrl),
		SupportEmail:   pulumi.StringPtrFromPtr(locals.SupportEmail),
		SupportUrl:     pulumi.StringPtrFromPtr(locals.SupportUrl),
		EnabledLocales: stringArray(spec.EnabledLocales),
		SandboxVersion: pulumi.StringPtrFromPtr(spec.SandboxVersion),

		// Sessions.
		SessionLifetime:              pulumi.Float64PtrFromPtr(spec.SessionLifetime),
		IdleSessionLifetime:          pulumi.Float64PtrFromPtr(spec.IdleSessionLifetime),
		EphemeralSessionLifetime:     pulumi.Float64PtrFromPtr(spec.EphemeralSessionLifetime),
		IdleEphemeralSessionLifetime: pulumi.Float64PtrFromPtr(spec.IdleEphemeralSessionLifetime),
		SessionCookie:                sessionCookieArgs(spec.SessionCookie),
		Sessions:                     sessionsArgs(spec.Sessions),

		// OAuth and OpenID Connect.
		ClientIdMetadataDocumentSupported:        pulumi.BoolPtrFromPtr(spec.ClientIdMetadataDocumentSupported),
		ResourceParameterProfile:                 pulumi.StringPtrFromPtr(spec.ResourceParameterProfile),
		DynamicClientRegistrationSecurityMode:    pulumi.StringPtrFromPtr(spec.DynamicClientRegistrationSecurityMode),
		PushedAuthorizationRequestsSupported:     pulumi.BoolPtrFromPtr(spec.PushedAuthorizationRequestsSupported),
		AcrValuesSupporteds:                      stringArray(spec.AcrValuesSupported),
		DisableAcrValuesSupported:                pulumi.BoolPtrFromPtr(spec.DisableAcrValuesSupported),
		AllowOrganizationNameInAuthenticationApi: pulumi.BoolPtrFromPtr(spec.AllowOrganizationNameInAuthenticationApi),
		DefaultRedirectionUri:                    pulumi.StringPtrFromPtr(spec.DefaultRedirectionUri),
		AllowedLogoutUrls:                        stringArray(spec.AllowedLogoutUrls),
		OidcLogout:                               oidcLogoutArgs(spec.OidcLogout),
		Mtls:                                     mtlsArgs(spec.Mtls),
		// The provider takes this one as the string "true" or "false" (its
		// third value, "null", is what leaving it unset sends).
		SkipNonVerifiableCallbackUriConfirmationPrompt: boolString(spec.SkipNonVerifiableCallbackUriConfirmationPrompt),

		// Defaults.
		DefaultAudience:   pulumi.StringPtrFromPtr(locals.DefaultAudience),
		DefaultDirectory:  pulumi.StringPtrFromPtr(locals.DefaultDirectory),
		DefaultTokenQuota: defaultTokenQuotaArgs(spec.DefaultTokenQuota),

		// Sign-in and MFA.
		CustomizeMfaInPostloginAction: pulumi.BoolPtrFromPtr(spec.CustomizeMfaInPostloginAction),
		PhoneConsolidatedExperience:   pulumi.BoolPtrFromPtr(spec.PhoneConsolidatedExperience),
		CountryCodes:                  countryCodesArgs(spec.CountryCodes),

		// The error page.
		ErrorPage: errorPageArgs(spec.ErrorPage),

		// Behavior flags.
		Flags: flagsArgs(spec.Flags),
	}
}

// stringArray sends a list only when it has entries: an empty list is not
// managed. (pulumi.ToStringArray(nil) would send an empty list, clearing the
// tenant's.)
func stringArray(values []string) pulumi.StringArrayInput {
	if len(values) == 0 {
		return nil
	}
	return pulumi.ToStringArray(values)
}

// boolString renders an optional bool as the provider's "true"/"false" string,
// or nil when unset.
func boolString(value *bool) pulumi.StringPtrInput {
	if value == nil {
		return nil
	}
	return pulumi.String(strconv.FormatBool(*value))
}

// intPtr widens an optional int32 to the provider's int, or nil when unset.
func intPtr(value *int32) pulumi.IntPtrInput {
	if value == nil {
		return nil
	}
	return pulumi.Int(int(*value))
}

func sessionCookieArgs(cookie *auth0tenantsettingsv1alpha1.Auth0TenantSettingsSessionCookie) auth0.TenantSessionCookiePtrInput {
	if cookie == nil {
		return nil
	}
	return &auth0.TenantSessionCookieArgs{Mode: pulumi.StringPtrFromPtr(cookie.Mode)}
}

// sessionsArgs declares the sessions block. Its anonymous block is sent only
// when declared; the provider removes the tenant's anonymous-session settings
// when sessions is declared without it.
func sessionsArgs(sessions *auth0tenantsettingsv1alpha1.Auth0TenantSettingsSessions) auth0.TenantSessionsPtrInput {
	if sessions == nil {
		return nil
	}
	args := &auth0.TenantSessionsArgs{
		OidcLogoutPromptEnabled: pulumi.Bool(sessions.GetOidcLogoutPromptEnabled()),
	}
	if anonymous := sessions.GetAnonymous(); anonymous != nil {
		args.Anonymous = &auth0.TenantSessionsAnonymousArgs{
			ActivateCookie:    pulumi.BoolPtrFromPtr(anonymous.ActivateCookie),
			LifetimeInMinutes: intPtr(anonymous.LifetimeInMinutes),
		}
	}
	return args
}

func oidcLogoutArgs(logout *auth0tenantsettingsv1alpha1.Auth0TenantSettingsOidcLogout) auth0.TenantOidcLogoutPtrInput {
	if logout == nil {
		return nil
	}
	return &auth0.TenantOidcLogoutArgs{
		RpLogoutEndSessionEndpointDiscovery: pulumi.Bool(logout.GetRpLogoutEndSessionEndpointDiscovery()),
	}
}

func mtlsArgs(mtls *auth0tenantsettingsv1alpha1.Auth0TenantSettingsMtls) auth0.TenantMtlsPtrInput {
	if mtls == nil {
		return nil
	}
	return &auth0.TenantMtlsArgs{
		Disable:               pulumi.BoolPtrFromPtr(mtls.Disable),
		EnableEndpointAliases: pulumi.BoolPtrFromPtr(mtls.EnableEndpointAliases),
	}
}

// defaultTokenQuotaArgs declares the default token quota. A declared quota
// with neither clients nor organizations is still sent (empty), which the
// provider turns into "remove the tenant's default quotas".
func defaultTokenQuotaArgs(quota *auth0tenantsettingsv1alpha1.Auth0TenantSettingsDefaultTokenQuota) auth0.TenantDefaultTokenQuotaPtrInput {
	if quota == nil {
		return nil
	}
	args := &auth0.TenantDefaultTokenQuotaArgs{}
	if clients := quota.GetClients().GetClientCredentials(); clients != nil {
		args.Clients = &auth0.TenantDefaultTokenQuotaClientsArgs{
			ClientCredentials: auth0.TenantDefaultTokenQuotaClientsClientCredentialsArgs{
				Enforce: pulumi.BoolPtrFromPtr(clients.Enforce),
				PerDay:  intPtr(clients.PerDay),
				PerHour: intPtr(clients.PerHour),
			},
		}
	}
	if organizations := quota.GetOrganizations().GetClientCredentials(); organizations != nil {
		args.Organizations = &auth0.TenantDefaultTokenQuotaOrganizationsArgs{
			ClientCredentials: auth0.TenantDefaultTokenQuotaOrganizationsClientCredentialsArgs{
				Enforce: pulumi.BoolPtrFromPtr(organizations.Enforce),
				PerDay:  intPtr(organizations.PerDay),
				PerHour: intPtr(organizations.PerHour),
			},
		}
	}
	return args
}

func countryCodesArgs(codes *auth0tenantsettingsv1alpha1.Auth0TenantSettingsCountryCodes) auth0.TenantCountryCodesPtrInput {
	if codes == nil {
		return nil
	}
	return &auth0.TenantCountryCodesArgs{
		Lists: pulumi.ToStringArray(codes.GetList()),
		Mode:  pulumi.String(codes.GetMode()),
	}
}

// errorPageArgs declares the error page. A declared page with no html, url or
// show_log_link true returns the tenant to Auth0's default page.
func errorPageArgs(page *auth0tenantsettingsv1alpha1.Auth0TenantSettingsErrorPage) auth0.TenantErrorPagePtrInput {
	if page == nil {
		return nil
	}
	return &auth0.TenantErrorPageArgs{
		Html:        pulumi.StringPtrFromPtr(page.Html),
		ShowLogLink: pulumi.BoolPtrFromPtr(page.ShowLogLink),
		Url:         pulumi.StringPtrFromPtr(page.Url),
	}
}

// flagsArgs declares the flags block with only the flags the spec sets; every
// other flag keeps the tenant's value. The provider sends enable_sso only when
// it changes.
func flagsArgs(flags *auth0tenantsettingsv1alpha1.Auth0TenantSettingsFlags) auth0.TenantFlagsPtrInput {
	if flags == nil {
		return nil
	}
	return &auth0.TenantFlagsArgs{
		AllowLegacyDelegationGrantTypes:    pulumi.BoolPtrFromPtr(flags.AllowLegacyDelegationGrantTypes),
		AllowLegacyRoGrantTypes:            pulumi.BoolPtrFromPtr(flags.AllowLegacyRoGrantTypes),
		AllowLegacyTokeninfoEndpoint:       pulumi.BoolPtrFromPtr(flags.AllowLegacyTokeninfoEndpoint),
		DashboardInsightsView:              pulumi.BoolPtrFromPtr(flags.DashboardInsightsView),
		DashboardLogStreamsNext:            pulumi.BoolPtrFromPtr(flags.DashboardLogStreamsNext),
		DisableClickjackProtectionHeaders:  pulumi.BoolPtrFromPtr(flags.DisableClickjackProtectionHeaders),
		DisableFieldsMapFix:                pulumi.BoolPtrFromPtr(flags.DisableFieldsMapFix),
		DisableManagementApiSmsObfuscation: pulumi.BoolPtrFromPtr(flags.DisableManagementApiSmsObfuscation),
		EnableAdfsWaadEmailVerification:    pulumi.BoolPtrFromPtr(flags.EnableAdfsWaadEmailVerification),
		EnableApisSection:                  pulumi.BoolPtrFromPtr(flags.EnableApisSection),
		EnableClientConnections:            pulumi.BoolPtrFromPtr(flags.EnableClientConnections),
		EnableCustomDomainInEmails:         pulumi.BoolPtrFromPtr(flags.EnableCustomDomainInEmails),
		EnableDynamicClientRegistration:    pulumi.BoolPtrFromPtr(flags.EnableDynamicClientRegistration),
		EnableIdtokenApi2:                  pulumi.BoolPtrFromPtr(flags.EnableIdtokenApi2),
		EnableLegacyLogsSearchV2:           pulumi.BoolPtrFromPtr(flags.EnableLegacyLogsSearchV2),
		EnableLegacyProfile:                pulumi.BoolPtrFromPtr(flags.EnableLegacyProfile),
		EnablePipeline2:                    pulumi.BoolPtrFromPtr(flags.EnablePipeline2),
		EnablePublicSignupUserExistsError:  pulumi.BoolPtrFromPtr(flags.EnablePublicSignupUserExistsError),
		EnableSso:                          pulumi.BoolPtrFromPtr(flags.EnableSso),
		MfaShowFactorListOnEnrollment:      pulumi.BoolPtrFromPtr(flags.MfaShowFactorListOnEnrollment),
		NoDiscloseEnterpriseConnections:    pulumi.BoolPtrFromPtr(flags.NoDiscloseEnterpriseConnections),
		RemoveAlgFromJwks:                  pulumi.BoolPtrFromPtr(flags.RemoveAlgFromJwks),
		RevokeRefreshTokenGrant:            pulumi.BoolPtrFromPtr(flags.RevokeRefreshTokenGrant),
		UseScopeDescriptionsForConsent:     pulumi.BoolPtrFromPtr(flags.UseScopeDescriptionsForConsent),
	}
}
