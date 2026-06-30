package auth0clientv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAuth0Client(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0Client Suite")
}

var _ = ginkgo.Describe("Auth0Client Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("auth0_client with SPA application type", func() {
			var input *Auth0Client

			ginkgo.BeforeEach(func() {
				input = &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "my-spa-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						Description:     "My React SPA Application",
						Callbacks: []string{
							"https://myapp.com/callback",
							"http://localhost:3000/callback",
						},
						AllowedLogoutUrls: []string{
							"https://myapp.com",
							"http://localhost:3000",
						},
						WebOrigins: []string{
							"https://myapp.com",
							"http://localhost:3000",
						},
						GrantTypes: []string{
							"authorization_code",
							"refresh_token",
						},
						OidcConformant: true,
						IsFirstParty:   true,
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with regular_web application type", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "my-web-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "regular_web",
						Description:     "Traditional web application",
						Callbacks: []string{
							"https://mywebapp.com/auth/callback",
						},
						AllowedLogoutUrls: []string{
							"https://mywebapp.com",
						},
						GrantTypes: []string{
							"authorization_code",
							"refresh_token",
						},
						JwtConfiguration: &Auth0JwtConfiguration{
							LifetimeInSeconds: 36000,
							Alg:               "RS256",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with native application type", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "my-mobile-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "native",
						Description:     "iOS/Android mobile application",
						Callbacks: []string{
							"com.myapp://callback",
							"myapp://callback",
						},
						AllowedLogoutUrls: []string{
							"com.myapp://logout",
						},
						GrantTypes: []string{
							"authorization_code",
							"refresh_token",
						},
						Mobile: &Auth0MobileConfiguration{
							Ios: &Auth0MobileIos{
								TeamId:              "ABCDE12345",
								AppBundleIdentifier: "com.example.myapp",
							},
							Android: &Auth0MobileAndroid{
								AppPackageName:         "com.example.myapp",
								Sha256CertFingerprints: []string{"D8:A0:1B:..."},
							},
						},
						NativeSocialLogin: &Auth0NativeSocialLogin{
							Apple: &Auth0NativeSocialLoginProvider{
								Enabled: true,
							},
							Facebook: &Auth0NativeSocialLoginProvider{
								Enabled: false,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with non_interactive (M2M) application type", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "backend-api-client",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "non_interactive",
						Description:     "Backend service for API access",
						GrantTypes: []string{
							"client_credentials",
						},
						JwtConfiguration: &Auth0JwtConfiguration{
							LifetimeInSeconds: 86400,
							Alg:               "RS256",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with refresh token configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "app-with-refresh-tokens",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						GrantTypes: []string{
							"authorization_code",
							"refresh_token",
						},
						RefreshToken: &Auth0RefreshTokenConfiguration{
							RotationType:      "rotating",
							ExpirationType:    "expiring",
							TokenLifetime:     2592000,
							IdleTokenLifetime: 1296000,
							Leeway:            60,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with organization usage", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "org-enabled-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType:             "spa",
						OrganizationUsage:           "require",
						OrganizationRequireBehavior: "pre_login_prompt",
						Callbacks: []string{
							"https://app.example.com/callback",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with client metadata", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "app-with-metadata",
						Labels: map[string]string{
							"team": "platform",
						},
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "regular_web",
						ClientMetadata: map[string]string{
							"app_version":   "1.2.3",
							"billing_tier":  "enterprise",
							"contact_email": "team@example.com",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with cross-origin authentication", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "embedded-login-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType:           "spa",
						CrossOriginAuthentication: true,
						CrossOriginLoc:            "https://myapp.com/cross-origin-callback",
						Callbacks: []string{
							"https://myapp.com/callback",
						},
						WebOrigins: []string{
							"https://myapp.com",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with OIDC backchannel logout", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "backchannel-logout-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "regular_web",
						OidcBackchannelLogout: &Auth0OidcBackchannelLogout{
							BackchannelLogoutUrls: []string{
								"https://myapp.com/backchannel-logout",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with all JWT algorithms", func() {
			ginkgo.It("should not return a validation error for HS256", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "hs256-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "regular_web",
						JwtConfiguration: &Auth0JwtConfiguration{
							Alg:           "HS256",
							SecretEncoded: true,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for PS256", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "ps256-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "regular_web",
						JwtConfiguration: &Auth0JwtConfiguration{
							Alg: "PS256",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with api_grants using direct value", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "m2m-with-api-grant",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "non_interactive",
						Description:     "M2M client with API access",
						GrantTypes:      []string{"client_credentials"},
						ApiGrants: []*Auth0ClientApiGrant{
							{
								Audience: &foreignkeyv1.StringValueOrRef{
									LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
										Value: "https://api.example.com/",
									},
								},
								Scopes: []string{"read:resources", "write:resources"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with api_grants using reference to Auth0ResourceServer", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "m2m-with-resource-server-ref",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "non_interactive",
						Description:     "M2M client referencing Auth0ResourceServer",
						GrantTypes:      []string{"client_credentials"},
						ApiGrants: []*Auth0ClientApiGrant{
							{
								Audience: &foreignkeyv1.StringValueOrRef{
									LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
										ValueFrom: &foreignkeyv1.ValueFromRef{
											Name: "my-api-resource-server",
										},
									},
								},
								Scopes: []string{"read:data"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with enabled_connections using direct value", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "spa-with-connections",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						Callbacks:       []string{"https://myapp.com/callback"},
						EnabledConnections: []*foreignkeyv1.StringValueOrRef{
							{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "Username-Password-Authentication"}},
							{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "google-oauth2"}},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with enabled_connections using reference to Auth0Connection", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "spa-with-connection-refs",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						Callbacks:       []string{"https://myapp.com/callback"},
						EnabledConnections: []*foreignkeyv1.StringValueOrRef{
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
									ValueFrom: &foreignkeyv1.ValueFromRef{
										Name: "my-database-connection",
									},
								},
							},
							{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
									ValueFrom: &foreignkeyv1.ValueFromRef{
										Name: "my-google-connection",
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_client with multiple api_grants to Management API and custom API", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "platform-admin-service",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "non_interactive",
						Description:     "Platform admin with multi-API access",
						GrantTypes:      []string{"client_credentials"},
						ApiGrants: []*Auth0ClientApiGrant{
							{
								Audience: &foreignkeyv1.StringValueOrRef{
									LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
										Value: "https://my-tenant.us.auth0.com/api/v2/",
									},
								},
								Scopes: []string{"read:users", "update:users"},
							},
							{
								Audience: &foreignkeyv1.StringValueOrRef{
									LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
										ValueFrom: &foreignkeyv1.ValueFromRef{
											Name: "my-custom-api",
										},
									},
								},
								Scopes: []string{"admin:resources"},
							},
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
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata:   nil,
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required spec", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: nil,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect api_version", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "wrong.api.version/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect kind", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required application_type", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "",
						Description:     "Missing app type",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid application_type value", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "invalid_app_type",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("description too long", func() {
			ginkgo.It("should return a validation error", func() {
				longDescription := "This is a very long description that exceeds the 140 character limit for Auth0 client descriptions. It should trigger a validation error because it is way too long for the field."
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						Description:     longDescription,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid JWT algorithm", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						JwtConfiguration: &Auth0JwtConfiguration{
							Alg: "ES256",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("JWT lifetime out of range (too high)", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						JwtConfiguration: &Auth0JwtConfiguration{
							LifetimeInSeconds: 3000000,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("JWT lifetime out of range (negative)", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						JwtConfiguration: &Auth0JwtConfiguration{
							LifetimeInSeconds: -1,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid refresh token rotation_type", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						RefreshToken: &Auth0RefreshTokenConfiguration{
							RotationType: "always-rotate",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid refresh token expiration_type", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						RefreshToken: &Auth0RefreshTokenConfiguration{
							ExpirationType: "auto-expire",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("negative token_lifetime", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						RefreshToken: &Auth0RefreshTokenConfiguration{
							TokenLifetime: -100,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("negative idle_token_lifetime", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						RefreshToken: &Auth0RefreshTokenConfiguration{
							IdleTokenLifetime: -50,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("negative leeway", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "spa",
						RefreshToken: &Auth0RefreshTokenConfiguration{
							Leeway: -10,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid organization_usage value", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType:   "spa",
						OrganizationUsage: "maybe",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid organization_require_behavior value", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType:             "spa",
						OrganizationUsage:           "require",
						OrganizationRequireBehavior: "always_prompt",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("api_grants with missing required audience", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "non_interactive",
						GrantTypes:      []string{"client_credentials"},
						ApiGrants: []*Auth0ClientApiGrant{
							{
								Audience: nil,
								Scopes:   []string{"read:resources"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("api_grants with reference missing required name", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0Client{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0Client",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-app",
					},
					Spec: &Auth0ClientSpec{
						ApplicationType: "non_interactive",
						GrantTypes:      []string{"client_credentials"},
						ApiGrants: []*Auth0ClientApiGrant{
							{
								Audience: &foreignkeyv1.StringValueOrRef{
									LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
										ValueFrom: &foreignkeyv1.ValueFromRef{
											Name: "", // Name is required
										},
									},
								},
								Scopes: []string{"read:resources"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
