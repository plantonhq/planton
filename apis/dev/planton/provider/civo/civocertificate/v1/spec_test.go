package civocertificatev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestCivoCertificateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CivoCertificateSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CivoCertificateSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("Let's Encrypt certificate with minimal fields", func() {

			ginkgo.It("should not return a validation error for single domain", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "simple-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "simple-cert",
						Type:            CivoCertificateType_letsEncrypt,
						CertificateSource: &CivoCertificateSpec_LetsEncrypt{
							LetsEncrypt: &CivoCertificateLetsEncryptParams{
								Domains: []string{"example.com"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for wildcard domain", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "wildcard-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "wildcard-cert",
						Type:            CivoCertificateType_letsEncrypt,
						CertificateSource: &CivoCertificateSpec_LetsEncrypt{
							LetsEncrypt: &CivoCertificateLetsEncryptParams{
								Domains: []string{"*.example.com", "example.com"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with auto-renew disabled", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-renew-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "no-renew-cert",
						Type:            CivoCertificateType_letsEncrypt,
						CertificateSource: &CivoCertificateSpec_LetsEncrypt{
							LetsEncrypt: &CivoCertificateLetsEncryptParams{
								Domains:          []string{"test.example.com"},
								DisableAutoRenew: true,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with description and tags", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "tagged-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "tagged-cert",
						Type:            CivoCertificateType_letsEncrypt,
						CertificateSource: &CivoCertificateSpec_LetsEncrypt{
							LetsEncrypt: &CivoCertificateLetsEncryptParams{
								Domains: []string{"api.example.com"},
							},
						},
						Description: "Production API certificate",
						Tags:        []string{"env:prod", "service:api", "critical:true"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("Custom certificate with all fields", func() {

			ginkgo.It("should not return a validation error for custom cert with chain", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "custom-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "custom-cert",
						Type:            CivoCertificateType_custom,
						CertificateSource: &CivoCertificateSpec_Custom{
							Custom: &CivoCertificateCustomParams{
								LeafCertificate:  "-----BEGIN CERTIFICATE-----\nMIIFake...\n-----END CERTIFICATE-----",
								PrivateKey:       "-----BEGIN RSA PRIVATE KEY-----\nMIIFake...\n-----END RSA PRIVATE KEY-----",
								CertificateChain: "-----BEGIN CERTIFICATE-----\nMIIIntermediate...\n-----END CERTIFICATE-----",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for custom cert without chain", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "custom-cert-no-chain",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "custom-cert-no-chain",
						Type:            CivoCertificateType_custom,
						CertificateSource: &CivoCertificateSpec_Custom{
							Custom: &CivoCertificateCustomParams{
								LeafCertificate: "-----BEGIN CERTIFICATE-----\nMIIFake...\n-----END CERTIFICATE-----",
								PrivateKey:      "-----BEGIN RSA PRIVATE KEY-----\nMIIFake...\n-----END RSA PRIVATE KEY-----",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("certificate_name validation", func() {

			ginkgo.It("should return a validation error when certificate_name is empty", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "",
						Type:            CivoCertificateType_letsEncrypt,
						CertificateSource: &CivoCertificateSpec_LetsEncrypt{
							LetsEncrypt: &CivoCertificateLetsEncryptParams{
								Domains: []string{"example.com"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when certificate_name exceeds 64 chars", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "this-is-a-very-long-certificate-name-that-exceeds-sixty-four-characters-limit",
						Type:            CivoCertificateType_letsEncrypt,
						CertificateSource: &CivoCertificateSpec_LetsEncrypt{
							LetsEncrypt: &CivoCertificateLetsEncryptParams{
								Domains: []string{"example.com"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("type validation", func() {

			ginkgo.It("should return a validation error when type is not set", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "test-cert",
						// Type not set
						CertificateSource: &CivoCertificateSpec_LetsEncrypt{
							LetsEncrypt: &CivoCertificateLetsEncryptParams{
								Domains: []string{"example.com"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("certificate_source validation (oneof)", func() {

			ginkgo.It("should return a validation error when certificate_source is not set", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "test-cert",
						Type:            CivoCertificateType_letsEncrypt,
						// CertificateSource not set
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("Let's Encrypt domains validation", func() {

			ginkgo.It("should return a validation error when domains list is empty", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "test-cert",
						Type:            CivoCertificateType_letsEncrypt,
						CertificateSource: &CivoCertificateSpec_LetsEncrypt{
							LetsEncrypt: &CivoCertificateLetsEncryptParams{
								Domains: []string{},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when domain is invalid (no TLD)", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "test-cert",
						Type:            CivoCertificateType_letsEncrypt,
						CertificateSource: &CivoCertificateSpec_LetsEncrypt{
							LetsEncrypt: &CivoCertificateLetsEncryptParams{
								Domains: []string{"localhost"}, // invalid: no TLD
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when domains are not unique", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "test-cert",
						Type:            CivoCertificateType_letsEncrypt,
						CertificateSource: &CivoCertificateSpec_LetsEncrypt{
							LetsEncrypt: &CivoCertificateLetsEncryptParams{
								Domains: []string{"example.com", "example.com"}, // duplicate
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("Custom certificate validation", func() {

			ginkgo.It("should return a validation error when leaf_certificate is empty", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "test-cert",
						Type:            CivoCertificateType_custom,
						CertificateSource: &CivoCertificateSpec_Custom{
							Custom: &CivoCertificateCustomParams{
								LeafCertificate: "", // empty
								PrivateKey:      "-----BEGIN RSA PRIVATE KEY-----\nMIIFake...\n-----END RSA PRIVATE KEY-----",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when private_key is empty", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "test-cert",
						Type:            CivoCertificateType_custom,
						CertificateSource: &CivoCertificateSpec_Custom{
							Custom: &CivoCertificateCustomParams{
								LeafCertificate: "-----BEGIN CERTIFICATE-----\nMIIFake...\n-----END CERTIFICATE-----",
								PrivateKey:      "", // empty
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("description validation", func() {

			ginkgo.It("should return a validation error when description exceeds 128 chars", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "test-cert",
						Type:            CivoCertificateType_letsEncrypt,
						CertificateSource: &CivoCertificateSpec_LetsEncrypt{
							LetsEncrypt: &CivoCertificateLetsEncryptParams{
								Domains: []string{"example.com"},
							},
						},
						Description: "This is a very long description that exceeds the maximum allowed length of one hundred and twenty eight characters for the certificate description field",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("tags validation", func() {

			ginkgo.It("should return a validation error when tags are not unique", func() {
				input := &CivoCertificate{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoCertificate",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cert",
					},
					Spec: &CivoCertificateSpec{
						CertificateName: "test-cert",
						Type:            CivoCertificateType_letsEncrypt,
						CertificateSource: &CivoCertificateSpec_LetsEncrypt{
							LetsEncrypt: &CivoCertificateLetsEncryptParams{
								Domains: []string{"example.com"},
							},
						},
						Tags: []string{"env:prod", "env:prod"}, // duplicate
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
