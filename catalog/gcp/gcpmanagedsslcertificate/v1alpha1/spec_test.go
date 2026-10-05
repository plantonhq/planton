package gcpmanagedsslcertificatev1alpha1

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
	ginkgo.RunSpecs(t, "GcpManagedSslCertificateSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpManagedSslCertificateSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpManagedSslCertificate {
		return &GcpManagedSslCertificate{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpManagedSslCertificate",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-cert",
			},
			Spec: &GcpManagedSslCertificateSpec{
				Domains: []string{"app.example.com"},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal spec with one domain", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept multiple domains", func() {
		target := minimal()
		target.Spec.Domains = []string{"example.com", "www.example.com", "api.example.com"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an explicit certificate_name", func() {
		target := minimal()
		target.Spec.CertificateName = "prod-app-cert"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a description", func() {
		target := minimal()
		target.Spec.Description = "TLS for the production app load balancer"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a domain with a trailing dot", func() {
		target := minimal()
		target.Spec.Domains = []string{"app.example.com."}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a multi-label subdomain", func() {
		target := minimal()
		target.Spec.Domains = []string{"v2.api.example.co.uk"}
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

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a spec with no domains", func() {
		target := minimal()
		target.Spec.Domains = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty domains list", func() {
		target := minimal()
		target.Spec.Domains = []string{}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid certificate_name", func() {
		target := minimal()
		target.Spec.CertificateName = "Invalid_Name"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "RFC1035")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a wildcard domain", func() {
		target := minimal()
		target.Spec.Domains = []string{"*.example.com"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "fully-qualified")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a bare hostname without a TLD", func() {
		target := minimal()
		target.Spec.Domains = []string{"localhost"}
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
