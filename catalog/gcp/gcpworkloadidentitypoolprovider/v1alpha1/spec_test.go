package gcpworkloadidentitypoolproviderv1alpha1

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
	ginkgo.RunSpecs(t, "GcpWorkloadIdentityPoolProviderSpec Suite")
}

var _ = ginkgo.Describe("GcpWorkloadIdentityPoolProviderSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid OIDC provider (the most common shape).
	minimal := func() *GcpWorkloadIdentityPoolProvider {
		return &GcpWorkloadIdentityPoolProvider{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpWorkloadIdentityPoolProvider",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-provider",
			},
			Spec: &GcpWorkloadIdentityPoolProviderSpec{
				WorkloadIdentityPoolId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "github-actions"},
				},
				WorkloadIdentityPoolProviderId: "github-oidc",
				AttributeMapping: map[string]string{
					"google.subject": "assertion.sub",
				},
				Issuer: &GcpWorkloadIdentityPoolProviderSpec_Oidc{
					Oidc: &GcpWorkloadIdentityPoolProviderOidc{
						IssuerUri: "https://token.actions.githubusercontent.com",
					},
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid OIDC provider", func() {
		err := validator.Validate(minimal())
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully populated OIDC provider", func() {
		msg := minimal()
		msg.Spec.ProjectId = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "my-gcp-project"},
		}
		msg.Spec.DisplayName = "GitHub Actions OIDC"
		msg.Spec.Description = "Trusts GitHub-minted tokens for the engineering org"
		msg.Spec.Disabled = true
		msg.Spec.AttributeMapping["attribute.repository"] = "assertion.repository"
		msg.Spec.AttributeCondition = `assertion.repository_owner == "my-org"`
		msg.Spec.GetOidc().AllowedAudiences = []string{"https://github.com/my-org"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an AWS provider without an attribute mapping", func() {
		msg := minimal()
		msg.Spec.AttributeMapping = nil
		msg.Spec.Issuer = &GcpWorkloadIdentityPoolProviderSpec_Aws{
			Aws: &GcpWorkloadIdentityPoolProviderAws{AccountId: "123456789012"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a SAML provider", func() {
		msg := minimal()
		msg.Spec.AttributeMapping = nil
		msg.Spec.Issuer = &GcpWorkloadIdentityPoolProviderSpec_Saml{
			Saml: &GcpWorkloadIdentityPoolProviderSaml{IdpMetadataXml: "<EntityDescriptor/>"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an X.509 provider with a trust store", func() {
		msg := minimal()
		msg.Spec.AttributeMapping = nil
		msg.Spec.Issuer = &GcpWorkloadIdentityPoolProviderSpec_X509{
			X509: &GcpWorkloadIdentityPoolProviderX509{
				TrustStore: &GcpWorkloadIdentityPoolProviderTrustStore{
					TrustAnchors: []*GcpWorkloadIdentityPoolProviderTrustAnchor{{
						PemCertificate: "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",
					}},
					IntermediateCas: []*GcpWorkloadIdentityPoolProviderIntermediateCa{{
						PemCertificate: "-----BEGIN CERTIFICATE-----\nMIIC\n-----END CERTIFICATE-----",
					}},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 4-character provider id (lower bound)", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolProviderId = "gh-1"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 32-character provider id (upper bound)", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolProviderId = strings.Repeat("a", 32)
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing pool reference", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing provider id", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolProviderId = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a provider id with the reserved gcp- prefix", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolProviderId = "gcp-github"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a provider id shorter than 4 characters", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolProviderId = "abc"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a provider id longer than 32 characters", func() {
		msg := minimal()
		msg.Spec.WorkloadIdentityPoolProviderId = strings.Repeat("a", 33)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a spec without an issuer", func() {
		msg := minimal()
		msg.Spec.Issuer = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an OIDC provider without a google.subject mapping", func() {
		msg := minimal()
		msg.Spec.AttributeMapping = map[string]string{
			"attribute.repository": "assertion.repository",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an OIDC provider with a non-URI issuer", func() {
		msg := minimal()
		msg.Spec.GetOidc().IssuerUri = "not a uri"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an OIDC provider with more than 10 audiences", func() {
		msg := minimal()
		audiences := make([]string, 11)
		for i := range audiences {
			audiences[i] = "https://example.com/aud"
		}
		msg.Spec.GetOidc().AllowedAudiences = audiences
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an AWS provider with a non-12-digit account id", func() {
		msg := minimal()
		msg.Spec.AttributeMapping = nil
		msg.Spec.Issuer = &GcpWorkloadIdentityPoolProviderSpec_Aws{
			Aws: &GcpWorkloadIdentityPoolProviderAws{AccountId: "12345"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a SAML provider without metadata", func() {
		msg := minimal()
		msg.Spec.AttributeMapping = nil
		msg.Spec.Issuer = &GcpWorkloadIdentityPoolProviderSpec_Saml{
			Saml: &GcpWorkloadIdentityPoolProviderSaml{},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an X.509 provider without a trust store", func() {
		msg := minimal()
		msg.Spec.AttributeMapping = nil
		msg.Spec.Issuer = &GcpWorkloadIdentityPoolProviderSpec_X509{
			X509: &GcpWorkloadIdentityPoolProviderX509{},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an X.509 trust store without anchors", func() {
		msg := minimal()
		msg.Spec.AttributeMapping = nil
		msg.Spec.Issuer = &GcpWorkloadIdentityPoolProviderSpec_X509{
			X509: &GcpWorkloadIdentityPoolProviderX509{
				TrustStore: &GcpWorkloadIdentityPoolProviderTrustStore{},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an attribute_condition longer than 4096 characters", func() {
		msg := minimal()
		msg.Spec.AttributeCondition = strings.Repeat("c", 4097)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
