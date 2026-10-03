package auth0tenantsettingsv1alpha1

import (
	"errors"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAuth0TenantSettings(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0TenantSettings Suite")
}

func settings(spec *Auth0TenantSettingsSpec) *Auth0TenantSettings {
	return &Auth0TenantSettings{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0TenantSettings",
		Metadata:   &shared.CatalogObjectMetadata{Name: "tenant-settings"},
		Spec:       spec,
	}
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value}}
}

func reference(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{
		Kind:      kind,
		Name:      name,
		FieldPath: fieldPath,
	}}}
}

func expectValid(spec *Auth0TenantSettingsSpec) {
	gomega.Expect(protovalidate.Validate(settings(spec))).To(gomega.Succeed())
}

// expectInvalid asserts the spec is refused and, when ruleID is given, that
// the error names it (see names).
func expectInvalid(spec *Auth0TenantSettingsSpec, ruleID string) {
	err := protovalidate.Validate(settings(spec))
	gomega.Expect(err).NotTo(gomega.BeNil())
	if ruleID != "" {
		gomega.Expect(names(err, ruleID)).To(gomega.BeTrue(), "expected %q among the violated rules or in the message: %s", ruleID, err.Error())
	}
}

// names reports whether a validation error names want: as the id of a rule it
// violates, or in its message (a field path or the rule's own sentence).
func names(err error, want string) bool {
	for _, id := range violatedRules(err) {
		if id == want {
			return true
		}
	}
	return strings.Contains(err.Error(), want)
}

// violatedRules returns the ids of the rules a validation error names, so a
// test pins the rule it exists for rather than any failure at all.
func violatedRules(err error) []string {
	var validationErr *protovalidate.ValidationError
	if !errors.As(err, &validationErr) {
		return nil
	}
	ids := make([]string, 0, len(validationErr.Violations))
	for _, violation := range validationErr.Violations {
		ids = append(ids, violation.Proto.GetRuleId())
	}
	return ids
}

