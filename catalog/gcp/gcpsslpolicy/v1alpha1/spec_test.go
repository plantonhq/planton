package gcpsslpolicyv1alpha1

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
	ginkgo.RunSpecs(t, "GcpSslPolicySpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpSslPolicySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpSslPolicy {
		return &GcpSslPolicy{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpSslPolicy",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-ssl-policy",
			},
			Spec: &GcpSslPolicySpec{},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal spec (GCP defaults: COMPATIBLE, TLS_1_0)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an explicit ssl_policy_name", func() {
		target := minimal()
		target.Spec.SslPolicyName = "prod-tls-floor"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a regional policy", func() {
		target := minimal()
		target.Spec.Region = "us-central1"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a description", func() {
		target := minimal()
		target.Spec.Description = "TLS 1.2 floor for PCI-scoped frontends"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each predefined profile", func() {
		for _, profile := range []string{"COMPATIBLE", "MODERN", "RESTRICTED"} {
			target := minimal()
			target.Spec.Profile = profile
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept each min_tls_version", func() {
		for _, v := range []string{"TLS_1_0", "TLS_1_1", "TLS_1_2"} {
			target := minimal()
			target.Spec.MinTlsVersion = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept a RESTRICTED profile with a TLS_1_2 floor", func() {
		target := minimal()
		target.Spec.Profile = "RESTRICTED"
		target.Spec.MinTlsVersion = "TLS_1_2"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept the FIPS_202205 profile with its required TLS_1_2 floor", func() {
		target := minimal()
		target.Spec.Profile = "FIPS_202205"
		target.Spec.MinTlsVersion = "TLS_1_2"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a TLS_1_3 floor on the RESTRICTED profile", func() {
		target := minimal()
		target.Spec.Profile = "RESTRICTED"
		target.Spec.MinTlsVersion = "TLS_1_3"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each post_quantum_key_exchange stance", func() {
		for _, v := range []string{"DEFAULT", "ENABLED", "DEFERRED"} {
			target := minimal()
			target.Spec.PostQuantumKeyExchange = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept a CUSTOM profile with cipher suites", func() {
		target := minimal()
		target.Spec.Profile = "CUSTOM"
		target.Spec.CustomFeatures = []string{
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject an invalid profile", func() {
		target := minimal()
		target.Spec.Profile = "STRICT"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "COMPATIBLE")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid min_tls_version", func() {
		target := minimal()
		target.Spec.MinTlsVersion = "TLS_1_4"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject the FIPS_202205 profile without a TLS_1_2 floor", func() {
		for _, v := range []string{"", "TLS_1_0", "TLS_1_1"} {
			target := minimal()
			target.Spec.Profile = "FIPS_202205"
			target.Spec.MinTlsVersion = v
			err := validator.Validate(target)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(strings.Contains(err.Error(), "TLS_1_2")).To(gomega.BeTrue())
		}
	})

	ginkgo.It("should reject a TLS_1_3 floor on any profile but RESTRICTED", func() {
		for _, p := range []string{"", "COMPATIBLE", "MODERN"} {
			target := minimal()
			target.Spec.Profile = p
			target.Spec.MinTlsVersion = "TLS_1_3"
			err := validator.Validate(target)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(strings.Contains(err.Error(), "RESTRICTED")).To(gomega.BeTrue())
		}
	})

	ginkgo.It("should reject an invalid post_quantum_key_exchange stance", func() {
		target := minimal()
		target.Spec.PostQuantumKeyExchange = "ALWAYS"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a CUSTOM profile without custom_features", func() {
		target := minimal()
		target.Spec.Profile = "CUSTOM"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "custom_features")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject custom_features on a non-CUSTOM profile", func() {
		target := minimal()
		target.Spec.Profile = "MODERN"
		target.Spec.CustomFeatures = []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "CUSTOM")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject custom_features with the default (empty) profile", func() {
		target := minimal()
		target.Spec.CustomFeatures = []string{"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a lowercase cipher suite name", func() {
		target := minimal()
		target.Spec.Profile = "CUSTOM"
		target.Spec.CustomFeatures = []string{"tls_ecdhe_rsa_with_aes_128_gcm_sha256"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid ssl_policy_name", func() {
		target := minimal()
		target.Spec.SslPolicyName = "Invalid_Name"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "RFC1035")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a name longer than 63 characters", func() {
		target := minimal()
		target.Spec.SslPolicyName = "a" + strings.Repeat("b", 63)
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid region", func() {
		target := minimal()
		target.Spec.Region = "US_CENTRAL1"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a description over 2048 characters", func() {
		target := minimal()
		target.Spec.Description = strings.Repeat("x", 2049)
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind constant", func() {
		target := minimal()
		target.Kind = "GcpSslCertificate"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong api_version", func() {
		target := minimal()
		target.ApiVersion = "gcp.planton.dev/v2"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
