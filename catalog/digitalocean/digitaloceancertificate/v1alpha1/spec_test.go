package digitaloceancertificatev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestDigitalOceanCertificateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanCertificateSpec Custom Validation Tests")
}

// letsEncryptCert returns a minimal valid Let's Encrypt certificate the tests mutate per case.
func letsEncryptCert() *DigitalOceanCertificate {
	return &DigitalOceanCertificate{
		ApiVersion: "digital-ocean.planton.dev/v1alpha1",
		Kind:       "DigitalOceanCertificate",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-certificate",
		},
		Spec: &DigitalOceanCertificateSpec{
			CertificateName: "web-cert",
			CertificateSource: &DigitalOceanCertificateSpec_LetsEncrypt{
				LetsEncrypt: &DigitalOceanCertificateLetsEncryptParams{
					Domains: []string{"example.com", "www.example.com"},
				},
			},
		},
	}
}

// customCert returns a minimal valid custom certificate the tests mutate per case.
func customCert() *DigitalOceanCertificate {
	return &DigitalOceanCertificate{
		ApiVersion: "digital-ocean.planton.dev/v1alpha1",
		Kind:       "DigitalOceanCertificate",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-certificate",
		},
		Spec: &DigitalOceanCertificateSpec{
			CertificateName: "uploaded-cert",
			CertificateSource: &DigitalOceanCertificateSpec_Custom{
				Custom: &DigitalOceanCertificateCustomParams{
					LeafCertificate: "-----BEGIN CERTIFICATE-----\nMIIB...\n-----END CERTIFICATE-----",
					PrivateKey:      "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("DigitalOceanCertificateSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal Let's Encrypt certificate", func() {
			gomega.Expect(protovalidate.Validate(letsEncryptCert())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a Let's Encrypt certificate with a wildcard domain", func() {
			input := letsEncryptCert()
			input.Spec.GetLetsEncrypt().Domains = []string{"example.com", "*.example.com"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a minimal custom certificate without a chain", func() {
			gomega.Expect(protovalidate.Validate(customCert())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a custom certificate with an intermediate chain", func() {
			input := customCert()
			input.Spec.GetCustom().CertificateChain = "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a certificate name of exactly 64 characters", func() {
			input := letsEncryptCert()
			input.Spec.CertificateName = "a234567890123456789012345678901234567890123456789012345678901234"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("certificate_name validation", func() {

		ginkgo.It("rejects an empty certificate name", func() {
			input := letsEncryptCert()
			input.Spec.CertificateName = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a certificate name longer than 64 characters", func() {
			input := letsEncryptCert()
			input.Spec.CertificateName = "a2345678901234567890123456789012345678901234567890123456789012345"
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("certificate_source oneof", func() {

		ginkgo.It("rejects a spec with no certificate source", func() {
			input := letsEncryptCert()
			input.Spec.CertificateSource = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("Let's Encrypt parameters", func() {

		ginkgo.It("rejects an empty domains list", func() {
			input := letsEncryptCert()
			input.Spec.GetLetsEncrypt().Domains = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects duplicate domains", func() {
			input := letsEncryptCert()
			input.Spec.GetLetsEncrypt().Domains = []string{"example.com", "example.com"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a bare label that is not a fully-qualified domain", func() {
			input := letsEncryptCert()
			input.Spec.GetLetsEncrypt().Domains = []string{"localhost"}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty domain entry", func() {
			input := letsEncryptCert()
			input.Spec.GetLetsEncrypt().Domains = []string{""}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("Custom certificate parameters", func() {

		ginkgo.It("rejects a custom certificate without a leaf certificate", func() {
			input := customCert()
			input.Spec.GetCustom().LeafCertificate = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a custom certificate without a private key", func() {
			input := customCert()
			input.Spec.GetCustom().PrivateKey = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("accepts an empty certificate chain (it is optional)", func() {
			input := customCert()
			input.Spec.GetCustom().CertificateChain = ""
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})
})
