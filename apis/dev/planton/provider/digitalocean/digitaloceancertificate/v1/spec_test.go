package digitaloceancertificatev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestDigitalOceanCertificateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanCertificateSpec Validation Suite")
}

var _ = ginkgo.Describe("DigitalOceanCertificateSpec validations", func() {

	// Helper function to create a minimal valid Let's Encrypt spec
	makeValidLetsEncryptSpec := func() *DigitalOceanCertificateSpec {
		return &DigitalOceanCertificateSpec{
			CertificateName: "prod-web-cert",
			Type:            DigitalOceanCertificateType_lets_encrypt,
			CertificateSource: &DigitalOceanCertificateSpec_LetsEncrypt{
				LetsEncrypt: &DigitalOceanCertificateLetsEncryptParams{
					Domains: []string{"example.com", "www.example.com"},
				},
			},
		}
	}

	// Helper function to create a minimal valid custom certificate spec
	makeValidCustomSpec := func() *DigitalOceanCertificateSpec {
		return &DigitalOceanCertificateSpec{
			CertificateName: "prod-custom-cert",
			Type:            DigitalOceanCertificateType_custom,
			CertificateSource: &DigitalOceanCertificateSpec_Custom{
				Custom: &DigitalOceanCertificateCustomParams{
					LeafCertificate:  "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
					PrivateKey:       "-----BEGIN PRIVATE KEY-----\nMIIE...\n-----END PRIVATE KEY-----",
					CertificateChain: "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----",
				},
			},
		}
	}

	ginkgo.Context("Required fields", func() {
		ginkgo.It("accepts a minimal valid Let's Encrypt spec", func() {
			spec := makeValidLetsEncryptSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a minimal valid custom certificate spec", func() {
			spec := makeValidCustomSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing certificate_name", func() {
			spec := makeValidLetsEncryptSpec()
			spec.CertificateName = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing type", func() {
			spec := makeValidLetsEncryptSpec()
			spec.Type = DigitalOceanCertificateType(999) // Invalid enum value
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing certificate_source (oneof)", func() {
			spec := &DigitalOceanCertificateSpec{
				CertificateName:   "test-cert",
				Type:              DigitalOceanCertificateType_lets_encrypt,
				CertificateSource: nil,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("certificate_name validation", func() {
		ginkgo.It("accepts certificate_name with 64 characters (max)", func() {
			spec := makeValidLetsEncryptSpec()
			spec.CertificateName = "a123456789b123456789c123456789d123456789e123456789f123456789abcd" // 64 chars
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects certificate_name exceeding 64 characters", func() {
			spec := makeValidLetsEncryptSpec()
			spec.CertificateName = "a123456789b123456789c123456789d123456789e123456789f123456789abcde" // 65 chars
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects empty certificate_name", func() {
			spec := makeValidLetsEncryptSpec()
			spec.CertificateName = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Let's Encrypt parameters validation", func() {
		ginkgo.It("rejects Let's Encrypt spec with empty domains list", func() {
			spec := makeValidLetsEncryptSpec()
			spec.GetLetsEncrypt().Domains = []string{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects Let's Encrypt spec with nil domains", func() {
			spec := makeValidLetsEncryptSpec()
			spec.GetLetsEncrypt().Domains = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts Let's Encrypt spec with valid domain", func() {
			spec := makeValidLetsEncryptSpec()
			spec.GetLetsEncrypt().Domains = []string{"example.com"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts Let's Encrypt spec with wildcard domain", func() {
			spec := makeValidLetsEncryptSpec()
			spec.GetLetsEncrypt().Domains = []string{"*.example.com"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts Let's Encrypt spec with multiple domains", func() {
			spec := makeValidLetsEncryptSpec()
			spec.GetLetsEncrypt().Domains = []string{"example.com", "www.example.com", "api.example.com"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects Let's Encrypt spec with duplicate domains", func() {
			spec := makeValidLetsEncryptSpec()
			spec.GetLetsEncrypt().Domains = []string{"example.com", "example.com"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects Let's Encrypt spec with invalid domain pattern (no TLD)", func() {
			spec := makeValidLetsEncryptSpec()
			spec.GetLetsEncrypt().Domains = []string{"localhost"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects Let's Encrypt spec with invalid domain pattern (IP address)", func() {
			spec := makeValidLetsEncryptSpec()
			spec.GetLetsEncrypt().Domains = []string{"192.168.1.1"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts Let's Encrypt spec with disable_auto_renew set to true", func() {
			spec := makeValidLetsEncryptSpec()
			spec.GetLetsEncrypt().DisableAutoRenew = true
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts Let's Encrypt spec with disable_auto_renew set to false (default)", func() {
			spec := makeValidLetsEncryptSpec()
			spec.GetLetsEncrypt().DisableAutoRenew = false
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Custom certificate parameters validation", func() {
		ginkgo.It("rejects custom spec with missing leaf_certificate", func() {
			spec := makeValidCustomSpec()
			spec.GetCustom().LeafCertificate = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects custom spec with missing private_key", func() {
			spec := makeValidCustomSpec()
			spec.GetCustom().PrivateKey = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts custom spec without certificate_chain (optional)", func() {
			spec := makeValidCustomSpec()
			spec.GetCustom().CertificateChain = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts custom spec with complete certificate chain", func() {
			spec := makeValidCustomSpec()
			spec.GetCustom().CertificateChain = "-----BEGIN CERTIFICATE-----\nMIIC...\n-----END CERTIFICATE-----"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts custom spec with realistic PEM-formatted leaf certificate", func() {
			spec := makeValidCustomSpec()
			spec.GetCustom().LeafCertificate = `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKL0UG+mRKSzMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMTcwODIzMTUzODU3WhcNMTgwODIzMTUzODU3WjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAz8H/xqDxqGxFQSQYpXVfbfX8PGD0ov9pCGPmEpRLFYH4FmRqYvPc+JMX
wLQDULVAqLhQvh1v6jjMRR6lMdXJP0QPmvqh0GJkKH7zJVyT7QzJBZCJsYL4t5bJ
-----END CERTIFICATE-----`
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Optional fields validation", func() {
		ginkgo.It("accepts spec with valid description (128 chars max)", func() {
			spec := makeValidLetsEncryptSpec()
			spec.Description = "This is a test certificate for production web and API endpoints with auto-renewal enabled for maximum uptime and security."
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects spec with description exceeding 128 characters", func() {
			spec := makeValidLetsEncryptSpec()
			spec.Description = "This is a very long description that exceeds the maximum allowed length of 128 characters and should be rejected by validation rules."
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts spec with valid tags", func() {
			spec := makeValidLetsEncryptSpec()
			spec.Tags = []string{"env:production", "team:platform", "managed:letsencrypt"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects spec with duplicate tags", func() {
			spec := makeValidLetsEncryptSpec()
			spec.Tags = []string{"env:production", "env:production"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts spec with empty tags list", func() {
			spec := makeValidLetsEncryptSpec()
			spec.Tags = []string{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Discriminated union (oneof) validation", func() {
		ginkgo.It("accepts spec with type=letsEncrypt and lets_encrypt params", func() {
			spec := makeValidLetsEncryptSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts spec with type=custom and custom params", func() {
			spec := makeValidCustomSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		// Note: protobuf oneof enforces mutual exclusivity at the message level,
		// so we can't set both lets_encrypt and custom simultaneously in a valid message.
		// The following test verifies that neither can be nil when required.
		ginkgo.It("rejects spec with type set but no certificate_source", func() {
			spec := &DigitalOceanCertificateSpec{
				CertificateName:   "test-cert",
				Type:              DigitalOceanCertificateType_lets_encrypt,
				CertificateSource: nil, // Missing oneof
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Edge cases", func() {
		ginkgo.It("accepts Let's Encrypt spec with subdomain", func() {
			spec := makeValidLetsEncryptSpec()
			spec.GetLetsEncrypt().Domains = []string{"api.staging.example.com"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts Let's Encrypt spec with hyphenated domain", func() {
			spec := makeValidLetsEncryptSpec()
			spec.GetLetsEncrypt().Domains = []string{"my-app.example.com"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts custom spec with minimal PEM content (1 char min_len)", func() {
			spec := makeValidCustomSpec()
			spec.GetCustom().LeafCertificate = "x" // min_len = 1
			spec.GetCustom().PrivateKey = "y"      // min_len = 1
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts certificate_name with hyphens and numbers", func() {
			spec := makeValidLetsEncryptSpec()
			spec.CertificateName = "prod-web-cert-2025-v2"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
