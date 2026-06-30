package cloudflarezerotrusttunnelroutev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func value(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validRoute() *CloudflareZeroTrustTunnelRoute {
	return &CloudflareZeroTrustTunnelRoute{
		ApiVersion: "cloudflare.planton.dev/v1",
		Kind:       "CloudflareZeroTrustTunnelRoute",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-route"},
		Spec: &CloudflareZeroTrustTunnelRouteSpec{
			AccountId: validAccountID,
			Network:   "10.0.0.0/24",
			TunnelId:  value("b8f2e1c0-1111-2222-3333-444455556666"),
		},
	}
}

func TestCloudflareZeroTrustTunnelRouteSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustTunnelRouteSpec Validation Suite")
}

var _ = ginkgo.Describe("CloudflareZeroTrustTunnelRouteSpec Validation", func() {
	ginkgo.Describe("Valid inputs", func() {
		ginkgo.It("accepts a minimal IPv4 route", func() {
			gomega.Expect(protovalidate.Validate(validRoute())).To(gomega.BeNil())
		})

		ginkgo.It("accepts an IPv6 route within a referenced virtual network", func() {
			r := validRoute()
			r.Spec.Network = "2001:db8::/48"
			r.Spec.VirtualNetworkId = value("aaaa1111-bbbb-2222-cccc-333344445555")
			r.Spec.Comment = "ipv6 segment"
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a single-host /32 route", func() {
			r := validRoute()
			r.Spec.Network = "10.0.0.5/32"
			gomega.Expect(protovalidate.Validate(r)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Invalid inputs", func() {
		ginkgo.It("rejects a non-hex account_id", func() {
			r := validRoute()
			r.Spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(r)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a network without a prefix length", func() {
			r := validRoute()
			r.Spec.Network = "10.0.0.0"
			gomega.Expect(protovalidate.Validate(r)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing network", func() {
			r := validRoute()
			r.Spec.Network = ""
			gomega.Expect(protovalidate.Validate(r)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing tunnel_id", func() {
			r := validRoute()
			r.Spec.TunnelId = nil
			gomega.Expect(protovalidate.Validate(r)).ToNot(gomega.BeNil())
		})
	})
})
