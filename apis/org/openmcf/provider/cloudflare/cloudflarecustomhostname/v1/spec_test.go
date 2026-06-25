package cloudflarecustomhostnamev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"}}
}

func validCustomHostname() *CloudflareCustomHostname {
	return &CloudflareCustomHostname{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareCustomHostname",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-custom-hostname"},
		Spec: &CloudflareCustomHostnameSpec{
			ZoneId:   zoneRef(),
			Hostname: "support.acme.com",
		},
	}
}

func TestCloudflareCustomHostnameSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareCustomHostnameSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareCustomHostnameSpec validations", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal valid custom hostname", func() {
			gomega.Expect(protovalidate.Validate(validCustomHostname())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a full ssl block", func() {
			in := validCustomHostname()
			in.Spec.Ssl = &CloudflareCustomHostnameSsl{
				CertificateAuthority: "google",
				Method:               "txt",
				Settings: &CloudflareCustomHostnameSslSettings{
					MinTlsVersion: "1.2",
					Http2:         "on",
					Tls_1_3:       "on",
				},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a custom cert bundle", func() {
			in := validCustomHostname()
			in.Spec.Ssl = &CloudflareCustomHostnameSsl{
				CustomCertBundle: []*CloudflareCustomHostnameSslCustomCertBundle{
					{CustomCertificate: "-----BEGIN CERTIFICATE-----\nx\n-----END CERTIFICATE-----", CustomKey: "-----BEGIN PRIVATE KEY-----\ny\n-----END PRIVATE KEY-----"},
				},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing zone_id", func() {
			in := validCustomHostname()
			in.Spec.ZoneId = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing hostname", func() {
			in := validCustomHostname()
			in.Spec.Hostname = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid ssl.bundle_method", func() {
			in := validCustomHostname()
			bad := "best"
			in.Spec.Ssl = &CloudflareCustomHostnameSsl{BundleMethod: &bad}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid ssl.settings.min_tls_version", func() {
			in := validCustomHostname()
			in.Spec.Ssl = &CloudflareCustomHostnameSsl{
				Settings: &CloudflareCustomHostnameSslSettings{MinTlsVersion: "2.0"},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a cert bundle missing its key", func() {
			in := validCustomHostname()
			in.Spec.Ssl = &CloudflareCustomHostnameSsl{
				CustomCertBundle: []*CloudflareCustomHostnameSslCustomCertBundle{
					{CustomCertificate: "-----BEGIN CERTIFICATE-----\nx\n-----END CERTIFICATE-----"},
				},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
