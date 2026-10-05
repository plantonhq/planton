package cloudflaremtlscertificatev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareMtlsCertificateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareMtlsCertificateSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

const testCaPem = "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----\n"

func keyRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "-----BEGIN PRIVATE KEY-----\ntest\n-----END PRIVATE KEY-----\n"},
	}
}

func validCertificate(spec *CloudflareMtlsCertificateSpec) *CloudflareMtlsCertificate {
	return &CloudflareMtlsCertificate{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareMtlsCertificate",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-mtls-certificate",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareMtlsCertificateSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a CA upload without a private key", func() {
			input := validCertificate(&CloudflareMtlsCertificateSpec{
				AccountId:    testAccountID,
				Name:         "origin-pull-ca",
				Ca:           proto.Bool(true),
				Certificates: testCaPem,
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a leaf upload carrying its private key", func() {
			input := validCertificate(&CloudflareMtlsCertificateSpec{
				AccountId:    testAccountID,
				Ca:           proto.Bool(false),
				Certificates: testCaPem,
				PrivateKey:   keyRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject an unset ca flag -- it must be stated explicitly", func() {
			input := validCertificate(&CloudflareMtlsCertificateSpec{
				AccountId:    testAccountID,
				Certificates: testCaPem,
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject missing certificates", func() {
			input := validCertificate(&CloudflareMtlsCertificateSpec{
				AccountId: testAccountID,
				Ca:        proto.Bool(true),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed account id", func() {
			input := validCertificate(&CloudflareMtlsCertificateSpec{
				AccountId:    "not-a-hex-account-id",
				Ca:           proto.Bool(true),
				Certificates: testCaPem,
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
