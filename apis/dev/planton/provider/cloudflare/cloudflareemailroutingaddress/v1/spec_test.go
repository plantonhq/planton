package cloudflareemailroutingaddressv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func validAddress() *CloudflareEmailRoutingAddress {
	return &CloudflareEmailRoutingAddress{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareEmailRoutingAddress",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-address"},
		Spec: &CloudflareEmailRoutingAddressSpec{
			AccountId: validAccountID,
			Email:     "ops@example.com",
		},
	}
}

func TestCloudflareEmailRoutingAddressSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareEmailRoutingAddressSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareEmailRoutingAddressSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a valid destination address", func() {
			gomega.Expect(protovalidate.Validate(validAddress())).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a non-hex account_id", func() {
			in := validAddress()
			in.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing email", func() {
			in := validAddress()
			in.Spec.Email = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a malformed email", func() {
			in := validAddress()
			in.Spec.Email = "not-an-email"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
