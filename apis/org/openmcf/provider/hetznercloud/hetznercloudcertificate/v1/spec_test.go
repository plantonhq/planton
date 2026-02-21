package hetznercloudcertificatev1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHetznerCloudCertificateSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HetznerCloudCertificateSpec Validation Suite")
}

var _ = Describe("HetznerCloudCertificateSpec validations", func() {

	Context("with an uploaded certificate", func() {
		It("should accept a valid uploaded certificate spec", func() {
			spec := &HetznerCloudCertificateSpec{
				Certificate: &HetznerCloudCertificateSpec_Uploaded{
					Uploaded: &UploadedCertificateConfig{
						Certificate: "-----BEGIN CERTIFICATE-----\nMIIExample\n-----END CERTIFICATE-----",
						PrivateKey:  "-----BEGIN PRIVATE KEY-----\nMIIExample\n-----END PRIVATE KEY-----",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should reject an uploaded certificate with empty certificate PEM", func() {
			spec := &HetznerCloudCertificateSpec{
				Certificate: &HetznerCloudCertificateSpec_Uploaded{
					Uploaded: &UploadedCertificateConfig{
						Certificate: "",
						PrivateKey:  "-----BEGIN PRIVATE KEY-----\nMIIExample\n-----END PRIVATE KEY-----",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject an uploaded certificate with empty private key", func() {
			spec := &HetznerCloudCertificateSpec{
				Certificate: &HetznerCloudCertificateSpec_Uploaded{
					Uploaded: &UploadedCertificateConfig{
						Certificate: "-----BEGIN CERTIFICATE-----\nMIIExample\n-----END CERTIFICATE-----",
						PrivateKey:  "",
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject an uploaded certificate with both fields empty", func() {
			spec := &HetznerCloudCertificateSpec{
				Certificate: &HetznerCloudCertificateSpec_Uploaded{
					Uploaded: &UploadedCertificateConfig{},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("with a managed certificate", func() {
		It("should accept a valid managed certificate with one domain", func() {
			spec := &HetznerCloudCertificateSpec{
				Certificate: &HetznerCloudCertificateSpec_Managed{
					Managed: &ManagedCertificateConfig{
						DomainNames: []string{"example.com"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should accept a valid managed certificate with multiple domains", func() {
			spec := &HetznerCloudCertificateSpec{
				Certificate: &HetznerCloudCertificateSpec_Managed{
					Managed: &ManagedCertificateConfig{
						DomainNames: []string{"example.com", "www.example.com", "api.example.com"},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).To(BeNil())
		})

		It("should reject a managed certificate with empty domain_names", func() {
			spec := &HetznerCloudCertificateSpec{
				Certificate: &HetznerCloudCertificateSpec_Managed{
					Managed: &ManagedCertificateConfig{
						DomainNames: []string{},
					},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})

		It("should reject a managed certificate with nil domain_names", func() {
			spec := &HetznerCloudCertificateSpec{
				Certificate: &HetznerCloudCertificateSpec_Managed{
					Managed: &ManagedCertificateConfig{},
				},
			}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("with no certificate variant set", func() {
		It("should reject a spec with no oneof branch set", func() {
			spec := &HetznerCloudCertificateSpec{}
			err := protovalidate.Validate(spec)
			Expect(err).ToNot(BeNil())
		})
	})
})
