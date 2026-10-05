package auth0resourceserverv1alpha1

import (
	"errors"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
						Name: "full-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:          "https://api.example.com/v1",
						Name:                "My Full API",
						SigningAlg:          "RS256",
						AllowOfflineAccess:  proto.Bool(true),
						TokenLifetime:       86400,
						TokenLifetimeForWeb: 7200,
						SkipConsentForVerifiableFirstPartyClients: proto.Bool(true),
						EnforcePolicies: proto.Bool(true),
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
						Name: "rbac-api",
					},
					Spec: &Auth0ResourceServerSpec{
						Identifier:      "https://api.rbac.com/",
						EnforcePolicies: proto.Bool(true),
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "WrongKind",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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
					ApiVersion: "auth0.planton.dev/v1alpha1",
					Kind:       "Auth0ResourceServer",
					Metadata: &shared.CatalogObjectMetadata{
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

// apiWith wraps a spec (its identifier filled in) in a complete resource.
func apiWith(spec *Auth0ResourceServerSpec) *Auth0ResourceServer {
	if spec.Identifier == "" {
		spec.Identifier = "https://api.example.com/"
	}
	return &Auth0ResourceServer{
		ApiVersion: "auth0.planton.dev/v1alpha1",
		Kind:       "Auth0ResourceServer",
		Metadata:   &shared.CatalogObjectMetadata{Name: "api"},
		Spec:       spec,
	}
}

func expectValid(spec *Auth0ResourceServerSpec) {
	gomega.Expect(protovalidate.Validate(apiWith(spec))).To(gomega.Succeed())
}

// expectInvalid asserts the spec is refused and the error names ruleID (see names).
func expectInvalid(spec *Auth0ResourceServerSpec, ruleID string) {
	err := protovalidate.Validate(apiWith(spec))
	gomega.Expect(err).NotTo(gomega.BeNil())
	gomega.Expect(names(err, ruleID)).To(gomega.BeTrue(), "expected %q among the violated rules or in the message: %s", ruleID, err.Error())
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

func validKey() *Auth0ResourceServerTokenEncryptionKey {
	return &Auth0ResourceServerTokenEncryptionKey{
		Algorithm: "RSA-OAEP-256",
		Pem:       "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA\n-----END PUBLIC KEY-----\n",
	}
}

var _ = ginkgo.Describe("Auth0ResourceServer token, access and grant settings", func() {

	ginkgo.Context("an API declaring every setting", func() {
		ginkgo.It("is valid", func() {
			expectValid(&Auth0ResourceServerSpec{
				AllowOnlineAccess:                      proto.Bool(true),
				AllowOnlineAccessWithEphemeralSessions: proto.Bool(false),
				ConsentPolicy:                          proto.String("transactional-authorization-with-mfa"),
				TokenLifetimeForAnonymousAccessTokens:  proto.Int32(86400),
				VerificationLocation:                   proto.String("https://api.example.com/.well-known/jwks.json"),
				SigningAlg:                             "HS256",
				SigningSecret:                          proto.String("a-shared-secret-of-32-characters"),
				AccessToken: &Auth0ResourceServerAccessToken{
					ClaimsMapping: &Auth0ResourceServerClaimsMapping{
						CustomClaims: []*Auth0ResourceServerCustomClaim{
							{Name: "country", Expression: "anonymous_session.metadata.country"},
							{Name: "tier", Expression: "anonymous_session.metadata.plan-tier"},
						},
					},
				},
				AuthorizationDetails: []*Auth0ResourceServerAuthorizationDetail{
					{Type: proto.String("payment")},
					{Type: proto.String("money_transfer")},
				},
				AuthorizationPolicy: &Auth0ResourceServerAuthorizationPolicy{PolicyId: proto.String("pol_123")},
				ProofOfPossession: &Auth0ResourceServerProofOfPossession{
					Mechanism:   proto.String("dpop"),
					Required:    proto.Bool(true),
					RequiredFor: proto.String("public_clients"),
				},
				SubjectTypeAuthorization: &Auth0ResourceServerSubjectTypeAuthorization{
					User:          &Auth0ResourceServerUserAuthorization{Policy: proto.String("require_client_grant")},
					Client:        &Auth0ResourceServerClientAuthorization{Policy: proto.String("deny_all")},
					AnonymousUser: &Auth0ResourceServerAnonymousUserAuthorization{Policy: proto.String("deny_all")},
				},
				TokenEncryption: &Auth0ResourceServerTokenEncryption{
					Format:        proto.String("compact-nested-jwe"),
					EncryptionKey: validKey(),
				},
				Scopes: []*Auth0ResourceServerScope{{Name: "read:items"}, {Name: "write:items"}},
				ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
					{SubjectType: "user", Scopes: []string{"read:items"}, AuthorizationDetailsTypes: []string{"payment"}},
					{SubjectType: "client", AllowAllScopes: proto.Bool(true), OrganizationUsage: proto.String("allow")},
				},
			})
		})
	})

	ginkgo.Context("consent_policy", func() {
		ginkgo.It("accepts null, the standard policy", func() {
			expectValid(&Auth0ResourceServerSpec{ConsentPolicy: proto.String("null")})
		})
		ginkgo.It("rejects a policy Auth0 does not offer", func() {
			expectInvalid(&Auth0ResourceServerSpec{ConsentPolicy: proto.String("standard")}, "consent_policy")
		})
	})

	ginkgo.Context("token_lifetime_for_anonymous_access_tokens", func() {
		ginkgo.It("accepts thirty days", func() {
			expectValid(&Auth0ResourceServerSpec{TokenLifetimeForAnonymousAccessTokens: proto.Int32(2592000)})
		})
		ginkgo.It("rejects less than a day", func() {
			expectInvalid(&Auth0ResourceServerSpec{TokenLifetimeForAnonymousAccessTokens: proto.Int32(3600)}, "token_lifetime_for_anonymous_access_tokens")
		})
		ginkgo.It("rejects more than thirty days", func() {
			expectInvalid(&Auth0ResourceServerSpec{TokenLifetimeForAnonymousAccessTokens: proto.Int32(2592001)}, "token_lifetime_for_anonymous_access_tokens")
		})
	})

	ginkgo.Context("signing_secret", func() {
		ginkgo.It("rejects a secret shorter than 16 characters", func() {
			expectInvalid(&Auth0ResourceServerSpec{SigningAlg: "HS256", SigningSecret: proto.String("too-short")}, "signing_secret")
		})
	})

	ginkgo.Context("access_token.claims_mapping", func() {
		ginkgo.It("accepts an empty mapping, which clears the claims", func() {
			expectValid(&Auth0ResourceServerSpec{AccessToken: &Auth0ResourceServerAccessToken{ClaimsMapping: &Auth0ResourceServerClaimsMapping{}}})
		})
		ginkgo.It("rejects more than 20 claims", func() {
			claims := make([]*Auth0ResourceServerCustomClaim, 21)
			for i := range claims {
				claims[i] = &Auth0ResourceServerCustomClaim{Name: "claim", Expression: "anonymous_session.metadata.value"}
			}
			expectInvalid(&Auth0ResourceServerSpec{AccessToken: &Auth0ResourceServerAccessToken{ClaimsMapping: &Auth0ResourceServerClaimsMapping{CustomClaims: claims}}}, "custom_claims")
		})
		ginkgo.It("rejects a claim without a name", func() {
			expectInvalid(&Auth0ResourceServerSpec{AccessToken: &Auth0ResourceServerAccessToken{ClaimsMapping: &Auth0ResourceServerClaimsMapping{
				CustomClaims: []*Auth0ResourceServerCustomClaim{{Expression: "anonymous_session.metadata.country"}},
			}}}, "name")
		})
		ginkgo.It("rejects a name longer than 255 characters", func() {
			long := make([]byte, 256)
			for i := range long {
				long[i] = 'a'
			}
			expectInvalid(&Auth0ResourceServerSpec{AccessToken: &Auth0ResourceServerAccessToken{ClaimsMapping: &Auth0ResourceServerClaimsMapping{
				CustomClaims: []*Auth0ResourceServerCustomClaim{{Name: string(long), Expression: "anonymous_session.metadata.country"}},
			}}}, "name")
		})
		ginkgo.It("rejects a claim without an expression", func() {
			expectInvalid(&Auth0ResourceServerSpec{AccessToken: &Auth0ResourceServerAccessToken{ClaimsMapping: &Auth0ResourceServerClaimsMapping{
				CustomClaims: []*Auth0ResourceServerCustomClaim{{Name: "country"}},
			}}}, "expression")
		})
		ginkgo.It("rejects an expression that is not a dot path", func() {
			expectInvalid(&Auth0ResourceServerSpec{AccessToken: &Auth0ResourceServerAccessToken{ClaimsMapping: &Auth0ResourceServerClaimsMapping{
				CustomClaims: []*Auth0ResourceServerCustomClaim{{Name: "country", Expression: "country"}},
			}}}, "spec.access_token.claims_mapping.custom_claims.expression.dot_path")
		})
	})

	ginkgo.Context("authorization_details", func() {
		ginkgo.It("accepts the single disable entry", func() {
			expectValid(&Auth0ResourceServerSpec{AuthorizationDetails: []*Auth0ResourceServerAuthorizationDetail{{Disable: proto.Bool(true)}}})
		})
		ginkgo.It("rejects a disable entry beside a type", func() {
			expectInvalid(&Auth0ResourceServerSpec{AuthorizationDetails: []*Auth0ResourceServerAuthorizationDetail{
				{Type: proto.String("payment")}, {Disable: proto.Bool(true)},
			}}, "spec.authorization_details.disable_alone")
		})
		ginkgo.It("rejects an entry naming no type", func() {
			expectInvalid(&Auth0ResourceServerSpec{AuthorizationDetails: []*Auth0ResourceServerAuthorizationDetail{{}}}, "spec.authorization_details.type_or_disable")
		})
		ginkgo.It("rejects an empty type", func() {
			expectInvalid(&Auth0ResourceServerSpec{AuthorizationDetails: []*Auth0ResourceServerAuthorizationDetail{{Type: proto.String("")}}}, "spec.authorization_details.type_or_disable")
		})
	})

	ginkgo.Context("authorization_policy", func() {
		ginkgo.It("rejects a policy id longer than 1024 characters", func() {
			long := make([]byte, 1025)
			for i := range long {
				long[i] = 'p'
			}
			expectInvalid(&Auth0ResourceServerSpec{AuthorizationPolicy: &Auth0ResourceServerAuthorizationPolicy{PolicyId: proto.String(string(long))}}, "policy_id")
		})
	})

	ginkgo.Context("proof_of_possession", func() {
		ginkgo.It("accepts mTLS for every application", func() {
			expectValid(&Auth0ResourceServerSpec{ProofOfPossession: &Auth0ResourceServerProofOfPossession{
				Mechanism: proto.String("mtls"), Required: proto.Bool(true), RequiredFor: proto.String("all_clients"),
			}})
		})
		ginkgo.It("accepts disable alone", func() {
			expectValid(&Auth0ResourceServerSpec{ProofOfPossession: &Auth0ResourceServerProofOfPossession{Disable: proto.Bool(true)}})
		})
		ginkgo.It("rejects disable with a mechanism", func() {
			expectInvalid(&Auth0ResourceServerSpec{ProofOfPossession: &Auth0ResourceServerProofOfPossession{
				Disable: proto.Bool(true), Mechanism: proto.String("dpop"),
			}}, "spec.proof_of_possession.disable_alone")
		})
		ginkgo.It("rejects disable with required: true", func() {
			expectInvalid(&Auth0ResourceServerSpec{ProofOfPossession: &Auth0ResourceServerProofOfPossession{
				Disable: proto.Bool(true), Required: proto.Bool(true),
			}}, "spec.proof_of_possession.disable_alone")
		})
		ginkgo.It("rejects a mechanism without required", func() {
			expectInvalid(&Auth0ResourceServerSpec{ProofOfPossession: &Auth0ResourceServerProofOfPossession{
				Mechanism: proto.String("dpop"),
			}}, "spec.proof_of_possession.mechanism_and_required")
		})
		ginkgo.It("rejects required without a mechanism", func() {
			expectInvalid(&Auth0ResourceServerSpec{ProofOfPossession: &Auth0ResourceServerProofOfPossession{
				Required: proto.Bool(true),
			}}, "spec.proof_of_possession.mechanism_and_required")
		})
		ginkgo.It("rejects mTLS for public clients only", func() {
			expectInvalid(&Auth0ResourceServerSpec{ProofOfPossession: &Auth0ResourceServerProofOfPossession{
				Mechanism: proto.String("mtls"), Required: proto.Bool(true), RequiredFor: proto.String("public_clients"),
			}}, "spec.proof_of_possession.mtls_all_clients")
		})
		ginkgo.It("rejects a mechanism Auth0 does not offer", func() {
			expectInvalid(&Auth0ResourceServerSpec{ProofOfPossession: &Auth0ResourceServerProofOfPossession{
				Mechanism: proto.String("token_binding"), Required: proto.Bool(true),
			}}, "mechanism")
		})
		ginkgo.It("rejects an unknown required_for", func() {
			expectInvalid(&Auth0ResourceServerSpec{ProofOfPossession: &Auth0ResourceServerProofOfPossession{
				Mechanism: proto.String("dpop"), Required: proto.Bool(true), RequiredFor: proto.String("confidential_clients"),
			}}, "required_for")
		})
	})

	ginkgo.Context("subject_type_authorization", func() {
		ginkgo.It("accepts a single policy", func() {
			expectValid(&Auth0ResourceServerSpec{SubjectTypeAuthorization: &Auth0ResourceServerSubjectTypeAuthorization{
				Client: &Auth0ResourceServerClientAuthorization{Policy: proto.String("require_client_grant")},
			}})
		})
		ginkgo.It("accepts allow_all for users", func() {
			expectValid(&Auth0ResourceServerSpec{SubjectTypeAuthorization: &Auth0ResourceServerSubjectTypeAuthorization{
				User: &Auth0ResourceServerUserAuthorization{Policy: proto.String("allow_all")},
			}})
		})
		ginkgo.It("rejects an unknown user policy", func() {
			expectInvalid(&Auth0ResourceServerSpec{SubjectTypeAuthorization: &Auth0ResourceServerSubjectTypeAuthorization{
				User: &Auth0ResourceServerUserAuthorization{Policy: proto.String("allow_first_party")},
			}}, "policy")
		})
		ginkgo.It("rejects allow_all for clients", func() {
			expectInvalid(&Auth0ResourceServerSpec{SubjectTypeAuthorization: &Auth0ResourceServerSubjectTypeAuthorization{
				Client: &Auth0ResourceServerClientAuthorization{Policy: proto.String("allow_all")},
			}}, "policy")
		})
		ginkgo.It("rejects allow_all for anonymous users", func() {
			expectInvalid(&Auth0ResourceServerSpec{SubjectTypeAuthorization: &Auth0ResourceServerSubjectTypeAuthorization{
				AnonymousUser: &Auth0ResourceServerAnonymousUserAuthorization{Policy: proto.String("allow_all")},
			}}, "policy")
		})
	})

	ginkgo.Context("token_encryption", func() {
		ginkgo.It("accepts disable alone", func() {
			expectValid(&Auth0ResourceServerSpec{TokenEncryption: &Auth0ResourceServerTokenEncryption{Disable: proto.Bool(true)}})
		})
		ginkgo.It("rejects disable with a key", func() {
			expectInvalid(&Auth0ResourceServerSpec{TokenEncryption: &Auth0ResourceServerTokenEncryption{
				Disable: proto.Bool(true), EncryptionKey: validKey(),
			}}, "spec.token_encryption.disable_alone")
		})
		ginkgo.It("rejects disable with a format", func() {
			expectInvalid(&Auth0ResourceServerSpec{TokenEncryption: &Auth0ResourceServerTokenEncryption{
				Disable: proto.Bool(true), Format: proto.String("compact-nested-jwe"),
			}}, "spec.token_encryption.disable_alone")
		})
		ginkgo.It("rejects a format without a key", func() {
			expectInvalid(&Auth0ResourceServerSpec{TokenEncryption: &Auth0ResourceServerTokenEncryption{
				Format: proto.String("compact-nested-jwe"),
			}}, "spec.token_encryption.format_with_key")
		})
		ginkgo.It("rejects a key without a format", func() {
			expectInvalid(&Auth0ResourceServerSpec{TokenEncryption: &Auth0ResourceServerTokenEncryption{
				EncryptionKey: validKey(),
			}}, "spec.token_encryption.format_with_key")
		})
		ginkgo.It("rejects a format Auth0 does not offer", func() {
			expectInvalid(&Auth0ResourceServerSpec{TokenEncryption: &Auth0ResourceServerTokenEncryption{
				Format: proto.String("compact-jwe"), EncryptionKey: validKey(),
			}}, "format")
		})
		ginkgo.It("rejects a key without an algorithm", func() {
			key := validKey()
			key.Algorithm = ""
			expectInvalid(&Auth0ResourceServerSpec{TokenEncryption: &Auth0ResourceServerTokenEncryption{
				Format: proto.String("compact-nested-jwe"), EncryptionKey: key,
			}}, "algorithm")
		})
		ginkgo.It("rejects an algorithm Auth0 does not offer", func() {
			key := validKey()
			key.Algorithm = "RSA1_5"
			expectInvalid(&Auth0ResourceServerSpec{TokenEncryption: &Auth0ResourceServerTokenEncryption{
				Format: proto.String("compact-nested-jwe"), EncryptionKey: key,
			}}, "algorithm")
		})
		ginkgo.It("rejects a key without a PEM", func() {
			key := validKey()
			key.Pem = ""
			expectInvalid(&Auth0ResourceServerSpec{TokenEncryption: &Auth0ResourceServerTokenEncryption{
				Format: proto.String("compact-nested-jwe"), EncryptionKey: key,
			}}, "pem")
		})
		ginkgo.It("rejects a kid longer than 128 characters", func() {
			key := validKey()
			long := make([]byte, 129)
			for i := range long {
				long[i] = 'k'
			}
			key.Kid = proto.String(string(long))
			expectInvalid(&Auth0ResourceServerSpec{TokenEncryption: &Auth0ResourceServerTokenEncryption{
				Format: proto.String("compact-nested-jwe"), EncryptionKey: key,
			}}, "kid")
		})
		ginkgo.It("rejects a key name longer than 128 characters", func() {
			key := validKey()
			long := make([]byte, 129)
			for i := range long {
				long[i] = 'n'
			}
			key.Name = proto.String(string(long))
			expectInvalid(&Auth0ResourceServerSpec{TokenEncryption: &Auth0ResourceServerTokenEncryption{
				Format: proto.String("compact-nested-jwe"), EncryptionKey: key,
			}}, "name")
		})
	})

	ginkgo.Context("third_party_client_default_grants", func() {
		ginkgo.It("rejects two grants for one subject type", func() {
			expectInvalid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{SubjectType: "user", Scopes: []string{"read:items"}},
				{SubjectType: "user", Scopes: []string{"write:items"}},
			}}, "spec.third_party_client_default_grants.one_per_subject_type")
		})
		ginkgo.It("rejects a grant without a subject type", func() {
			expectInvalid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{Scopes: []string{"read:items"}},
			}}, "subject_type")
		})
		ginkgo.It("rejects the anonymous_user subject type", func() {
			expectInvalid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{SubjectType: "anonymous_user", Scopes: []string{"read:items"}},
			}}, "subject_type")
		})
		ginkgo.It("rejects a grant with neither scopes nor allow_all_scopes", func() {
			expectInvalid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{SubjectType: "user"},
			}}, "spec.third_party_client_default_grants.scopes_or_all")
		})
		ginkgo.It("rejects scopes beside allow_all_scopes", func() {
			expectInvalid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{SubjectType: "user", AllowAllScopes: proto.Bool(true), Scopes: []string{"read:items"}},
			}}, "spec.third_party_client_default_grants.scopes_or_all")
		})
		ginkgo.It("rejects an empty scope", func() {
			expectInvalid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{SubjectType: "user", Scopes: []string{""}},
			}}, "scopes")
		})
		ginkgo.It("rejects a scope longer than 280 characters", func() {
			long := make([]byte, 281)
			for i := range long {
				long[i] = 's'
			}
			expectInvalid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{SubjectType: "user", Scopes: []string{string(long)}},
			}}, "scopes")
		})
		ginkgo.It("rejects authorization_details_types on a client grant", func() {
			expectInvalid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{SubjectType: "client", Scopes: []string{"read:items"}, AuthorizationDetailsTypes: []string{"payment"}},
			}}, "spec.third_party_client_default_grants.authorization_details_types_user_only")
		})
		ginkgo.It("rejects an empty authorization_details type", func() {
			expectInvalid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{SubjectType: "user", Scopes: []string{"read:items"}, AuthorizationDetailsTypes: []string{""}},
			}}, "authorization_details_types")
		})
		ginkgo.It("rejects an authorization_details type longer than 255 characters", func() {
			long := make([]byte, 256)
			for i := range long {
				long[i] = 't'
			}
			expectInvalid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{SubjectType: "user", Scopes: []string{"read:items"}, AuthorizationDetailsTypes: []string{string(long)}},
			}}, "authorization_details_types")
		})
		ginkgo.It("rejects an unknown organization_usage", func() {
			expectInvalid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{SubjectType: "client", Scopes: []string{"read:items"}, OrganizationUsage: proto.String("optional")},
			}}, "organization_usage")
		})
		ginkgo.It("accepts allow_any_organization false", func() {
			expectValid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{SubjectType: "client", Scopes: []string{"read:items"}, AllowAnyOrganization: proto.Bool(false)},
			}})
		})
		ginkgo.It("rejects allow_any_organization true", func() {
			expectInvalid(&Auth0ResourceServerSpec{ThirdPartyClientDefaultGrants: []*Auth0ResourceServerThirdPartyClientDefaultGrant{
				{SubjectType: "client", Scopes: []string{"read:items"}, AllowAnyOrganization: proto.Bool(true)},
			}}, "spec.third_party_client_default_grants.no_allow_any_organization")
		})
	})
})
