package azurefrontdoorcustomdomainv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureFrontDoorCustomDomainSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFrontDoorCustomDomainSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const profileId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd"

const secretId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd/secrets/wildcard-example-com"

const dnsZoneId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/dnszones/example.com"

// minimal valid spec: a managed-certificate domain (Azure's default TLS
// posture).
func minimalSpec() *AzureFrontDoorCustomDomain {
	return &AzureFrontDoorCustomDomain{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFrontDoorCustomDomain",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-front-door-custom-domain",
		},
		Spec: &AzureFrontDoorCustomDomainSpec{
			ProfileId:  literal(profileId),
			DomainName: "www-example-com",
			HostName:   "www.example.com",
			Tls:        &AzureFrontDoorCustomDomainTls{},
		},
	}
}

var _ = ginkgo.Describe("AzureFrontDoorCustomDomainSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal managed-certificate domain", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept an explicit managed certificate type", func() {
			input := minimalSpec()
			input.Spec.Tls.CertificateType = AzureFrontDoorCustomDomainCertificateType_MANAGED_CERTIFICATE
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a customer certificate paired with a secret", func() {
			input := minimalSpec()
			input.Spec.Tls.CertificateType = AzureFrontDoorCustomDomainCertificateType_CUSTOMER_CERTIFICATE
			input.Spec.Tls.SecretId = literal(secretId)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a wildcard hostname with a customer certificate", func() {
			input := minimalSpec()
			input.Spec.HostName = "*.example.com"
			input.Spec.Tls.CertificateType = AzureFrontDoorCustomDomainCertificateType_CUSTOMER_CERTIFICATE
			input.Spec.Tls.SecretId = literal(secretId)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a DNS zone reference", func() {
			input := minimalSpec()
			input.Spec.DnsZoneId = literal(dnsZoneId)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a predefined cipher suite set", func() {
			input := minimalSpec()
			input.Spec.Tls.CipherSuite = &AzureFrontDoorCustomDomainCipherSuite{
				Type: AzureFrontDoorCustomDomainCipherSuiteSetType_TLS12_2023,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept customized ciphers with TLS 1.2 suites only", func() {
			input := minimalSpec()
			input.Spec.Tls.CipherSuite = &AzureFrontDoorCustomDomainCipherSuite{
				Type: AzureFrontDoorCustomDomainCipherSuiteSetType_CUSTOMIZED,
				CustomCiphers: &AzureFrontDoorCustomDomainCustomCiphers{
					Tls12: []string{"ECDHE_RSA_AES256_GCM_SHA384"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept customized ciphers listing both TLS 1.3 suites", func() {
			input := minimalSpec()
			input.Spec.Tls.CipherSuite = &AzureFrontDoorCustomDomainCipherSuite{
				Type: AzureFrontDoorCustomDomainCipherSuiteSetType_CUSTOMIZED,
				CustomCiphers: &AzureFrontDoorCustomDomainCustomCiphers{
					Tls12: []string{"ECDHE_RSA_AES128_GCM_SHA256", "ECDHE_RSA_AES256_GCM_SHA384"},
					Tls13: []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a 64-character hostname with a managed certificate", func() {
			input := minimalSpec()
			input.Spec.HostName = strings.Repeat("a", 51) + ".example.com"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing profile reference", func() {
			input := minimalSpec()
			input.Spec.ProfileId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing domain name", func() {
			input := minimalSpec()
			input.Spec.DomainName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a domain name with dots", func() {
			input := minimalSpec()
			input.Spec.DomainName = "www.example.com"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing host name", func() {
			input := minimalSpec()
			input.Spec.HostName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a single-label host name", func() {
			input := minimalSpec()
			input.Spec.HostName = "localhost"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a host name label starting with a hyphen", func() {
			input := minimalSpec()
			input.Spec.HostName = "-www.example.com"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a wildcard anywhere but the first label", func() {
			input := minimalSpec()
			input.Spec.HostName = "www.*.example.com"
			input.Spec.Tls.CertificateType = AzureFrontDoorCustomDomainCertificateType_CUSTOMER_CERTIFICATE
			input.Spec.Tls.SecretId = literal(secretId)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing TLS block", func() {
			input := minimalSpec()
			input.Spec.Tls = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a wildcard hostname with a managed certificate", func() {
			input := minimalSpec()
			input.Spec.HostName = "*.example.com"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a hostname over 64 characters with a managed certificate", func() {
			input := minimalSpec()
			input.Spec.HostName = strings.Repeat("a", 60) + ".example.com"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a hostname over 64 characters with a customer certificate", func() {
			input := minimalSpec()
			input.Spec.HostName = strings.Repeat("a", 60) + ".example.com"
			input.Spec.Tls.CertificateType = AzureFrontDoorCustomDomainCertificateType_CUSTOMER_CERTIFICATE
			input.Spec.Tls.SecretId = literal(secretId)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should reject a customer certificate without a secret", func() {
			input := minimalSpec()
			input.Spec.Tls.CertificateType = AzureFrontDoorCustomDomainCertificateType_CUSTOMER_CERTIFICATE
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a managed certificate paired with a secret", func() {
			input := minimalSpec()
			input.Spec.Tls.SecretId = literal(secretId)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a cipher suite without a type", func() {
			input := minimalSpec()
			input.Spec.Tls.CipherSuite = &AzureFrontDoorCustomDomainCipherSuite{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject CUSTOMIZED without custom ciphers", func() {
			input := minimalSpec()
			input.Spec.Tls.CipherSuite = &AzureFrontDoorCustomDomainCipherSuite{
				Type: AzureFrontDoorCustomDomainCipherSuiteSetType_CUSTOMIZED,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject custom ciphers on a predefined set", func() {
			input := minimalSpec()
			input.Spec.Tls.CipherSuite = &AzureFrontDoorCustomDomainCipherSuite{
				Type: AzureFrontDoorCustomDomainCipherSuiteSetType_TLS12_2022,
				CustomCiphers: &AzureFrontDoorCustomDomainCustomCiphers{
					Tls12: []string{"ECDHE_RSA_AES128_GCM_SHA256"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject customized ciphers with an empty TLS 1.2 list", func() {
			input := minimalSpec()
			input.Spec.Tls.CipherSuite = &AzureFrontDoorCustomDomainCipherSuite{
				Type: AzureFrontDoorCustomDomainCipherSuiteSetType_CUSTOMIZED,
				CustomCiphers: &AzureFrontDoorCustomDomainCustomCiphers{
					Tls13: []string{"TLS_AES_128_GCM_SHA256", "TLS_AES_256_GCM_SHA384"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown TLS 1.2 cipher", func() {
			input := minimalSpec()
			input.Spec.Tls.CipherSuite = &AzureFrontDoorCustomDomainCipherSuite{
				Type: AzureFrontDoorCustomDomainCipherSuiteSetType_CUSTOMIZED,
				CustomCiphers: &AzureFrontDoorCustomDomainCustomCiphers{
					Tls12: []string{"DHE_RSA_AES128_GCM_SHA256"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a single TLS 1.3 cipher (both are mandatory when set)", func() {
			input := minimalSpec()
			input.Spec.Tls.CipherSuite = &AzureFrontDoorCustomDomainCipherSuite{
				Type: AzureFrontDoorCustomDomainCipherSuiteSetType_CUSTOMIZED,
				CustomCiphers: &AzureFrontDoorCustomDomainCustomCiphers{
					Tls12: []string{"ECDHE_RSA_AES128_GCM_SHA256"},
					Tls13: []string{"TLS_AES_128_GCM_SHA256"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a wrong kind", func() {
			input := minimalSpec()
			input.Kind = "AzureFrontDoorCustomDomains"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject missing metadata", func() {
			input := minimalSpec()
			input.Metadata = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
