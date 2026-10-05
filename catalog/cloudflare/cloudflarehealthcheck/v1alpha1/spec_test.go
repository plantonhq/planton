package cloudflarehealthcheckv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareHealthcheckSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareHealthcheckSpec Custom Validation Tests")
}

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "0da42c8d2132a9ddaf714f9e7c920711"},
	}
}

func validHealthcheck(spec *CloudflareHealthcheckSpec) *CloudflareHealthcheck {
	return &CloudflareHealthcheck{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareHealthcheck",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-healthcheck",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareHealthcheckSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal HTTP check", func() {
			input := validHealthcheck(&CloudflareHealthcheckSpec{
				ZoneId:  zoneRef(),
				Name:    "origin-http",
				Address: "origin.example.com",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an HTTPS check with full http_config", func() {
			input := validHealthcheck(&CloudflareHealthcheckSpec{
				ZoneId:               zoneRef(),
				Name:                 "origin-https",
				Address:              "origin.example.com",
				Type:                 proto.String("HTTPS"),
				CheckRegions:         []string{"WEU", "ENAM"},
				ConsecutiveFails:     proto.Int32(2),
				ConsecutiveSuccesses: proto.Int32(2),
				Interval:             proto.Int32(60),
				Retries:              proto.Int32(2),
				Timeout:              proto.Int32(5),
				HttpConfig: &CloudflareHealthcheckHttpConfig{
					Method:          proto.String("GET"),
					Path:            proto.String("/healthz"),
					Port:            proto.Int32(443),
					ExpectedCodes:   []string{"200", "2xx"},
					ExpectedBody:    "ok",
					FollowRedirects: proto.Bool(true),
					AllowInsecure:   proto.Bool(false),
					Headers: map[string]*CloudflareHealthcheckHeaderValues{
						"Host": {Values: []string{"origin.example.com"}},
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a TCP check with tcp_config", func() {
			input := validHealthcheck(&CloudflareHealthcheckSpec{
				ZoneId:  zoneRef(),
				Name:    "origin-tcp",
				Address: "203.0.113.10",
				Type:    proto.String("TCP"),
				TcpConfig: &CloudflareHealthcheckTcpConfig{
					Port: proto.Int32(5432),
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject tcp_config on the default (HTTP) type", func() {
			input := validHealthcheck(&CloudflareHealthcheckSpec{
				ZoneId:    zoneRef(),
				Name:      "origin-http",
				Address:   "origin.example.com",
				TcpConfig: &CloudflareHealthcheckTcpConfig{},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject http_config on a TCP check", func() {
			input := validHealthcheck(&CloudflareHealthcheckSpec{
				ZoneId:     zoneRef(),
				Name:       "origin-tcp",
				Address:    "203.0.113.10",
				Type:       proto.String("TCP"),
				HttpConfig: &CloudflareHealthcheckHttpConfig{},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown type", func() {
			input := validHealthcheck(&CloudflareHealthcheckSpec{
				ZoneId:  zoneRef(),
				Name:    "origin",
				Address: "origin.example.com",
				Type:    proto.String("UDP"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown check region", func() {
			input := validHealthcheck(&CloudflareHealthcheckSpec{
				ZoneId:       zoneRef(),
				Name:         "origin",
				Address:      "origin.example.com",
				CheckRegions: []string{"EU"},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a header with no values", func() {
			input := validHealthcheck(&CloudflareHealthcheckSpec{
				ZoneId:  zoneRef(),
				Name:    "origin",
				Address: "origin.example.com",
				HttpConfig: &CloudflareHealthcheckHttpConfig{
					Headers: map[string]*CloudflareHealthcheckHeaderValues{
						"Host": {},
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an out-of-range port", func() {
			input := validHealthcheck(&CloudflareHealthcheckSpec{
				ZoneId:  zoneRef(),
				Name:    "origin",
				Address: "origin.example.com",
				HttpConfig: &CloudflareHealthcheckHttpConfig{
					Port: proto.Int32(70000),
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing address", func() {
			input := validHealthcheck(&CloudflareHealthcheckSpec{
				ZoneId: zoneRef(),
				Name:   "origin",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
