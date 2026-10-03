package cloudflarecustomsslcertificatev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareCustomSslCertificateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareCustomSslCertificateSpec Custom Validation Tests")
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

func validCertificate(spec *CloudflareCustomSslCertificateSpec) *CloudflareCustomSslCertificate {
	return &CloudflareCustomSslCertificate{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareCustomSslCertificate",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-custom-ssl",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareCustomSslCertificateSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal upload", func() {
			input := validCertificate(&CloudflareCustomSslCertificateSpec{
				ZoneId:      zoneRef(),
				Certificate: testCertPem,
				PrivateKey:  keyRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a full sni_custom upload with geo restrictions and staging deploy", func() {
			input := validCertificate(&CloudflareCustomSslCertificateSpec{
				ZoneId:       zoneRef(),
				Certificate:  testCertPem,
				PrivateKey:   keyRef(),
				Type:         proto.String("sni_custom"),
				BundleMethod: proto.String("optimal"),
				GeoRestrictions: &CloudflareCustomSslCertificateGeoRestrictions{
					Label: proto.String("highest_security"),
				},
				Deploy: proto.String("staging"),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a policy expression alongside geo restrictions", func() {
			// The provider has NO conflict rule between policy and
			// geo_restrictions at v5.23.0 -- both are accepted together.
			input := validCertificate(&CloudflareCustomSslCertificateSpec{
				ZoneId:      zoneRef(),
				Certificate: testCertPem,
				PrivateKey:  keyRef(),
				Policy:      "(country: US) or (region: EU)",
				GeoRestrictions: &CloudflareCustomSslCertificateGeoRestrictions{
					Label: proto.String("us"),
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing certificate", func() {
			input := validCertificate(&CloudflareCustomSslCertificateSpec{
				ZoneId:     zoneRef(),
				PrivateKey: keyRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing private key", func() {
			input := validCertificate(&CloudflareCustomSslCertificateSpec{
				ZoneId:      zoneRef(),
				Certificate: testCertPem,
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown type", func() {
			input := validCertificate(&CloudflareCustomSslCertificateSpec{
				ZoneId:      zoneRef(),
				Certificate: testCertPem,
				PrivateKey:  keyRef(),
				Type:        proto.String("dedicated"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown bundle method", func() {
			input := validCertificate(&CloudflareCustomSslCertificateSpec{
				ZoneId:       zoneRef(),
				Certificate:  testCertPem,
				PrivateKey:   keyRef(),
				BundleMethod: proto.String("compat"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown geo restriction label", func() {
			input := validCertificate(&CloudflareCustomSslCertificateSpec{
				ZoneId:      zoneRef(),
				Certificate: testCertPem,
				PrivateKey:  keyRef(),
				GeoRestrictions: &CloudflareCustomSslCertificateGeoRestrictions{
					Label: proto.String("apac"),
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown deploy target", func() {
			input := validCertificate(&CloudflareCustomSslCertificateSpec{
				ZoneId:      zoneRef(),
				Certificate: testCertPem,
				PrivateKey:  keyRef(),
				Deploy:      proto.String("canary"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
