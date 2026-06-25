package cloudflareloadbalancerpoolv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

const validAccountID = "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"

func ref(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validPool() *CloudflareLoadBalancerPool {
	return &CloudflareLoadBalancerPool{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareLoadBalancerPool",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-pool"},
		Spec: &CloudflareLoadBalancerPoolSpec{
			AccountId: validAccountID,
			Name:      "web-pool",
			Origins: []*CloudflareLoadBalancerPoolOrigin{
				{Name: "origin-1", Address: ref("203.0.113.10")},
			},
			Monitor: ref("a1b2c3d4e5f60718293a4b5c6d7e8f90"),
		},
	}
}

func TestCloudflareLoadBalancerPoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareLoadBalancerPoolSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareLoadBalancerPoolSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal pool", func() {
			gomega.Expect(protovalidate.Validate(validPool())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a rich pool with steering, shedding, regions, and geo", func() {
			in := validPool()
			lat, lon := 37.77, -122.42
			weight := 0.5
			enabled := true
			in.Spec.Origins[0].Weight = &weight
			in.Spec.Origins[0].Enabled = &enabled
			in.Spec.Origins[0].Port = 8443
			in.Spec.Origins[0].HostHeader = "app.internal"
			in.Spec.CheckRegions = []CloudflareLoadBalancerPoolCheckRegion{
				CloudflareLoadBalancerPoolCheckRegion_WNAM,
				CloudflareLoadBalancerPoolCheckRegion_WEU,
			}
			in.Spec.MinimumOrigins = 1
			in.Spec.Latitude = &lat
			in.Spec.Longitude = &lon
			in.Spec.LoadShedding = &CloudflareLoadBalancerPoolLoadShedding{
				DefaultPercent: 10, DefaultPolicy: "random", SessionPercent: 5, SessionPolicy: "hash",
			}
			in.Spec.OriginSteering = &CloudflareLoadBalancerPoolOriginSteering{Policy: "least_connections"}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a non-hex account_id", func() {
			in := validPool()
			in.Spec.AccountId = "not-valid"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a name with illegal characters", func() {
			in := validPool()
			in.Spec.Name = "web pool!"
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a pool with no origins", func() {
			in := validPool()
			in.Spec.Origins = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an origin missing its address", func() {
			in := validPool()
			in.Spec.Origins[0].Address = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an origin weight above 1", func() {
			in := validPool()
			w := 1.5
			in.Spec.Origins[0].Weight = &w
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid origin-steering policy", func() {
			in := validPool()
			in.Spec.OriginSteering = &CloudflareLoadBalancerPoolOriginSteering{Policy: "round_robin"}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a load-shedding session_policy other than hash", func() {
			in := validPool()
			in.Spec.LoadShedding = &CloudflareLoadBalancerPoolLoadShedding{SessionPolicy: "random"}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an out-of-range latitude", func() {
			in := validPool()
			lat := 200.0
			in.Spec.Latitude = &lat
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
