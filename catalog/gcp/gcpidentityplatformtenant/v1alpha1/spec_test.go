package gcpidentityplatformtenantv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpIdentityPlatformTenantSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpIdentityPlatformTenantSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpIdentityPlatformTenant {
		return &GcpIdentityPlatformTenant{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpIdentityPlatformTenant",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-tenant",
			},
			Spec: &GcpIdentityPlatformTenantSpec{
				DisplayName: "acme-corp",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal tenant", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal and all sign-in switches", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		target.Spec.AllowPasswordSignup = true
		target.Spec.EnableEmailLinkSignin = true
		target.Spec.DisableAuth = false
		target.Spec.ClientPermissions = &GcpIdentityPlatformTenantClientPermissions{
			DisabledUserSignup:   true,
			DisabledUserDeletion: true,
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a default supported IdP with credentials", func() {
		target := minimal()
		target.Spec.DefaultSupportedIdps = []*GcpIdentityPlatformTenantDefaultSupportedIdp{{
			IdpId:        "google.com",
			ClientId:     "1234-abc.apps.googleusercontent.com",
			ClientSecret: "console-issued-secret",
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an OIDC provider with tenant-required display_name", func() {
		target := minimal()
		target.Spec.OauthIdpConfigs = []*GcpIdentityPlatformTenantOauthIdp{{
			Name:        "oidc.corp-sso",
			DisplayName: "Corp SSO",
			Issuer:      "https://accounts.example.com",
			ClientId:    "corp-client",
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a SAML provider with the tenant-required sp_config", func() {
		target := minimal()
		target.Spec.InboundSamlConfigs = []*GcpIdentityPlatformTenantInboundSaml{{
			Name:        "saml.okta-prod",
			DisplayName: "Okta",
			IdpConfig: &GcpIdentityPlatformTenantSamlIdpConfig{
				IdpEntityId: "http://www.okta.com/exk123",
				SsoUrl:      "https://corp.okta.com/app/sso/saml",
			},
			SpConfig: &GcpIdentityPlatformTenantSamlSpConfig{
				CallbackUri: "https://my-project.firebaseapp.com/__/auth/handler",
				SpEntityId:  "my-project-sp",
			},
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing display_name", func() {
		target := minimal()
		target.Spec.DisplayName = ""
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown default-supported IdP", func() {
		target := minimal()
		target.Spec.DefaultSupportedIdps = []*GcpIdentityPlatformTenantDefaultSupportedIdp{{
			IdpId:        "myidp.example.com",
			ClientId:     "c",
			ClientSecret: "s",
		}}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an OIDC provider without display_name (tenant-level rule)", func() {
		target := minimal()
		target.Spec.OauthIdpConfigs = []*GcpIdentityPlatformTenantOauthIdp{{
			Name:     "oidc.corp-sso",
			Issuer:   "https://accounts.example.com",
			ClientId: "corp-client",
		}}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an OIDC name without the oidc. prefix", func() {
		target := minimal()
		target.Spec.OauthIdpConfigs = []*GcpIdentityPlatformTenantOauthIdp{{
			Name:        "corp-sso",
			DisplayName: "Corp SSO",
			Issuer:      "https://accounts.example.com",
			ClientId:    "corp-client",
		}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "oidc.")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a SAML provider without sp_config (tenant-level rule)", func() {
		target := minimal()
		target.Spec.InboundSamlConfigs = []*GcpIdentityPlatformTenantInboundSaml{{
			Name:        "saml.okta-prod",
			DisplayName: "Okta",
			IdpConfig: &GcpIdentityPlatformTenantSamlIdpConfig{
				IdpEntityId: "e",
				SsoUrl:      "https://sso.example.com",
			},
		}}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a non-https SAML callback_uri", func() {
		target := minimal()
		target.Spec.InboundSamlConfigs = []*GcpIdentityPlatformTenantInboundSaml{{
			Name:        "saml.okta-prod",
			DisplayName: "Okta",
			IdpConfig: &GcpIdentityPlatformTenantSamlIdpConfig{
				IdpEntityId: "e",
				SsoUrl:      "https://sso.example.com",
			},
			SpConfig: &GcpIdentityPlatformTenantSamlSpConfig{
				CallbackUri: "http://insecure.example.com/handler",
				SpEntityId:  "sp",
			},
		}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "https://")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "ABANDON")).To(gomega.BeTrue())
	})
})
