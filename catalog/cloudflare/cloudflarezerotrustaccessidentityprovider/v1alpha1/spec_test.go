package cloudflarezerotrustaccessidentityproviderv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareZeroTrustAccessIdentityProviderSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustAccessIdentityProviderSpec Custom Validation Tests")
}

const testAccountId = "023e105f4ecef8ad9ca31a8372d0c353"

func boolPtr(b bool) *bool { return &b }

func literalRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validIdp(spec *CloudflareZeroTrustAccessIdentityProviderSpec) *CloudflareZeroTrustAccessIdentityProvider {
	return &CloudflareZeroTrustAccessIdentityProvider{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustAccessIdentityProvider",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-idp",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareZeroTrustAccessIdentityProviderSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a one-time PIN provider with no config", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "email-pin",
				Type:      "onetimepin",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a GitHub OAuth provider", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "github-login",
				Type:      "github",
				Config: &CloudflareZeroTrustAccessIdentityProviderConfig{
					ClientId:     "example-client-id",
					ClientSecret: literalRef("example-client-secret"),
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an azureAD provider with its gated fields", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "corp-entra",
				Type:      "azureAD",
				Config: &CloudflareZeroTrustAccessIdentityProviderConfig{
					ClientId:                 "example-client-id",
					ClientSecret:             literalRef("example-client-secret"),
					DirectoryId:              "6a7e50b8-8e0c-4d0a-9d1c-000000000000",
					Prompt:                   "select_account",
					SupportGroups:            boolPtr(true),
					ConditionalAccessEnabled: boolPtr(true),
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a SAML provider with encryption and its certificate set", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "corp-saml",
				Type:      "saml",
				Config: &CloudflareZeroTrustAccessIdentityProviderConfig{
					IssuerUrl:        "https://idp.example.com/entity",
					SsoTargetUrl:     "https://idp.example.com/sso",
					IdpPublicCerts:   []string{"-----BEGIN CERTIFICATE-----..."},
					EnableEncryption: boolPtr(true),
				},
				SamlCertificateSetId: "8e0c6a7e-50b8-4d0a-9d1c-000000000000",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept SCIM config with seat and user deprovision together", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "corp-okta",
				Type:      "okta",
				Config: &CloudflareZeroTrustAccessIdentityProviderConfig{
					ClientId:     "example-client-id",
					ClientSecret: literalRef("example-client-secret"),
					OktaAccount:  "https://example.okta.com",
				},
				ScimConfig: &CloudflareZeroTrustAccessIdentityProviderScimConfig{
					Enabled:                true,
					IdentityUpdateBehavior: "automatic",
					SeatDeprovision:        true,
					UserDeprovision:        true,
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a provider with no scope", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				Name: "github-login",
				Type: "github",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a provider with both scopes", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				ZoneId:    literalRef("023e105f4ecef8ad9ca31a8372d0c353"),
				Name:      "github-login",
				Type:      "github",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown provider type", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "mystery",
				Type:      "ldap",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject azureAD fields on a non-azureAD provider", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "github-login",
				Type:      "github",
				Config: &CloudflareZeroTrustAccessIdentityProviderConfig{
					ClientId:     "example-client-id",
					ClientSecret: literalRef("example-client-secret"),
					DirectoryId:  "6a7e50b8-8e0c-4d0a-9d1c-000000000000",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject SAML fields on an oidc provider", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "generic-oidc",
				Type:      "oidc",
				Config: &CloudflareZeroTrustAccessIdentityProviderConfig{
					AuthUrl:      "https://idp.example.com/authorize",
					TokenUrl:     "https://idp.example.com/token",
					CertsUrl:     "https://idp.example.com/jwks",
					SsoTargetUrl: "https://idp.example.com/sso",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject okta fields on a onelogin provider", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "corp-onelogin",
				Type:      "onelogin",
				Config: &CloudflareZeroTrustAccessIdentityProviderConfig{
					OneloginAccount: "https://example.onelogin.com",
					OktaAccount:     "https://example.okta.com",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject encryption without a certificate set", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "corp-saml",
				Type:      "saml",
				Config: &CloudflareZeroTrustAccessIdentityProviderConfig{
					IssuerUrl:        "https://idp.example.com/entity",
					EnableEncryption: boolPtr(true),
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject SCIM on a one-time PIN provider", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "email-pin",
				Type:      "onetimepin",
				ScimConfig: &CloudflareZeroTrustAccessIdentityProviderScimConfig{
					Enabled: true,
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject seat_deprovision without user_deprovision", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "corp-okta",
				Type:      "okta",
				ScimConfig: &CloudflareZeroTrustAccessIdentityProviderScimConfig{
					Enabled:         true,
					SeatDeprovision: true,
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid prompt value", func() {
			input := validIdp(&CloudflareZeroTrustAccessIdentityProviderSpec{
				AccountId: testAccountId,
				Name:      "corp-entra",
				Type:      "azureAD",
				Config: &CloudflareZeroTrustAccessIdentityProviderConfig{
					Prompt: "always",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
