package cloudflarewebanalyticssitev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareWebAnalyticsSiteSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareWebAnalyticsSiteSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func zoneRef(zoneId string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: zoneId},
	}
}

func validSite(spec *CloudflareWebAnalyticsSiteSpec) *CloudflareWebAnalyticsSite {
	return &CloudflareWebAnalyticsSite{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareWebAnalyticsSite",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-web-analytics-site",
		},
		Spec: spec,
	}
}

func hostSpec() *CloudflareWebAnalyticsSiteSpec {
	return &CloudflareWebAnalyticsSiteSpec{
		AccountId: testAccountID,
		Host:      "www.example.com",
	}
}

var _ = ginkgo.Describe("CloudflareWebAnalyticsSiteSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a host-identified site", func() {
			gomega.Expect(protovalidate.Validate(validSite(hostSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a zone-identified site with auto install", func() {
			spec := &CloudflareWebAnalyticsSiteSpec{
				AccountId:   testAccountID,
				ZoneTag:     zoneRef("023e105f4ecef8ad9ca31a8372d0c353"),
				AutoInstall: proto.Bool(true),
			}
			gomega.Expect(protovalidate.Validate(validSite(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept measurement rules on a zone-identified site", func() {
			// Rules require zone_tag: Cloudflare only creates the ruleset
			// (the container rules attach to) for zone-linked sites.
			spec := &CloudflareWebAnalyticsSiteSpec{
				AccountId: testAccountID,
				ZoneTag:   zoneRef("023e105f4ecef8ad9ca31a8372d0c353"),
				Lite:      proto.Bool(true),
			}
			spec.Rules = []*CloudflareWebAnalyticsSiteRule{
				{Host: "www.example.com", Paths: []string{"/checkout/*"}, Inclusive: proto.Bool(false)},
				{Paths: []string{"/*"}, Inclusive: proto.Bool(true)},
			}
			gomega.Expect(protovalidate.Validate(validSite(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject both host and zone_tag set", func() {
			spec := hostSpec()
			spec.ZoneTag = zoneRef("023e105f4ecef8ad9ca31a8372d0c353")
			gomega.Expect(protovalidate.Validate(validSite(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject neither host nor zone_tag set", func() {
			spec := &CloudflareWebAnalyticsSiteSpec{AccountId: testAccountID}
			gomega.Expect(protovalidate.Validate(validSite(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject rules on a host-identified site", func() {
			// Measured live: host-identified sites have NO ruleset in any
			// API response, so rules have nothing to attach to.
			spec := hostSpec()
			spec.Rules = []*CloudflareWebAnalyticsSiteRule{
				{Paths: []string{"/*"}, Inclusive: proto.Bool(true)},
			}
			gomega.Expect(protovalidate.Validate(validSite(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed account id", func() {
			spec := hostSpec()
			spec.AccountId = "zz"
			gomega.Expect(protovalidate.Validate(validSite(spec))).NotTo(gomega.BeNil())
		})
	})
})
