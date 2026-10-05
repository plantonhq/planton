package awscertmanagercertv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsCertManagerCert(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsCertManagerCert Suite")
}

var _ = ginkgo.Describe("AwsCertManagerCert", func() {

	var input *AwsCertManagerCert

	ginkgo.BeforeEach(func() {
		input = &AwsCertManagerCert{
			ApiVersion: "aws.planton.dev/v1alpha1",
			Kind:       "AwsCertManagerCert",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "a-test-name",
			},
			Spec: &AwsCertManagerCertSpec{
				Region:            "us-west-2",
				PrimaryDomainName: "example.com",
				AlternateDomainNames: []string{
					"www.example.com",
					"test.example.com",
				},
				ValidationMethod: "DNS",
				Route53HostedZoneId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-zone-id"},
				},
			},
		}
	})

	ginkgo.Context("when valid input is passed", func() {
		ginkgo.It("should not return a validation error", func() {
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Domain Pattern Validations", func() {

		ginkgo.Context("PrimaryDomainName", func() {
			ginkgo.It("should accept a valid apex domain", func() {
				input.Spec.PrimaryDomainName = "example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a valid wildcard domain", func() {
				input.Spec.PrimaryDomainName = "*.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should reject a domain missing a TLD", func() {
				input.Spec.PrimaryDomainName = "example"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject multiple wildcard asterisks", func() {
				input.Spec.PrimaryDomainName = "**.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a domain with invalid characters", func() {
				input.Spec.PrimaryDomainName = "exa@mple.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("AlternateDomainNames", func() {
			ginkgo.It("should accept multiple valid domains", func() {
				input.Spec.AlternateDomainNames = []string{"www.example.com", "*.foo.org"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should reject if any domain is invalid", func() {
				input.Spec.AlternateDomainNames = []string{"www.example.com", "invalid@@domain"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("Creation Modes (exactly_one_creation_mode)", func() {
		ginkgo.It("should reject a spec with neither a domain nor imported material", func() {
			input.Spec.PrimaryDomainName = ""
			input.Spec.AlternateDomainNames = nil
			input.Spec.ValidationMethod = ""
			input.Spec.Route53HostedZoneId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a spec with both a domain and imported material", func() {
			input.Spec.Imported = &AwsCertManagerCertImported{
				CertificateBody: "-----BEGIN CERTIFICATE-----",
				PrivateKey:      "-----BEGIN PRIVATE KEY-----",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a pure imported spec", func() {
			input.Spec.PrimaryDomainName = ""
			input.Spec.AlternateDomainNames = nil
			input.Spec.ValidationMethod = ""
			input.Spec.Route53HostedZoneId = nil
			input.Spec.Imported = &AwsCertManagerCertImported{
				CertificateBody: "-----BEGIN CERTIFICATE-----",
				PrivateKey:      "-----BEGIN PRIVATE KEY-----",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject imported material missing the private key", func() {
			input.Spec.PrimaryDomainName = ""
			input.Spec.AlternateDomainNames = nil
			input.Spec.ValidationMethod = ""
			input.Spec.Route53HostedZoneId = nil
			input.Spec.Imported = &AwsCertManagerCertImported{
				CertificateBody: "-----BEGIN CERTIFICATE-----",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Imported arm exclusions (imported_excludes_issuance_fields)", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.PrimaryDomainName = ""
			input.Spec.AlternateDomainNames = nil
			input.Spec.ValidationMethod = ""
			input.Spec.Route53HostedZoneId = nil
			input.Spec.Imported = &AwsCertManagerCertImported{
				CertificateBody: "-----BEGIN CERTIFICATE-----",
				PrivateKey:      "-----BEGIN PRIVATE KEY-----",
			}
		})

		ginkgo.It("should reject imported with a validation method", func() {
			input.Spec.ValidationMethod = "DNS"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject imported with a key algorithm", func() {
			input.Spec.KeyAlgorithm = "RSA_2048"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject imported with options", func() {
			input.Spec.Options = &AwsCertManagerCertOptions{
				CertificateTransparencyLoggingPreference: "DISABLED",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject imported with a Route53 zone", func() {
			input.Spec.Route53HostedZoneId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-zone-id"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Private CA arm (private_ca_excludes_validation)", func() {
		ginkgo.BeforeEach(func() {
			input.Spec.ValidationMethod = ""
			input.Spec.Route53HostedZoneId = nil
			input.Spec.CertificateAuthorityArn = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:acm-pca:us-west-2:123456789012:certificate-authority/abc-123"},
			}
		})

		ginkgo.It("should accept a private certificate", func() {
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a private certificate with a validation method", func() {
			input.Spec.ValidationMethod = "DNS"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a private certificate with a Route53 zone", func() {
			input.Spec.Route53HostedZoneId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-zone-id"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed CA ARN literal", func() {
			input.Spec.CertificateAuthorityArn = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "not-an-arn"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept a reference-shaped CA (the ARN pattern applies to literals only)", func() {
			input.Spec.CertificateAuthorityArn = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "internal-root-ca"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Validation method coupling (route53_zone_requires_dns_validation)", func() {
		ginkgo.It("should reject EMAIL validation with a Route53 zone", func() {
			input.Spec.ValidationMethod = "EMAIL"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept EMAIL validation without a zone", func() {
			input.Spec.ValidationMethod = "EMAIL"
			input.Spec.Route53HostedZoneId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept HTTP validation without a zone", func() {
			input.Spec.ValidationMethod = "HTTP"
			input.Spec.Route53HostedZoneId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject HTTP validation with a Route53 zone", func() {
			input.Spec.ValidationMethod = "HTTP"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown validation method", func() {
			input.Spec.ValidationMethod = "TXT"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Early renewal (early_renewal_requires_private_ca)", func() {
		privateCa := &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:acm-pca:us-west-2:123456789012:certificate-authority/abc-123"},
		}

		ginkgo.It("should accept an RFC 3339 duration on a private certificate", func() {
			input.Spec.ValidationMethod = ""
			input.Spec.Route53HostedZoneId = nil
			input.Spec.CertificateAuthorityArn = privateCa
			input.Spec.EarlyRenewalDuration = "P90D"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a Go-style duration on a private certificate", func() {
			input.Spec.ValidationMethod = ""
			input.Spec.Route53HostedZoneId = nil
			input.Spec.CertificateAuthorityArn = privateCa
			input.Spec.EarlyRenewalDuration = "2160h"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject early renewal on a publicly validated certificate (no CA set)", func() {
			input.Spec.EarlyRenewalDuration = "P90D"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed duration", func() {
			input.Spec.ValidationMethod = ""
			input.Spec.Route53HostedZoneId = nil
			input.Spec.CertificateAuthorityArn = privateCa
			input.Spec.EarlyRenewalDuration = "ninety days"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Key algorithm", func() {
		ginkgo.It("should accept a supported EC algorithm", func() {
			input.Spec.KeyAlgorithm = "EC_prime256v1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject an unsupported algorithm", func() {
			input.Spec.KeyAlgorithm = "RSA_1024"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Options", func() {
		ginkgo.It("should accept exportable certificates with CT logging disabled", func() {
			input.Spec.Options = &AwsCertManagerCertOptions{
				CertificateTransparencyLoggingPreference: "DISABLED",
				Export:                                   "ENABLED",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject an invalid export value", func() {
			input.Spec.Options = &AwsCertManagerCertOptions{Export: "MAYBE"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Validation options", func() {
		ginkgo.It("should accept a parent-domain validation option", func() {
			input.Spec.ValidationOptions = []*AwsCertManagerCertValidationOption{
				{DomainName: "app.example.com", ValidationDomain: "example.com"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should reject a validation option missing the validation domain", func() {
			input.Spec.ValidationOptions = []*AwsCertManagerCertValidationOption{
				{DomainName: "app.example.com"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
