package auth0connectionv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAuth0Connection(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0Connection Suite")
}

var _ = ginkgo.Describe("Auth0Connection Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("auth0_connection with database strategy (auth0)", func() {
			var input *Auth0Connection

			ginkgo.BeforeEach(func() {
				input = &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "my-user-database",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy:    "auth0",
						DisplayName: "Email Sign Up",
						EnabledClients: []*foreignkeyv1.StringValueOrRef{
							{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "client-id-1"}},
							{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "client-id-2"}},
						},
						DatabaseOptions: &Auth0DatabaseOptions{
							PasswordPolicy:         "good",
							BruteForceProtection:   true,
							PasswordHistorySize:    5,
							PasswordNoPersonalInfo: true,
							PasswordDictionary:     true,
						},
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_connection with Google OAuth strategy", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "google-social-login",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy:    "google-oauth2",
						DisplayName: "Sign in with Google",
						EnabledClients: []*foreignkeyv1.StringValueOrRef{
							{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-app-client-id"}},
						},
						SocialOptions: &Auth0SocialOptions{
							ClientId:     "google-client-id.apps.googleusercontent.com",
							ClientSecret: "google-client-secret",
							Scopes:       []string{"openid", "profile", "email"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_connection with GitHub OAuth strategy", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "github-login",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy:    "github",
						DisplayName: "Continue with GitHub",
						SocialOptions: &Auth0SocialOptions{
							ClientId:     "github-oauth-app-id",
							ClientSecret: "github-oauth-app-secret",
							Scopes:       []string{"read:user", "user:email"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_connection with SAML enterprise strategy", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "okta-saml-sso",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy:           "samlp",
						DisplayName:        "Company SSO",
						IsDomainConnection: true,
						Realms:             []string{"company.com", "company.io"},
						SamlOptions: &Auth0SamlOptions{
							SignInEndpoint:     "https://company.okta.com/app/sso/saml",
							SigningCert:        "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
							EntityId:           "http://www.okta.com/exk123abc",
							SignRequest:        true,
							SignatureAlgorithm: "rsa-sha256",
							DigestAlgorithm:    "sha256",
							AttributeMappings: map[string]string{
								"email":       "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
								"given_name":  "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
								"family_name": "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_connection with OIDC enterprise strategy", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "keycloak-oidc",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy:    "oidc",
						DisplayName: "Login with Keycloak",
						OidcOptions: &Auth0OidcOptions{
							Issuer:       "https://keycloak.example.com/realms/my-realm",
							ClientId:     "auth0-client",
							ClientSecret: "keycloak-client-secret",
							Scopes:       []string{"openid", "profile", "email"},
							Type:         "front_channel",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_connection with Azure AD enterprise strategy", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "azure-ad-enterprise",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy:           "waad",
						DisplayName:        "Microsoft Entra ID",
						IsDomainConnection: true,
						AzureAdOptions: &Auth0AzureAdOptions{
							ClientId:                 "azure-app-id-guid",
							ClientSecret:             "azure-client-secret",
							Domain:                   "contoso.onmicrosoft.com",
							TenantId:                 "tenant-guid",
							UseCommonEndpoint:        false,
							MaxGroupsToRetrieve:      50,
							ShouldTrustEmailVerified: true,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_connection with Facebook strategy", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "facebook-login",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy:     "facebook",
						DisplayName:  "Continue with Facebook",
						ShowAsButton: true,
						SocialOptions: &Auth0SocialOptions{
							ClientId:     "facebook-app-id",
							ClientSecret: "facebook-app-secret",
							Scopes:       []string{"email", "public_profile"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_connection with metadata", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "custom-connection",
						Labels: map[string]string{
							"team": "identity",
						},
					},
					Spec: &Auth0ConnectionSpec{
						Strategy:    "auth0",
						DisplayName: "Custom Database",
						Metadata: map[string]string{
							"integration-id": "int-12345",
							"created-by":     "platform-team",
						},
						DatabaseOptions: &Auth0DatabaseOptions{
							PasswordPolicy:         "excellent",
							BruteForceProtection:   true,
							PasswordHistorySize:    10,
							PasswordNoPersonalInfo: true,
							PasswordDictionary:     true,
							MfaEnabled:             true,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing required metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata:   nil,
					Spec: &Auth0ConnectionSpec{
						Strategy: "auth0",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required spec", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-connection",
					},
					Spec: nil,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect api_version", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "wrong.api.version/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-connection",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "auth0",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect kind", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-connection",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "auth0",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required strategy", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-connection",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy:    "",
						DisplayName: "Missing Strategy",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid strategy value", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-connection",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "invalid-strategy-type",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid password_policy value", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-connection",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "auth0",
						DatabaseOptions: &Auth0DatabaseOptions{
							PasswordPolicy: "super-strong",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("password_history_size out of range (too high)", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-connection",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "auth0",
						DatabaseOptions: &Auth0DatabaseOptions{
							PasswordPolicy:      "good",
							PasswordHistorySize: 50,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("password_history_size out of range (negative)", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-connection",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "auth0",
						DatabaseOptions: &Auth0DatabaseOptions{
							PasswordPolicy:      "good",
							PasswordHistorySize: -1,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("social_options missing client_id", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-google",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "google-oauth2",
						SocialOptions: &Auth0SocialOptions{
							ClientId:     "",
							ClientSecret: "some-secret",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("social_options missing client_secret", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-google",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "google-oauth2",
						SocialOptions: &Auth0SocialOptions{
							ClientId:     "some-client-id",
							ClientSecret: "",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("saml_options missing sign_in_endpoint", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-saml",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "samlp",
						SamlOptions: &Auth0SamlOptions{
							SignInEndpoint: "",
							SigningCert:    "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("saml_options missing signing_cert", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-saml",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "samlp",
						SamlOptions: &Auth0SamlOptions{
							SignInEndpoint: "https://idp.example.com/sso",
							SigningCert:    "",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("saml_options invalid protocol_binding", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-saml",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "samlp",
						SamlOptions: &Auth0SamlOptions{
							SignInEndpoint:  "https://idp.example.com/sso",
							SigningCert:     "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
							ProtocolBinding: "invalid-binding",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("saml_options invalid signature_algorithm", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-saml",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "samlp",
						SamlOptions: &Auth0SamlOptions{
							SignInEndpoint:     "https://idp.example.com/sso",
							SigningCert:        "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
							SignatureAlgorithm: "rsa-sha512",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("oidc_options missing issuer", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-oidc",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "oidc",
						OidcOptions: &Auth0OidcOptions{
							Issuer:   "",
							ClientId: "some-client-id",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("oidc_options missing client_id", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-oidc",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "oidc",
						OidcOptions: &Auth0OidcOptions{
							Issuer:   "https://idp.example.com",
							ClientId: "",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("oidc_options invalid type", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-oidc",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "oidc",
						OidcOptions: &Auth0OidcOptions{
							Issuer:   "https://idp.example.com",
							ClientId: "client-id",
							Type:     "invalid_channel",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("azure_ad_options missing client_id", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-waad",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "waad",
						AzureAdOptions: &Auth0AzureAdOptions{
							ClientId:     "",
							ClientSecret: "secret",
							Domain:       "contoso.onmicrosoft.com",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("azure_ad_options missing client_secret", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-waad",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "waad",
						AzureAdOptions: &Auth0AzureAdOptions{
							ClientId:     "client-id",
							ClientSecret: "",
							Domain:       "contoso.onmicrosoft.com",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("azure_ad_options missing domain", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-waad",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "waad",
						AzureAdOptions: &Auth0AzureAdOptions{
							ClientId:     "client-id",
							ClientSecret: "secret",
							Domain:       "",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("azure_ad_options negative max_groups_to_retrieve", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Connection{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Connection",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-waad",
					},
					Spec: &Auth0ConnectionSpec{
						Strategy: "waad",
						AzureAdOptions: &Auth0AzureAdOptions{
							ClientId:            "client-id",
							ClientSecret:        "secret",
							Domain:              "contoso.onmicrosoft.com",
							MaxGroupsToRetrieve: -5,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
