package cloudflarecertificatepackv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func validCertificatePack() *CloudflareCertificatePack {
	return &CloudflareCertificatePack{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareCertificatePack",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-cert-pack"},
		Spec: &CloudflareCertificatePackSpec{
			ZoneId:               &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"}},
			CertificateAuthority: "google",
			ValidationMethod:     "txt",
			ValidityDays:         90,
			Hosts:                []string{"example.com", "*.example.com"},
		},
	}
}

func TestCloudflareCertificatePackSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareCertificatePackSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareCertificatePackSpec validations", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal valid certificate pack", func() {
			gomega.Expect(protovalidate.Validate(validCertificatePack())).To(gomega.BeNil())
		})

		ginkgo.It("accepts each certificate_authority value", func() {
			for _, ca := range []string{"google", "lets_encrypt", "ssl_com"} {
				in := validCertificatePack()
				in.Spec.CertificateAuthority = ca
				gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
			}
		})

		ginkgo.It("accepts each validation_method value", func() {
			for _, m := range []string{"txt", "http", "email"} {
				in := validCertificatePack()
				in.Spec.ValidationMethod = m
				gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
			}
		})

		ginkgo.It("accepts each validity_days value", func() {
			for _, d := range []int64{14, 30, 90, 365} {
				in := validCertificatePack()
				in.Spec.ValidityDays = d
				gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
			}
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing zone_id", func() {
			in := validCertificatePack()
			in.Spec.ZoneId = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid certificate_authority", func() {
			in := validCertificatePack()
			in.Spec.CertificateAuthority = "digicert"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid validation_method", func() {
			in := validCertificatePack()
			in.Spec.ValidationMethod = "dns"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid validity_days", func() {
			in := validCertificatePack()
			in.Spec.ValidityDays = 60
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty hosts list", func() {
			in := validCertificatePack()
			in.Spec.Hosts = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
