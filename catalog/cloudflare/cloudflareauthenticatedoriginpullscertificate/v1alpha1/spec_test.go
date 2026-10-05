package cloudflareauthenticatedoriginpullscertificatev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareAuthenticatedOriginPullsCertificateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareAuthenticatedOriginPullsCertificateSpec Custom Validation Tests")
}

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "0da42c8d2132a9ddaf714f9e7c920711"},
	}
}

func keyRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----\n"},
	}
}

const testCertPem = "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----\n"

func validCertificate(spec *CloudflareAuthenticatedOriginPullsCertificateSpec) *CloudflareAuthenticatedOriginPullsCertificate {
	return &CloudflareAuthenticatedOriginPullsCertificate{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareAuthenticatedOriginPullsCertificate",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-aop-certificate",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareAuthenticatedOriginPullsCertificateSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a zone-scoped upload with the scope defaulted", func() {
			input := validCertificate(&CloudflareAuthenticatedOriginPullsCertificateSpec{
				ZoneId:      zoneRef(),
				Certificate: testCertPem,
				PrivateKey:  keyRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a hostname-scoped upload", func() {
			input := validCertificate(&CloudflareAuthenticatedOriginPullsCertificateSpec{
				ZoneId:      zoneRef(),
				Scope:       proto.String("hostname"),
				Certificate: testCertPem,
				PrivateKey:  keyRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject an unknown scope", func() {
			input := validCertificate(&CloudflareAuthenticatedOriginPullsCertificateSpec{
				ZoneId:      zoneRef(),
				Scope:       proto.String("account"),
				Certificate: testCertPem,
				PrivateKey:  keyRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing certificate", func() {
			input := validCertificate(&CloudflareAuthenticatedOriginPullsCertificateSpec{
				ZoneId:     zoneRef(),
				PrivateKey: keyRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing private key", func() {
			input := validCertificate(&CloudflareAuthenticatedOriginPullsCertificateSpec{
				ZoneId:      zoneRef(),
				Certificate: testCertPem,
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
