package cloudflareorigincacertificatev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func validOriginCaCertificate() *CloudflareOriginCaCertificate {
	return &CloudflareOriginCaCertificate{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareOriginCaCertificate",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-origin-cert"},
		Spec: &CloudflareOriginCaCertificateSpec{
			Hostnames: []string{"example.com", "*.example.com"},
		},
	}
}

func TestCloudflareOriginCaCertificateSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareOriginCaCertificateSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareOriginCaCertificateSpec validations", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal valid certificate (defaults applied by middleware)", func() {
			gomega.Expect(protovalidate.Validate(validOriginCaCertificate())).To(gomega.BeNil())
		})

		ginkgo.It("accepts each request_type value", func() {
			for _, rt := range []string{"origin-rsa", "origin-ecc", "keyless-certificate"} {
				in := validOriginCaCertificate()
				in.Spec.RequestType = &rt
				gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
			}
		})

		ginkgo.It("accepts each requested_validity value", func() {
			for _, v := range []int64{7, 30, 90, 365, 730, 1095, 5475} {
				in := validOriginCaCertificate()
				in.Spec.RequestedValidity = &v
				gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
			}
		})

		ginkgo.It("accepts a user-supplied CSR", func() {
			in := validOriginCaCertificate()
			in.Spec.Csr = "-----BEGIN CERTIFICATE REQUEST-----\nMIIBfake\n-----END CERTIFICATE REQUEST-----"
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects an empty hostnames list", func() {
			in := validOriginCaCertificate()
			in.Spec.Hostnames = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty-string hostname", func() {
			in := validOriginCaCertificate()
			in.Spec.Hostnames = []string{""}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid request_type", func() {
			in := validOriginCaCertificate()
			bad := "origin-dsa"
			in.Spec.RequestType = &bad
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid requested_validity", func() {
			in := validOriginCaCertificate()
			var bad int64 = 42
			in.Spec.RequestedValidity = &bad
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
