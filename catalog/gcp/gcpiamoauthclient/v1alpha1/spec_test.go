package gcpiamoauthclientv1alpha1

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
	ginkgo.RunSpecs(t, "GcpIamOauthClientSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpIamOauthClientSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpIamOauthClient {
		return &GcpIamOauthClient{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpIamOauthClient",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-oauth-client",
			},
			Spec: &GcpIamOauthClientSpec{
				AllowedGrantTypes:   []string{"AUTHORIZATION_CODE_GRANT"},
				AllowedScopes:       []string{"https://www.googleapis.com/auth/cloud-platform"},
				AllowedRedirectUris: []*foreignkeyv1.StringValueOrRef{litRef("https://app.example.com/callback")},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal client", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept both grant types together", func() {
		target := minimal()
		target.Spec.AllowedGrantTypes = []string{"AUTHORIZATION_CODE_GRANT", "REFRESH_TOKEN_GRANT"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept CONFIDENTIAL_CLIENT", func() {
		target := minimal()
		target.Spec.ClientType = "CONFIDENTIAL_CLIENT"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject PUBLIC_CLIENT — GCP refuses to create one (live API truth)", func() {
		target := minimal()
		target.Spec.ClientType = "PUBLIC_CLIENT"
		gomega.Expect(validator.Validate(target)).NotTo(gomega.Succeed())
	})

	ginkgo.It("should accept a confidential client with credentials", func() {
		target := minimal()
		target.Spec.ClientType = "CONFIDENTIAL_CLIENT"
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		target.Spec.OauthClientId = "my-app-client"
		target.Spec.DisplayName = "My app"
		target.Spec.Description = "Server-side web app"
		target.Spec.Credentials = []*GcpIamOauthClientCredential{{
			CredentialId: "primary",
			DisplayName:  "Primary secret",
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

	ginkgo.It("should reject empty allowed_grant_types", func() {
		target := minimal()
		target.Spec.AllowedGrantTypes = nil
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown grant type", func() {
		target := minimal()
		target.Spec.AllowedGrantTypes = []string{"IMPLICIT_GRANT"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "REFRESH_TOKEN_GRANT")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject empty allowed_scopes", func() {
		target := minimal()
		target.Spec.AllowedScopes = nil
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject empty allowed_redirect_uris", func() {
		target := minimal()
		target.Spec.AllowedRedirectUris = nil
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid client_type", func() {
		target := minimal()
		target.Spec.ClientType = "SECRET_CLIENT"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "CONFIDENTIAL_CLIENT")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a credential without credential_id", func() {
		target := minimal()
		target.Spec.Credentials = []*GcpIamOauthClientCredential{{
			DisplayName: "Unnamed",
		}}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "ABANDON")).To(gomega.BeTrue())
	})
})
