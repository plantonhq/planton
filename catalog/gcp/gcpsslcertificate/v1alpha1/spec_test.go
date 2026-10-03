package gcpsslcertificatev1alpha1

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
	ginkgo.RunSpecs(t, "GcpSslCertificateSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

// Syntactically-shaped PEM stand-ins: spec validation checks the certificate's
// PEM framing, not cryptographic validity (GCP verifies the real material at
// deploy time). The private key carries no framing rule — it is sensitive, so
// consuming platforms store a managed-secret reference in the field.
const (
	testCertPem  = "-----BEGIN CERTIFICATE-----\nMIIBszCCAVmgAwIBAgIUTest\n-----END CERTIFICATE-----\n"
	testPkcs8Pem = "-----BEGIN PRIVATE KEY-----\nMIGHAgEAMBMGByqGSM49Test\n-----END PRIVATE KEY-----\n"
	testRsaPem   = "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEATest\n-----END RSA PRIVATE KEY-----\n"
	testEcPem    = "-----BEGIN EC PRIVATE KEY-----\nMHcCAQEEITest\n-----END EC PRIVATE KEY-----\n"
)

var _ = ginkgo.Describe("GcpSslCertificateSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpSslCertificate {
		return &GcpSslCertificate{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpSslCertificate",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-ssl-cert",
			},
			Spec: &GcpSslCertificateSpec{
				Certificate: testCertPem,
				PrivateKey:  testPkcs8Pem,
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal spec with PEM cert and key", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a project_id literal", func() {
		target := minimal()
		target.Spec.ProjectId = litRef("my-gcp-project-123")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an explicit certificate_name", func() {
		target := minimal()
		target.Spec.CertificateName = "prod-app-cert-2026"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a regional certificate", func() {
		target := minimal()
		target.Spec.Region = "us-central1"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a description", func() {
		target := minimal()
		target.Spec.Description = "Wildcard cert from the corporate CA; rotate yearly"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a PKCS#1 RSA private key", func() {
		target := minimal()
		target.Spec.PrivateKey = testRsaPem
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a SEC1 EC private key", func() {
		target := minimal()
		target.Spec.PrivateKey = testEcPem
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

	ginkgo.It("should reject a missing certificate", func() {
		target := minimal()
		target.Spec.Certificate = ""
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a missing private_key", func() {
		target := minimal()
		target.Spec.PrivateKey = ""
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a non-PEM certificate", func() {
		target := minimal()
		target.Spec.Certificate = "MIIBszCCAVmgAwIBAgIUTest"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "PEM")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a private key pasted into the certificate field", func() {
		target := minimal()
		target.Spec.Certificate = testPkcs8Pem
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

	ginkgo.It("should reject a name longer than 63 characters", func() {
		target := minimal()
		target.Spec.CertificateName = "a" + strings.Repeat("b", 63)
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
		target.Kind = "GcpManagedSslCertificate"
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
