package auth0resourceserverv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestAuth0ResourceServer(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Auth0ResourceServer Suite")
}

var _ = ginkgo.Describe("Auth0ResourceServer Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("auth0_resource_server with minimal configuration", func() {
			var input *Auth0ResourceServer

			ginkgo.BeforeEach(func() {
				input = &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "my-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier: "https://api.example.com/",
					},
				}
			})

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_resource_server with full configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:          "https://api.example.com/v1",
						Name:                "My Full API",
						SigningAlg:          "RS256",
						AllowOfflineAccess:  true,
						TokenLifetime:       86400,
						TokenLifetimeForWeb: 7200,
						SkipConsentForVerifiableFirstPartyClients: true,
						EnforcePolicies: true,
						TokenDialect:    "access_token_authz",
						Scopes: []*Auth0ResourceServerScope{
							{
								Name:        "read:users",
								Description: "Read access to user profiles",
							},
							{
								Name:        "write:users",
								Description: "Create and update users",
							},
							{
								Name:        "delete:users",
								Description: "Delete users",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_resource_server with RS256 signing algorithm", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "rs256-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier: "https://api.rs256.com/",
						SigningAlg: "RS256",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_resource_server with HS256 signing algorithm", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "hs256-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier: "https://api.hs256.com/",
						SigningAlg: "HS256",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_resource_server with PS256 signing algorithm", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "ps256-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier: "https://api.ps256.com/",
						SigningAlg: "PS256",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_resource_server with all token dialects", func() {
			ginkgo.It("should not return a validation error for access_token", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "access-token-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:   "https://api.at.com/",
						TokenDialect: "access_token",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for access_token_authz", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "authz-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:   "https://api.authz.com/",
						TokenDialect: "access_token_authz",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for rfc9068_profile", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "rfc9068-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:   "https://api.rfc.com/",
						TokenDialect: "rfc9068_profile",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for rfc9068_profile_authz", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "rfc9068-authz-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:   "https://api.rfcauthz.com/",
						TokenDialect: "rfc9068_profile_authz",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_resource_server with RBAC configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "rbac-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:      "https://api.rbac.com/",
						EnforcePolicies: true,
						TokenDialect:    "access_token_authz",
						Scopes: []*Auth0ResourceServerScope{
							{
								Name:        "read:items",
								Description: "Read items",
							},
							{
								Name:        "write:items",
								Description: "Write items",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_resource_server with valid token lifetime bounds", func() {
			ginkgo.It("should not return a validation error for minimum token_lifetime", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "min-lifetime-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:    "https://api.min.com/",
						TokenLifetime: 0,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for maximum token_lifetime", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "max-lifetime-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:    "https://api.max.com/",
						TokenLifetime: 2592000,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for valid token_lifetime_for_web", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "web-lifetime-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:          "https://api.web.com/",
						TokenLifetime:       86400,
						TokenLifetimeForWeb: 7200,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("auth0_resource_server with scopes", func() {
			ginkgo.It("should not return a validation error for multiple scopes", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "scoped-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier: "https://api.scoped.com/",
						Scopes: []*Auth0ResourceServerScope{
							{
								Name:        "read:products",
								Description: "Read product catalog",
							},
							{
								Name:        "write:products",
								Description: "Create and update products",
							},
							{
								Name:        "delete:products",
								Description: "Delete products from catalog",
							},
							{
								Name:        "read:orders",
								Description: "Read order history",
							},
							{
								Name:        "write:orders",
								Description: "Create and update orders",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for scope without description", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "nodesc-scope-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier: "https://api.nodesc.com/",
						Scopes: []*Auth0ResourceServerScope{
							{
								Name: "read:data",
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
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata:   nil,
					Spec: &Auth0ResourceServerSpec{
						Identifier: "https://api.example.com/",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required spec", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-api",
					},
					Spec: nil,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect api_version", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "wrong.api.version/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier: "https://api.example.com/",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("incorrect kind", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier: "https://api.example.com/",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("missing required identifier", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier: "",
						Name:       "Missing Identifier API",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid signing_alg value", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier: "https://api.example.com/",
						SigningAlg: "ES256",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid token_dialect value", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:   "https://api.example.com/",
						TokenDialect: "jwt_token",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("token_lifetime exceeds maximum", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:    "https://api.example.com/",
						TokenLifetime: 3000000,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("token_lifetime is negative", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:    "https://api.example.com/",
						TokenLifetime: -100,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("token_lifetime_for_web exceeds maximum", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:          "https://api.example.com/",
						TokenLifetimeForWeb: 3000000,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("token_lifetime_for_web is negative", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:          "https://api.example.com/",
						TokenLifetimeForWeb: -50,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("scope with missing required name", func() {
			ginkgo.It("should return a validation error", func() {
				input := &Auth0ResourceServer{
					ApiVersion: "auth0.planton.dev/v1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier: "https://api.example.com/",
						Scopes: []*Auth0ResourceServerScope{
							{
								Name:        "",
								Description: "Scope without name",
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