var _ = ginkgo.Describe("Auth0TenantSettings Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts every presentation setting", func() {
			expectValid(&Auth0TenantSettingsSpec{
				FriendlyName: "Acme",
				PictureUrl:   "https://assets.acme.com/logo.png",
				SupportEmail: "support@acme.com",
				SupportUrl:   "https://acme.com/support",
			})
		})

		ginkgo.It("accepts the default domain alone, read from a verified custom domain", func() {
			expectValid(&Auth0TenantSettingsSpec{
				DefaultCustomDomain: reference(catalogkind.CatalogKind_Auth0CustomDomainVerification,
					"sign-in-domain-verification", "status.outputs.domain"),
			})
		})

		ginkgo.It("accepts the canonical domain as the default", func() {
			expectValid(&Auth0TenantSettingsSpec{DefaultCustomDomain: literal("acme.eu.auth0.com")})
		})

		ginkgo.It("accepts one setting, leaving the others unmanaged", func() {
			expectValid(&Auth0TenantSettingsSpec{FriendlyName: "Acme"})
		})

		ginkgo.It("accepts the whole surface at once", func() {
			expectValid(&Auth0TenantSettingsSpec{
				FriendlyName:                 "Acme",
				EnabledLocales:               []string{"en", "fr-CA", "pt-BR"},
				SandboxVersion:               proto.String("22"),
				SessionLifetime:              proto.Float64(168),
				IdleSessionLifetime:          proto.Float64(0.5),
				EphemeralSessionLifetime:     proto.Float64(72),
				IdleEphemeralSessionLifetime: proto.Float64(0.0167),
				SessionCookie:                &Auth0TenantSettingsSessionCookie{Mode: proto.String("persistent")},
				Sessions: &Auth0TenantSettingsSessions{
					OidcLogoutPromptEnabled: proto.Bool(true),
					Anonymous: &Auth0TenantSettingsAnonymousSessions{
						ActivateCookie:    proto.Bool(false),
						LifetimeInMinutes: proto.Int32(525600),
					},
				},
				ClientIdMetadataDocumentSupported:              proto.Bool(true),
				ResourceParameterProfile:                       proto.String("compatibility"),
				DynamicClientRegistrationSecurityMode:          proto.String("strict"),
				PushedAuthorizationRequestsSupported:           proto.Bool(true),
				AcrValuesSupported:                             []string{"urn:mace:incommon:iap:silver"},
				AllowOrganizationNameInAuthenticationApi:       proto.Bool(false),
				DefaultRedirectionUri:                          proto.String("https://acme.com/login"),
				AllowedLogoutUrls:                              []string{"https://acme.com", "https://*.acme.com/signed-out"},
				OidcLogout:                                     &Auth0TenantSettingsOidcLogout{RpLogoutEndSessionEndpointDiscovery: proto.Bool(true)},
				Mtls:                                           &Auth0TenantSettingsMtls{EnableEndpointAliases: proto.Bool(true)},
				SkipNonVerifiableCallbackUriConfirmationPrompt: proto.Bool(false),
				DefaultAudience: reference(catalogkind.CatalogKind_Auth0ResourceServer,
					"platform-api", "status.outputs.identifier"),
				DefaultDirectory: reference(catalogkind.CatalogKind_Auth0Connection,
					"users", "status.outputs.name"),
				DefaultTokenQuota: &Auth0TenantSettingsDefaultTokenQuota{
					Clients: &Auth0TenantSettingsTokenQuota{ClientCredentials: &Auth0TenantSettingsClientCredentialsQuota{
						Enforce: proto.Bool(true), PerDay: proto.Int32(1000), PerHour: proto.Int32(100),
					}},
					Organizations: &Auth0TenantSettingsTokenQuota{ClientCredentials: &Auth0TenantSettingsClientCredentialsQuota{
						PerDay: proto.Int32(5000),
					}},
				},
				CustomizeMfaInPostloginAction: proto.Bool(true),
				PhoneConsolidatedExperience:   proto.Bool(true),
				CountryCodes:                  &Auth0TenantSettingsCountryCodes{List: []string{"US", "CA"}, Mode: "allow"},
				ErrorPage: &Auth0TenantSettingsErrorPage{
					Url:         proto.String("https://acme.com/error"),
					ShowLogLink: proto.Bool(false),
				},
				Flags: &Auth0TenantSettingsFlags{
					EnableDynamicClientRegistration:   proto.Bool(true),
					EnableClientConnections:           proto.Bool(false),
					EnablePublicSignupUserExistsError: proto.Bool(false),
					AllowLegacyRoGrantTypes:           proto.Bool(false),
				},
			})
		})

		ginkgo.DescribeTable("counts every setting toward at least one",
			func(spec *Auth0TenantSettingsSpec) { expectValid(spec) },
			ginkgo.Entry("enabled_locales", &Auth0TenantSettingsSpec{EnabledLocales: []string{"en"}}),
			ginkgo.Entry("sandbox_version", &Auth0TenantSettingsSpec{SandboxVersion: proto.String("22")}),
			ginkgo.Entry("session_lifetime", &Auth0TenantSettingsSpec{SessionLifetime: proto.Float64(24)}),
			ginkgo.Entry("idle_session_lifetime", &Auth0TenantSettingsSpec{IdleSessionLifetime: proto.Float64(1)}),
			ginkgo.Entry("ephemeral_session_lifetime", &Auth0TenantSettingsSpec{EphemeralSessionLifetime: proto.Float64(8)}),
			ginkgo.Entry("idle_ephemeral_session_lifetime", &Auth0TenantSettingsSpec{IdleEphemeralSessionLifetime: proto.Float64(2)}),
			ginkgo.Entry("session_cookie", &Auth0TenantSettingsSpec{SessionCookie: &Auth0TenantSettingsSessionCookie{Mode: proto.String("non-persistent")}}),
			ginkgo.Entry("sessions", &Auth0TenantSettingsSpec{Sessions: &Auth0TenantSettingsSessions{OidcLogoutPromptEnabled: proto.Bool(false)}}),
			ginkgo.Entry("client_id_metadata_document_supported", &Auth0TenantSettingsSpec{ClientIdMetadataDocumentSupported: proto.Bool(false)}),
			ginkgo.Entry("resource_parameter_profile", &Auth0TenantSettingsSpec{ResourceParameterProfile: proto.String("audience")}),
			ginkgo.Entry("dynamic_client_registration_security_mode", &Auth0TenantSettingsSpec{DynamicClientRegistrationSecurityMode: proto.String("permissive")}),
			ginkgo.Entry("pushed_authorization_requests_supported", &Auth0TenantSettingsSpec{PushedAuthorizationRequestsSupported: proto.Bool(false)}),
			ginkgo.Entry("acr_values_supported", &Auth0TenantSettingsSpec{AcrValuesSupported: []string{"urn:acme:loa:2"}}),
			ginkgo.Entry("disable_acr_values_supported", &Auth0TenantSettingsSpec{DisableAcrValuesSupported: proto.Bool(true)}),
			ginkgo.Entry("allow_organization_name_in_authentication_api", &Auth0TenantSettingsSpec{AllowOrganizationNameInAuthenticationApi: proto.Bool(true)}),
			ginkgo.Entry("default_redirection_uri, even empty", &Auth0TenantSettingsSpec{DefaultRedirectionUri: proto.String("")}),
			ginkgo.Entry("allowed_logout_urls", &Auth0TenantSettingsSpec{AllowedLogoutUrls: []string{"https://acme.com"}}),
			ginkgo.Entry("oidc_logout", &Auth0TenantSettingsSpec{OidcLogout: &Auth0TenantSettingsOidcLogout{RpLogoutEndSessionEndpointDiscovery: proto.Bool(false)}}),
			ginkgo.Entry("mtls", &Auth0TenantSettingsSpec{Mtls: &Auth0TenantSettingsMtls{Disable: proto.Bool(true)}}),
			ginkgo.Entry("skip_non_verifiable_callback_uri_confirmation_prompt", &Auth0TenantSettingsSpec{SkipNonVerifiableCallbackUriConfirmationPrompt: proto.Bool(false)}),
			ginkgo.Entry("default_audience", &Auth0TenantSettingsSpec{DefaultAudience: literal("https://api.acme.com")}),
			ginkgo.Entry("default_directory", &Auth0TenantSettingsSpec{DefaultDirectory: literal("Username-Password-Authentication")}),
			ginkgo.Entry("default_token_quota", &Auth0TenantSettingsSpec{DefaultTokenQuota: &Auth0TenantSettingsDefaultTokenQuota{}}),
			ginkgo.Entry("customize_mfa_in_postlogin_action", &Auth0TenantSettingsSpec{CustomizeMfaInPostloginAction: proto.Bool(false)}),
			ginkgo.Entry("phone_consolidated_experience", &Auth0TenantSettingsSpec{PhoneConsolidatedExperience: proto.Bool(true)}),
			ginkgo.Entry("country_codes", &Auth0TenantSettingsSpec{CountryCodes: &Auth0TenantSettingsCountryCodes{List: []string{"GB"}, Mode: "deny"}}),
			ginkgo.Entry("error_page", &Auth0TenantSettingsSpec{ErrorPage: &Auth0TenantSettingsErrorPage{Html: proto.String("<h1>{{ error | escape }}</h1>")}}),
			ginkgo.Entry("flags", &Auth0TenantSettingsSpec{Flags: &Auth0TenantSettingsFlags{EnableSso: proto.Bool(true)}}),
		)

		ginkgo.It("accepts a lifetime below an hour that is not whole", func() {
			expectValid(&Auth0TenantSettingsSpec{SessionLifetime: proto.Float64(0.25)})
		})

		ginkgo.It("accepts mtls with only one of its switches on", func() {
			expectValid(&Auth0TenantSettingsSpec{Mtls: &Auth0TenantSettingsMtls{Disable: proto.Bool(false), EnableEndpointAliases: proto.Bool(true)}})
		})

		ginkgo.It("accepts disabling ACR values alongside an empty list", func() {
			expectValid(&Auth0TenantSettingsSpec{DisableAcrValuesSupported: proto.Bool(true)})
		})

		ginkgo.It("accepts a false disable_acr_values_supported beside a list", func() {
			expectValid(&Auth0TenantSettingsSpec{DisableAcrValuesSupported: proto.Bool(false), AcrValuesSupported: []string{"urn:acme:loa:2"}})
		})

		ginkgo.It("accepts an error page with an empty url", func() {
			expectValid(&Auth0TenantSettingsSpec{ErrorPage: &Auth0TenantSettingsErrorPage{Url: proto.String("")}})
		})

		ginkgo.It("accepts the quota's upper bound", func() {
			expectValid(&Auth0TenantSettingsSpec{DefaultTokenQuota: &Auth0TenantSettingsDefaultTokenQuota{
				Clients: &Auth0TenantSettingsTokenQuota{ClientCredentials: &Auth0TenantSettingsClientCredentialsQuota{
					PerHour: proto.Int32(2147483647),
				}},
			}})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("refuses a spec that manages nothing", func() {
			expectInvalid(&Auth0TenantSettingsSpec{}, "configure at least one tenant setting")
		})

		ginkgo.It("refuses a support email that is not an address", func() {
			expectInvalid(&Auth0TenantSettingsSpec{SupportEmail: "support"}, "")
		})

		ginkgo.It("refuses a logo that is not a URL", func() {
			expectInvalid(&Auth0TenantSettingsSpec{PictureUrl: "logo.png"}, "")
		})

		ginkgo.It("refuses a support page that is not a URL", func() {
			expectInvalid(&Auth0TenantSettingsSpec{SupportUrl: "acme support"}, "")
		})

		ginkgo.It("refuses an empty default domain", func() {
			expectInvalid(&Auth0TenantSettingsSpec{DefaultCustomDomain: literal("")}, "")
		})

		ginkgo.It("refuses an empty default audience", func() {
			expectInvalid(&Auth0TenantSettingsSpec{DefaultAudience: literal("")}, "")
		})

		ginkgo.It("refuses an empty default directory", func() {
			expectInvalid(&Auth0TenantSettingsSpec{DefaultDirectory: literal("")}, "")
		})

		ginkgo.It("refuses a language Auth0 does not offer", func() {
			expectInvalid(&Auth0TenantSettingsSpec{EnabledLocales: []string{"en", "klingon"}}, "")
		})

		ginkgo.It("refuses a runtime version longer than Auth0 stores", func() {
			expectInvalid(&Auth0TenantSettingsSpec{SandboxVersion: proto.String("123456789")}, "")
		})

		ginkgo.It("refuses a session lifetime under the provider's minimum", func() {
			expectInvalid(&Auth0TenantSettingsSpec{SessionLifetime: proto.Float64(0.001)}, "")
		})

		ginkgo.It("refuses an idle session lifetime under the provider's minimum", func() {
			expectInvalid(&Auth0TenantSettingsSpec{IdleSessionLifetime: proto.Float64(0)}, "")
		})

		ginkgo.It("refuses an ephemeral lifetime under the provider's minimum", func() {
			expectInvalid(&Auth0TenantSettingsSpec{EphemeralSessionLifetime: proto.Float64(0.01)}, "")
		})

		ginkgo.It("refuses an idle ephemeral lifetime under the provider's minimum", func() {
			expectInvalid(&Auth0TenantSettingsSpec{IdleEphemeralSessionLifetime: proto.Float64(0.01)}, "")
		})

		ginkgo.DescribeTable("refuses a lifetime of an hour or more that is not whole",
			func(spec *Auth0TenantSettingsSpec, id string) { expectInvalid(spec, id) },
			ginkgo.Entry("session_lifetime", &Auth0TenantSettingsSpec{SessionLifetime: proto.Float64(1.5)}, "session_lifetime.whole_hours"),
			ginkgo.Entry("idle_session_lifetime", &Auth0TenantSettingsSpec{IdleSessionLifetime: proto.Float64(2.25)}, "idle_session_lifetime.whole_hours"),
			ginkgo.Entry("ephemeral_session_lifetime", &Auth0TenantSettingsSpec{EphemeralSessionLifetime: proto.Float64(12.5)}, "ephemeral_session_lifetime.whole_hours"),
			ginkgo.Entry("idle_ephemeral_session_lifetime", &Auth0TenantSettingsSpec{IdleEphemeralSessionLifetime: proto.Float64(1.1)}, "idle_ephemeral_session_lifetime.whole_hours"),
		)

		ginkgo.It("refuses a session cookie mode Auth0 does not know", func() {
			expectInvalid(&Auth0TenantSettingsSpec{SessionCookie: &Auth0TenantSettingsSessionCookie{Mode: proto.String("forever")}}, "")
		})

		ginkgo.It("refuses sessions without the logout prompt setting", func() {
			expectInvalid(&Auth0TenantSettingsSpec{Sessions: &Auth0TenantSettingsSessions{
				Anonymous: &Auth0TenantSettingsAnonymousSessions{ActivateCookie: proto.Bool(true)},
			}}, "")
		})

		ginkgo.It("refuses an anonymous session lifetime outside a minute to a year", func() {
			expectInvalid(&Auth0TenantSettingsSpec{Sessions: &Auth0TenantSettingsSessions{
				OidcLogoutPromptEnabled: proto.Bool(true),
				Anonymous:               &Auth0TenantSettingsAnonymousSessions{LifetimeInMinutes: proto.Int32(525601)},
			}}, "")
			expectInvalid(&Auth0TenantSettingsSpec{Sessions: &Auth0TenantSettingsSessions{
				OidcLogoutPromptEnabled: proto.Bool(true),
				Anonymous:               &Auth0TenantSettingsAnonymousSessions{LifetimeInMinutes: proto.Int32(0)},
			}}, "")
		})

		ginkgo.It("refuses a resource parameter profile Auth0 does not know", func() {
			expectInvalid(&Auth0TenantSettingsSpec{ResourceParameterProfile: proto.String("resource")}, "")
		})

		ginkgo.It("refuses a dynamic client registration security mode Auth0 does not know", func() {
			expectInvalid(&Auth0TenantSettingsSpec{DynamicClientRegistrationSecurityMode: proto.String("lenient")}, "")
		})

		ginkgo.It("refuses ACR values beside disabling them", func() {
			expectInvalid(&Auth0TenantSettingsSpec{
				DisableAcrValuesSupported: proto.Bool(true),
				AcrValuesSupported:        []string{"urn:acme:loa:2"},
			}, "spec.acr_values_or_disable")
		})

		ginkgo.It("refuses a login route that is not HTTPS", func() {
			expectInvalid(&Auth0TenantSettingsSpec{DefaultRedirectionUri: proto.String("http://acme.com/login")}, "default_redirection_uri.https")
		})

		ginkgo.It("refuses a login route that is not a URL", func() {
			expectInvalid(&Auth0TenantSettingsSpec{DefaultRedirectionUri: proto.String("https://acme.com/log in")}, "default_redirection_uri.https")
		})

		ginkgo.It("refuses oidc_logout without its discovery setting", func() {
			expectInvalid(&Auth0TenantSettingsSpec{OidcLogout: &Auth0TenantSettingsOidcLogout{}}, "")
		})

		ginkgo.It("refuses mtls that both disables and enables aliases", func() {
			expectInvalid(&Auth0TenantSettingsSpec{Mtls: &Auth0TenantSettingsMtls{
				Disable:               proto.Bool(true),
				EnableEndpointAliases: proto.Bool(true),
			}}, "mtls.disable_or_aliases")
		})

		ginkgo.It("refuses a token quota without its client-credentials quota", func() {
			expectInvalid(&Auth0TenantSettingsSpec{DefaultTokenQuota: &Auth0TenantSettingsDefaultTokenQuota{
				Clients: &Auth0TenantSettingsTokenQuota{},
			}}, "")
		})

		ginkgo.It("refuses a quota of zero tokens", func() {
			expectInvalid(&Auth0TenantSettingsSpec{DefaultTokenQuota: &Auth0TenantSettingsDefaultTokenQuota{
				Organizations: &Auth0TenantSettingsTokenQuota{ClientCredentials: &Auth0TenantSettingsClientCredentialsQuota{
					PerDay: proto.Int32(0),
				}},
			}}, "")
			expectInvalid(&Auth0TenantSettingsSpec{DefaultTokenQuota: &Auth0TenantSettingsDefaultTokenQuota{
				Clients: &Auth0TenantSettingsTokenQuota{ClientCredentials: &Auth0TenantSettingsClientCredentialsQuota{
					PerHour: proto.Int32(0),
				}},
			}}, "")
		})

		ginkgo.It("refuses a country filter with no countries", func() {
			expectInvalid(&Auth0TenantSettingsSpec{CountryCodes: &Auth0TenantSettingsCountryCodes{Mode: "allow"}}, "")
		})

		ginkgo.It("refuses a country code that is not two uppercase letters", func() {
			expectInvalid(&Auth0TenantSettingsSpec{CountryCodes: &Auth0TenantSettingsCountryCodes{List: []string{"us"}, Mode: "allow"}}, "")
			expectInvalid(&Auth0TenantSettingsSpec{CountryCodes: &Auth0TenantSettingsCountryCodes{List: []string{"USA"}, Mode: "allow"}}, "")
		})

		ginkgo.It("refuses a country filter without a mode, or with an unknown one", func() {
			expectInvalid(&Auth0TenantSettingsSpec{CountryCodes: &Auth0TenantSettingsCountryCodes{List: []string{"US"}}}, "")
			expectInvalid(&Auth0TenantSettingsSpec{CountryCodes: &Auth0TenantSettingsCountryCodes{List: []string{"US"}, Mode: "block"}}, "")
		})

		ginkgo.It("refuses an error page url that is not a URL", func() {
			expectInvalid(&Auth0TenantSettingsSpec{ErrorPage: &Auth0TenantSettingsErrorPage{Url: proto.String("not a url")}}, "error_page.url_absolute")
		})
	})
})
