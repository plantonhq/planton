package awscognitoidentityproviderv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	fkv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsCognitoIdentityProviderSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsCognitoIdentityProviderSpec Validation Suite")
}

var _ = ginkgo.Describe("AwsCognitoIdentityProviderSpec validations", func() {

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal Google provider", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "google-idp"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "Google",
				ProviderType: AwsCognitoIdentityProviderType_Google,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientId:        "123456789.apps.googleusercontent.com",
						ClientSecret:    "GOCSPX-secret",
						AuthorizeScopes: "email profile openid",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a Facebook provider with api_version", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "facebook-idp"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "Facebook",
				ProviderType: AwsCognitoIdentityProviderType_Facebook,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Facebook{
					Facebook: &AwsCognitoIdpFacebookConfig{
						ClientId:        "1234567890",
						ClientSecret:    "fb-secret",
						AuthorizeScopes: "email,public_profile",
						ApiVersion:      "v17.0",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a Login with Amazon provider", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "amazon-idp"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "LoginWithAmazon",
				ProviderType: AwsCognitoIdentityProviderType_LoginWithAmazon,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_LoginWithAmazon{
					LoginWithAmazon: &AwsCognitoIdpLoginWithAmazonConfig{
						ClientId:        "amzn1.application-oa2-client.example",
						ClientSecret:    "amazon-secret",
						AuthorizeScopes: "profile postal_code",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a Sign in with Apple provider", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "apple-idp"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "SignInWithApple",
				ProviderType: AwsCognitoIdentityProviderType_SignInWithApple,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_SignInWithApple{
					SignInWithApple: &AwsCognitoIdpSignInWithAppleConfig{
						ClientId:        "com.example.app",
						TeamId:          "ABCDE12345",
						KeyId:           "KEY123456",
						PrivateKey:      "-----BEGIN PRIVATE KEY-----\nMIGTAgEA...\n-----END PRIVATE KEY-----",
						AuthorizeScopes: "email name",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a minimal OIDC provider", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "okta-oidc"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "CorpOkta",
				ProviderType: AwsCognitoIdentityProviderType_OIDC,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Oidc{
					Oidc: &AwsCognitoIdpOidcConfig{
						ClientId:   "0oa1b2c3d4e5f6g7h8i9",
						OidcIssuer: "https://dev-123456.okta.com",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a full OIDC provider with all optional URL overrides", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "azure-oidc"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "AzureAD",
				ProviderType: AwsCognitoIdentityProviderType_OIDC,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Oidc{
					Oidc: &AwsCognitoIdpOidcConfig{
						ClientId:                "app-client-id",
						OidcIssuer:              "https://login.microsoftonline.com/tenant-id/v2.0",
						AuthorizeScopes:         "openid email profile",
						ClientSecret:            "oidc-client-secret",
						AttributesRequestMethod: "POST",
						AuthorizeUrl:            "https://login.microsoftonline.com/tenant-id/oauth2/v2.0/authorize",
						TokenUrl:                "https://login.microsoftonline.com/tenant-id/oauth2/v2.0/token",
						AttributesUrl:           "https://graph.microsoft.com/oidc/userinfo",
						JwksUri:                 "https://login.microsoftonline.com/tenant-id/discovery/v2.0/keys",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a SAML provider with metadata_file", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "saml-idp"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "AzureAD-SAML",
				ProviderType: AwsCognitoIdentityProviderType_SAML,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Saml{
					Saml: &AwsCognitoIdpSamlConfig{
						MetadataFile: "<EntityDescriptor>...</EntityDescriptor>",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a SAML provider with metadata_url and all options", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "saml-full"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "ADFS-Corp",
				ProviderType: AwsCognitoIdentityProviderType_SAML,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Saml{
					Saml: &AwsCognitoIdpSamlConfig{
						MetadataUrl:             "https://adfs.corp.example.com/FederationMetadata/2007-06/FederationMetadata.xml",
						IdpSignOut:              true,
						IdpInit:                 true,
						EncryptedResponses:      true,
						RequestSigningAlgorithm: "rsa-sha256",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a provider with attribute_mapping", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "google-mapped"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "Google",
				ProviderType: AwsCognitoIdentityProviderType_Google,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientId:        "123456789.apps.googleusercontent.com",
						ClientSecret:    "GOCSPX-secret",
						AuthorizeScopes: "email profile openid",
					},
				},
				AttributeMapping: map[string]string{
					"email":    "email",
					"username": "sub",
					"name":     "name",
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a provider with idp_identifiers", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "oidc-with-ids"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "CorpSSO",
				ProviderType: AwsCognitoIdentityProviderType_OIDC,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Oidc{
					Oidc: &AwsCognitoIdpOidcConfig{
						ClientId:   "client-id",
						OidcIssuer: "https://sso.corp.example.com",
					},
				},
				IdpIdentifiers: []string{"corp-sso", "enterprise-login"},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a provider with valueFrom for user_pool_id", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "google-ref"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region: "us-west-2",
				UserPoolId: &fkv1.StringValueOrRef{
					LiteralOrRef: &fkv1.StringValueOrRef_ValueFrom{
						ValueFrom: &fkv1.ValueFromRef{
							Kind:      cloudresourcekind.CloudResourceKind_AwsCognitoUserPool,
							Name:      "my-pool",
							FieldPath: "status.outputs.user_pool_id",
						},
					},
				},
				ProviderName: "Google",
				ProviderType: AwsCognitoIdentityProviderType_Google,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientId:        "123456789.apps.googleusercontent.com",
						ClientSecret:    "GOCSPX-secret",
						AuthorizeScopes: "email profile openid",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: missing required fields
	// -------------------------------------------------------------------------

	ginkgo.It("fails when user_pool_id is missing", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "no-pool"},
			Spec: &AwsCognitoIdentityProviderSpec{
				ProviderName: "Google",
				ProviderType: AwsCognitoIdentityProviderType_Google,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientId:        "client-id",
						ClientSecret:    "client-secret",
						AuthorizeScopes: "email",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when provider_name is empty", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "no-name"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "",
				ProviderType: AwsCognitoIdentityProviderType_Google,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientId:        "client-id",
						ClientSecret:    "client-secret",
						AuthorizeScopes: "email",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when provider_type is unspecified", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "no-type"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "Google",
				ProviderType: AwsCognitoIdentityProviderType_unspecified,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientId:        "client-id",
						ClientSecret:    "client-secret",
						AuthorizeScopes: "email",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when no provider_config oneof branch is set", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "no-config"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "Google",
				ProviderType: AwsCognitoIdentityProviderType_Google,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: provider_type / config mismatch
	// -------------------------------------------------------------------------

	ginkgo.It("fails when provider_type is Google but SAML config is set", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "mismatch-google-saml"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "Google",
				ProviderType: AwsCognitoIdentityProviderType_Google,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Saml{
					Saml: &AwsCognitoIdpSamlConfig{
						MetadataUrl: "https://example.com/metadata",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when provider_type is OIDC but Google config is set", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "mismatch-oidc-google"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "CorpSSO",
				ProviderType: AwsCognitoIdentityProviderType_OIDC,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientId:        "client-id",
						ClientSecret:    "client-secret",
						AuthorizeScopes: "email",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when provider_type is SAML but Facebook config is set", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "mismatch-saml-fb"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "ADFS",
				ProviderType: AwsCognitoIdentityProviderType_SAML,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Facebook{
					Facebook: &AwsCognitoIdpFacebookConfig{
						ClientId:        "fb-app-id",
						ClientSecret:    "fb-secret",
						AuthorizeScopes: "email",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: provider_name constraints
	// -------------------------------------------------------------------------

	ginkgo.It("fails when provider_name exceeds 32 characters", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "long-name"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "ThisProviderNameExceedsThirtyTwoCharactersLimit",
				ProviderType: AwsCognitoIdentityProviderType_Google,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientId:        "client-id",
						ClientSecret:    "client-secret",
						AuthorizeScopes: "email",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: SAML metadata mutual exclusion
	// -------------------------------------------------------------------------

	ginkgo.It("fails when SAML has both metadata_file and metadata_url", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "saml-both"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "ADFS",
				ProviderType: AwsCognitoIdentityProviderType_SAML,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Saml{
					Saml: &AwsCognitoIdpSamlConfig{
						MetadataFile: "<EntityDescriptor>...</EntityDescriptor>",
						MetadataUrl:  "https://example.com/metadata",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when SAML has neither metadata_file nor metadata_url", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "saml-none"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "ADFS",
				ProviderType: AwsCognitoIdentityProviderType_SAML,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Saml{
					Saml: &AwsCognitoIdpSamlConfig{},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: OIDC attributes_request_method
	// -------------------------------------------------------------------------

	ginkgo.It("fails when OIDC attributes_request_method is invalid", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "oidc-bad-method"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "BadOIDC",
				ProviderType: AwsCognitoIdentityProviderType_OIDC,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Oidc{
					Oidc: &AwsCognitoIdpOidcConfig{
						ClientId:                "client-id",
						OidcIssuer:              "https://example.com",
						AttributesRequestMethod: "PUT",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: nested message required fields
	// -------------------------------------------------------------------------

	ginkgo.It("fails when Google config is missing client_id", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "google-no-client"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "Google",
				ProviderType: AwsCognitoIdentityProviderType_Google,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientSecret:    "client-secret",
						AuthorizeScopes: "email",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when OIDC config is missing oidc_issuer", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "oidc-no-issuer"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "BadOIDC",
				ProviderType: AwsCognitoIdentityProviderType_OIDC,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Oidc{
					Oidc: &AwsCognitoIdpOidcConfig{
						ClientId: "client-id",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when SignInWithApple config is missing team_id", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "apple-no-team"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "Apple",
				ProviderType: AwsCognitoIdentityProviderType_SignInWithApple,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_SignInWithApple{
					SignInWithApple: &AwsCognitoIdpSignInWithAppleConfig{
						ClientId:        "com.example.app",
						KeyId:           "KEY123456",
						PrivateKey:      "-----BEGIN PRIVATE KEY-----\nkey\n-----END PRIVATE KEY-----",
						AuthorizeScopes: "email name",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: idp_identifiers constraints
	// -------------------------------------------------------------------------

	ginkgo.It("fails when idp_identifiers exceeds 50 items", func() {
		identifiers := make([]string, 51)
		for i := range identifiers {
			identifiers[i] = "id"
		}
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "too-many-ids"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "us-east-1_Ab1Cd2EfG"}},
				ProviderName: "Google",
				ProviderType: AwsCognitoIdentityProviderType_Google,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientId:        "client-id",
						ClientSecret:    "client-secret",
						AuthorizeScopes: "email",
					},
				},
				IdpIdentifiers: identifiers,
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: api.proto constants
	// -------------------------------------------------------------------------

	ginkgo.It("fails when api_version is wrong", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "wrong.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "pool-id"}},
				ProviderName: "Google",
				ProviderType: AwsCognitoIdentityProviderType_Google,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientId: "id", ClientSecret: "secret", AuthorizeScopes: "email",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when kind is wrong", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "WrongKind",
			Metadata:   &shared.CloudResourceMetadata{Name: "test"},
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "pool-id"}},
				ProviderName: "Google",
				ProviderType: AwsCognitoIdentityProviderType_Google,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientId: "id", ClientSecret: "secret", AuthorizeScopes: "email",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when metadata is missing", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Spec: &AwsCognitoIdentityProviderSpec{
				Region:       "us-west-2",
				UserPoolId:   &fkv1.StringValueOrRef{LiteralOrRef: &fkv1.StringValueOrRef_Value{Value: "pool-id"}},
				ProviderName: "Google",
				ProviderType: AwsCognitoIdentityProviderType_Google,
				ProviderConfig: &AwsCognitoIdentityProviderSpec_Google{
					Google: &AwsCognitoIdpGoogleConfig{
						ClientId: "id", ClientSecret: "secret", AuthorizeScopes: "email",
					},
				},
			},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when spec is missing", func() {
		input := &AwsCognitoIdentityProvider{
			ApiVersion: "aws.planton.dev/v1",
			Kind:       "AwsCognitoIdentityProvider",
			Metadata:   &shared.CloudResourceMetadata{Name: "test"},
		}
		err := protovalidate.Validate(input)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
