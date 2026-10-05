package cloudflareipaccessrulev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareIpAccessRuleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareIpAccessRuleSpec Custom Validation Tests")
}

const testAccountId = "023e105f4ecef8ad9ca31a8372d0c353"

func zoneRef(zoneId string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: zoneId},
	}
}

func validRule(spec *CloudflareIpAccessRuleSpec) *CloudflareIpAccessRule {
	return &CloudflareIpAccessRule{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareIpAccessRule",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ip-access-rule",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareIpAccessRuleSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept an account-wide IP block", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				Mode:      "block",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "ip",
					Value:  "203.0.113.7",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a zone-scoped country challenge", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				ZoneId: zoneRef("0da42c8d2132a9ddaf714f9e7c920711"),
				Mode:   "managed_challenge",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "country",
					Value:  "US",
				},
				Notes: "challenge all US traffic",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a long-form IPv6 rule", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				Mode:      "whitelist",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "ip6",
					Value:  "2001:0db8:0000:0000:0000:0000:0000:0001",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a /24 IPv4 range and a /64 IPv6 range", func() {
			v4 := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				Mode:      "js_challenge",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "ip_range",
					Value:  "203.0.113.0/24",
				},
			})
			gomega.Expect(protovalidate.Validate(v4)).To(gomega.BeNil())

			v6 := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				Mode:      "js_challenge",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "ip_range",
					Value:  "2001:db8::/64",
				},
			})
			gomega.Expect(protovalidate.Validate(v6)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an ASN rule", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				Mode:      "challenge",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "asn",
					Value:  "AS13335",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject both scopes set", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				ZoneId:    zoneRef("0da42c8d2132a9ddaf714f9e7c920711"),
				Mode:      "block",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "ip",
					Value:  "203.0.113.7",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject neither scope set", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				Mode: "block",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "ip",
					Value:  "203.0.113.7",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown mode", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				Mode:      "allow",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "ip",
					Value:  "203.0.113.7",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a compressed IPv6 value on target ip6", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				Mode:      "block",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "ip6",
					Value:  "2001:db8::1",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an IPv4 range with a disallowed prefix", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				Mode:      "block",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "ip_range",
					Value:  "203.0.113.0/25",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an IPv6 range with a disallowed prefix", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				Mode:      "block",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "ip_range",
					Value:  "2001:db8::/66",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a lowercase country code", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				Mode:      "block",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "country",
					Value:  "us",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a bare AS number without the AS prefix", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				Mode:      "block",
				Configuration: &CloudflareIpAccessRuleConfiguration{
					Target: "asn",
					Value:  "13335",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing configuration", func() {
			input := validRule(&CloudflareIpAccessRuleSpec{
				AccountId: testAccountId,
				Mode:      "block",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
