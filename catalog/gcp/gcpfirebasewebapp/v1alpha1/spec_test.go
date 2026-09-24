package gcpfirebasewebappv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpFirebaseWebAppSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpFirebaseWebAppSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpFirebaseWebApp {
		return &GcpFirebaseWebApp{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpFirebaseWebApp",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-firebase-web-app",
			},
			Spec: &GcpFirebaseWebAppSpec{
				DisplayName: "My Web App",
			},
		}
	}

	expectError := func(target *GcpFirebaseWebApp, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept the minimal registration (a display name)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a Firebase project reference, an API key reference, and every deletion_policy", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.ProjectId = litRef("my-gcp-project-123")
			target.Spec.ApiKeyId = litRef("9f3a2c1e-4b5d-4e6f-8a7b-0c1d2e3f4a5b")
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	ginkgo.It("should accept reCAPTCHA v3, reCAPTCHA Enterprise, both together (the migration shape), and debug tokens", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseWebAppAppCheck{
			RecaptchaV3: &GcpFirebaseWebAppAppCheckRecaptchaV3{SiteSecret: "6Lc_secret_value", TokenTtl: "3600s"},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.AppCheck.RecaptchaEnterprise = &GcpFirebaseWebAppAppCheckRecaptchaEnterprise{SiteKey: "6Lc_site_key_value", TokenTtl: "1800s"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target.Spec.AppCheck.RecaptchaV3 = nil
		target.Spec.AppCheck.DebugTokens = []*GcpFirebaseWebAppAppCheckDebugToken{
			{DisplayName: "localhost", Token: "3b6c0e2a-8f4d-4c1b-9a7e-2d5f8b1c4e6a"},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing display name", func() {
		target := minimal()
		target.Spec.DisplayName = ""
		expectError(target, "display_name")
	})

	ginkgo.It("should reject a reCAPTCHA v3 block without its site secret and an Enterprise block without its site key", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseWebAppAppCheck{
			RecaptchaV3: &GcpFirebaseWebAppAppCheckRecaptchaV3{TokenTtl: "3600s"},
		}
		expectError(target, "site_secret")

		target.Spec.AppCheck = &GcpFirebaseWebAppAppCheck{
			RecaptchaEnterprise: &GcpFirebaseWebAppAppCheckRecaptchaEnterprise{TokenTtl: "3600s"},
		}
		expectError(target, "site_key")
	})

	ginkgo.It("should reject a malformed token_ttl on either provider", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseWebAppAppCheck{
			RecaptchaV3: &GcpFirebaseWebAppAppCheckRecaptchaV3{SiteSecret: "s", TokenTtl: "3600"},
		}
		expectError(target, "token_ttl must be a duration")

		target.Spec.AppCheck = &GcpFirebaseWebAppAppCheck{
			RecaptchaEnterprise: &GcpFirebaseWebAppAppCheckRecaptchaEnterprise{SiteKey: "k", TokenTtl: "PT1H"},
		}
		expectError(target, "token_ttl must be a duration")
	})

	ginkgo.It("should reject a debug token without a name or a value, and duplicated names", func() {
		target := minimal()
		target.Spec.AppCheck = &GcpFirebaseWebAppAppCheck{
			DebugTokens: []*GcpFirebaseWebAppAppCheckDebugToken{{Token: "3b6c0e2a-8f4d-4c1b-9a7e-2d5f8b1c4e6a"}},
		}
		expectError(target, "display_name")

		target.Spec.AppCheck.DebugTokens = []*GcpFirebaseWebAppAppCheckDebugToken{{DisplayName: "localhost"}}
		expectError(target, "token")

		target.Spec.AppCheck.DebugTokens = []*GcpFirebaseWebAppAppCheckDebugToken{
			{DisplayName: "localhost", Token: "3b6c0e2a-8f4d-4c1b-9a7e-2d5f8b1c4e6a"},
			{DisplayName: "localhost", Token: "7d2e9f1b-4a6c-4e8d-b0f3-1c5a7e9b2d4f"},
		}
		expectError(target, "at most once")
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "DETACH"
		expectError(target, "deletion_policy must be one of")
	})

	ginkgo.It("should reject a wrong kind or apiVersion", func() {
		target := minimal()
		target.Kind = "GcpFirebaseWebsite"
		expectError(target, "kind")

		target = minimal()
		target.ApiVersion = "gcp.planton.dev/v1"
		expectError(target, "api_version")
	})
})
