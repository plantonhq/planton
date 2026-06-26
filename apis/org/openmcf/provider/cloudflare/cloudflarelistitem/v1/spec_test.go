package cloudflarelistitemv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func value(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func boolPtr(b bool) *bool { return &b }

func baseItem() *CloudflareListItem {
	return &CloudflareListItem{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareListItem",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-list-item"},
		Spec: &CloudflareListItemSpec{
			AccountId: validAccountID,
			ListId:    value("2c0fc9fa937b11eaa1b71c4d701ab86e"),
		},
	}
}

func ipItem() *CloudflareListItem {
	in := baseItem()
	in.Spec.Item = &CloudflareListItemSpec_Ip{Ip: "203.0.113.0/24"}
	return in
}

func TestCloudflareListItemSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareListItemSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareListItemSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts an ip item", func() {
			gomega.Expect(protovalidate.Validate(ipItem())).To(gomega.BeNil())
		})

		ginkgo.It("accepts an asn item", func() {
			in := baseItem()
			in.Spec.Item = &CloudflareListItemSpec_Asn{Asn: 13335}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a non-wildcard hostname without exclude_exact_hostname", func() {
			in := baseItem()
			in.Spec.Item = &CloudflareListItemSpec_Hostname{
				Hostname: &CloudflareListItemHostname{UrlHostname: "api.example.com"},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a wildcard hostname with exclude_exact_hostname set", func() {
			in := baseItem()
			in.Spec.Item = &CloudflareListItemSpec_Hostname{
				Hostname: &CloudflareListItemHostname{UrlHostname: "*.example.com", ExcludeExactHostname: boolPtr(true)},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a redirect item", func() {
			in := baseItem()
			in.Spec.Item = &CloudflareListItemSpec_Redirect{
				Redirect: &CloudflareListItemRedirect{SourceUrl: "example.com/old", TargetUrl: "https://example.com/new", StatusCode: 301},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a comment", func() {
			in := ipItem()
			in.Spec.Comment = "datacenter egress"
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects when no item value is set", func() {
			in := baseItem()
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a non-hex account_id", func() {
			in := ipItem()
			in.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing list_id", func() {
			in := ipItem()
			in.Spec.ListId = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a wildcard hostname missing exclude_exact_hostname", func() {
			in := baseItem()
			in.Spec.Item = &CloudflareListItemSpec_Hostname{
				Hostname: &CloudflareListItemHostname{UrlHostname: "*.example.com"},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a non-wildcard hostname that sets exclude_exact_hostname", func() {
			in := baseItem()
			in.Spec.Item = &CloudflareListItemSpec_Hostname{
				Hostname: &CloudflareListItemHostname{UrlHostname: "api.example.com", ExcludeExactHostname: boolPtr(false)},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid redirect status_code", func() {
			in := baseItem()
			in.Spec.Item = &CloudflareListItemSpec_Redirect{
				Redirect: &CloudflareListItemRedirect{SourceUrl: "example.com/old", TargetUrl: "https://example.com/new", StatusCode: 418},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an asn above the 32-bit range", func() {
			in := baseItem()
			in.Spec.Item = &CloudflareListItemSpec_Asn{Asn: 4294967296}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
