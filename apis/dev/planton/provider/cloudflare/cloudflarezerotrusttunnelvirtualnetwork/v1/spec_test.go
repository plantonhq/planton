package cloudflarezerotrusttunnelvirtualnetworkv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func validVirtualNetwork() *CloudflareZeroTrustTunnelVirtualNetwork {
	return &CloudflareZeroTrustTunnelVirtualNetwork{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareZeroTrustTunnelVirtualNetwork",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-vnet"},
		Spec: &CloudflareZeroTrustTunnelVirtualNetworkSpec{
			AccountId: validAccountID,
			Name:      "prod-vnet",
		},
	}
}

func TestCloudflareZeroTrustTunnelVirtualNetworkSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustTunnelVirtualNetworkSpec Validation Suite")
}

var _ = ginkgo.Describe("CloudflareZeroTrustTunnelVirtualNetworkSpec Validation", func() {
	ginkgo.Describe("Valid inputs", func() {
		ginkgo.It("accepts a minimal virtual network", func() {
			gomega.Expect(protovalidate.Validate(validVirtualNetwork())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a comment and default flag", func() {
			vn := validVirtualNetwork()
			vn.Spec.Comment = "isolates the staging 10.0.0.0/8 overlap"
			isDefault := true
			vn.Spec.IsDefaultNetwork = &isDefault
			gomega.Expect(protovalidate.Validate(vn)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Invalid inputs", func() {
		ginkgo.It("rejects a non-hex account_id", func() {
			vn := validVirtualNetwork()
			vn.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(vn)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a short account_id", func() {
			vn := validVirtualNetwork()
			vn.Spec.AccountId = "0a1b2c3d"
			gomega.Expect(protovalidate.Validate(vn)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing name", func() {
			vn := validVirtualNetwork()
			vn.Spec.Name = ""
			gomega.Expect(protovalidate.Validate(vn)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a name beyond the maximum length", func() {
			vn := validVirtualNetwork()
			vn.Spec.Name = strings.Repeat("a", 101)
			gomega.Expect(protovalidate.Validate(vn)).ToNot(gomega.BeNil())
		})
	})
})
