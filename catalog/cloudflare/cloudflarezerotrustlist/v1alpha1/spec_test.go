package cloudflarezerotrustlistv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestCloudflareZeroTrustListSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustListSpec Custom Validation Tests")
}

const testAccountId = "023e105f4ecef8ad9ca31a8372d0c353"

func validList(spec *CloudflareZeroTrustListSpec) *CloudflareZeroTrustList {
	return &CloudflareZeroTrustList{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustList",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-zt-list",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareZeroTrustListSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal empty DOMAIN list", func() {
			input := validList(&CloudflareZeroTrustListSpec{
				AccountId: testAccountId,
				Name:      "blocked-domains",
				Type:      "DOMAIN",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an IP list with items", func() {
			input := validList(&CloudflareZeroTrustListSpec{
				AccountId: testAccountId,
				Name:      "office-egress-ips",
				Type:      "IP",
				Items: []*CloudflareZeroTrustListItem{
					{Value: "203.0.113.10", Description: "office A"},
					{Value: "203.0.113.11"},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing account_id", func() {
			input := validList(&CloudflareZeroTrustListSpec{
				Name: "blocked-domains",
				Type: "DOMAIN",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a non-hex account_id", func() {
			input := validList(&CloudflareZeroTrustListSpec{
				AccountId: "not-a-hex-id",
				Name:      "blocked-domains",
				Type:      "DOMAIN",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a lowercase list type (the API stores uppercase)", func() {
			input := validList(&CloudflareZeroTrustListSpec{
				AccountId: testAccountId,
				Name:      "blocked-domains",
				Type:      "domain",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown list type", func() {
			input := validList(&CloudflareZeroTrustListSpec{
				AccountId: testAccountId,
				Name:      "blocked-domains",
				Type:      "HOSTNAME",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an item without a value", func() {
			input := validList(&CloudflareZeroTrustListSpec{
				AccountId: testAccountId,
				Name:      "office-egress-ips",
				Type:      "IP",
				Items: []*CloudflareZeroTrustListItem{
					{Description: "an entry that matches nothing"},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
