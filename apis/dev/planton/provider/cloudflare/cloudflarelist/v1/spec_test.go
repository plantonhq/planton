package cloudflarelistv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func validList() *CloudflareList {
	return &CloudflareList{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareList",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-list"},
		Spec: &CloudflareListSpec{
			AccountId: validAccountID,
			Kind:      CloudflareListKind_ip,
			Name:      "office_allowlist",
		},
	}
}

func TestCloudflareListSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareListSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareListSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal valid list", func() {
			gomega.Expect(protovalidate.Validate(validList())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a description and each list kind", func() {
			for _, k := range []CloudflareListKind{
				CloudflareListKind_ip,
				CloudflareListKind_redirect,
				CloudflareListKind_hostname,
				CloudflareListKind_asn,
			} {
				in := validList()
				in.Spec.Kind = k
				in.Spec.Description = "curated set used by edge rules"
				gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
			}
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a non-hex account_id", func() {
			in := validList()
			in.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unspecified kind", func() {
			in := validList()
			in.Spec.Kind = CloudflareListKind_list_kind_unspecified
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing name", func() {
			in := validList()
			in.Spec.Name = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a name that does not start with a letter", func() {
			in := validList()
			in.Spec.Name = "1bad"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a name with disallowed characters", func() {
			in := validList()
			in.Spec.Name = "bad-name"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a name over 50 characters", func() {
			in := validList()
			in.Spec.Name = strings.Repeat("a", 51)
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
