package cloudflarezerotrustdnslocationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareZeroTrustDnsLocationSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustDnsLocationSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func validLocation(spec *CloudflareZeroTrustDnsLocationSpec) *CloudflareZeroTrustDnsLocation {
	return &CloudflareZeroTrustDnsLocation{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustDnsLocation",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-location",
		},
		Spec: spec,
	}
}

func baseSpec() *CloudflareZeroTrustDnsLocationSpec {
	return &CloudflareZeroTrustDnsLocationSpec{
		AccountId: testAccountID,
		Name:      "hq-office",
	}
}

func allEndpoints() *CloudflareZeroTrustDnsLocationEndpoints {
	return &CloudflareZeroTrustDnsLocationEndpoints{
		Doh: &CloudflareZeroTrustDnsLocationDohEndpoint{
			Enabled:      proto.Bool(true),
			RequireToken: proto.Bool(true),
		},
		Dot:  &CloudflareZeroTrustDnsLocationNetworkEndpoint{Enabled: proto.Bool(false)},
		Ipv4: &CloudflareZeroTrustDnsLocationIpv4Endpoint{Enabled: proto.Bool(true)},
		Ipv6: &CloudflareZeroTrustDnsLocationNetworkEndpoint{Enabled: proto.Bool(false)},
	}
}

var _ = ginkgo.Describe("CloudflareZeroTrustDnsLocationSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal location", func() {
			gomega.Expect(protovalidate.Validate(validLocation(baseSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a full endpoints tree", func() {
			spec := baseSpec()
			spec.Endpoints = allEndpoints()
			gomega.Expect(protovalidate.Validate(validLocation(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept source networks", func() {
			spec := baseSpec()
			spec.Networks = []*CloudflareZeroTrustDnsLocationNetwork{
				{Network: "203.0.113.0/24"},
			}
			gomega.Expect(protovalidate.Validate(validLocation(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept max_ttl inherit without a ttl", func() {
			spec := baseSpec()
			spec.MaxTtl = &CloudflareZeroTrustDnsLocationMaxTtl{Mode: "inherit"}
			gomega.Expect(protovalidate.Validate(validLocation(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept max_ttl override with a ttl", func() {
			spec := baseSpec()
			spec.MaxTtl = &CloudflareZeroTrustDnsLocationMaxTtl{
				Mode:    "override",
				TtlSecs: proto.Int64(300),
			}
			gomega.Expect(protovalidate.Validate(validLocation(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject an endpoints tree missing a type -- Cloudflare takes all four at once", func() {
			spec := baseSpec()
			spec.Endpoints = allEndpoints()
			spec.Endpoints.Dot = nil
			gomega.Expect(protovalidate.Validate(validLocation(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject max_ttl override without a ttl", func() {
			spec := baseSpec()
			spec.MaxTtl = &CloudflareZeroTrustDnsLocationMaxTtl{Mode: "override"}
			gomega.Expect(protovalidate.Validate(validLocation(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject max_ttl inherit with a ttl", func() {
			spec := baseSpec()
			spec.MaxTtl = &CloudflareZeroTrustDnsLocationMaxTtl{
				Mode:    "inherit",
				TtlSecs: proto.Int64(300),
			}
			gomega.Expect(protovalidate.Validate(validLocation(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown max_ttl mode", func() {
			spec := baseSpec()
			spec.MaxTtl = &CloudflareZeroTrustDnsLocationMaxTtl{Mode: "cap"}
			gomega.Expect(protovalidate.Validate(validLocation(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a ttl below 60 seconds", func() {
			spec := baseSpec()
			spec.MaxTtl = &CloudflareZeroTrustDnsLocationMaxTtl{
				Mode:    "override",
				TtlSecs: proto.Int64(59),
			}
			gomega.Expect(protovalidate.Validate(validLocation(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty network row", func() {
			spec := baseSpec()
			spec.Networks = []*CloudflareZeroTrustDnsLocationNetwork{{}}
			gomega.Expect(protovalidate.Validate(validLocation(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing name", func() {
			spec := baseSpec()
			spec.Name = ""
			gomega.Expect(protovalidate.Validate(validLocation(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed account_id", func() {
			spec := baseSpec()
			spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(validLocation(spec))).NotTo(gomega.BeNil())
		})
	})
})
