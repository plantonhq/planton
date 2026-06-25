package cloudflareloadbalancerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func ref(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validLoadBalancer() *CloudflareLoadBalancer {
	return &CloudflareLoadBalancer{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareLoadBalancer",
		Metadata:   &shared.CloudResourceMetadata{Name: "test-load-balancer"},
		Spec: &CloudflareLoadBalancerSpec{
			Hostname:     "lb.example.com",
			ZoneId:       ref("023e105f4ecef8ad9ca31a8372d0c353"),
			DefaultPools: []*foreignkeyv1.StringValueOrRef{ref("pool-primary")},
			FallbackPool: ref("pool-fallback"),
		},
	}
}

func TestCloudflareLoadBalancerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareLoadBalancerSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CloudflareLoadBalancerSpec Custom Validation Tests", func() {
	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal load balancer", func() {
			gomega.Expect(protovalidate.Validate(validLoadBalancer())).To(gomega.BeNil())
		})

		ginkgo.It("accepts geo steering with region/country pools and affinity", func() {
			in := validLoadBalancer()
			in.Spec.SteeringPolicy = CloudflareLoadBalancerSteeringPolicy_geo
			in.Spec.SessionAffinity = CloudflareLoadBalancerSessionAffinity_header
			in.Spec.SessionAffinityTtl = 1800
			in.Spec.RegionPools = []*CloudflareLoadBalancerGeoPools{
				{Code: "WNAM", PoolIds: []*foreignkeyv1.StringValueOrRef{ref("pool-west")}},
			}
			in.Spec.CountryPools = []*CloudflareLoadBalancerGeoPools{
				{Code: "US", PoolIds: []*foreignkeyv1.StringValueOrRef{ref("pool-us")}},
			}
			in.Spec.SessionAffinityAttributes = &CloudflareLoadBalancerSessionAffinityAttributes{
				Headers: []string{"X-Session"}, RequireAllHeaders: true, Samesite: "Lax", Secure: "Always", ZeroDowntimeFailover: "sticky",
			}
			in.Spec.LocationStrategy = &CloudflareLoadBalancerLocationStrategy{Mode: "resolver_ip", PreferEcs: "geo"}
			in.Spec.RandomSteering = &CloudflareLoadBalancerRandomSteering{DefaultWeight: 0.5, PoolWeights: map[string]float64{"pool-primary": 0.8}}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing hostname", func() {
			in := validLoadBalancer()
			in.Spec.Hostname = ""
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an empty default_pools list", func() {
			in := validLoadBalancer()
			in.Spec.DefaultPools = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing fallback_pool", func() {
			in := validLoadBalancer()
			in.Spec.FallbackPool = nil
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid samesite value", func() {
			in := validLoadBalancer()
			in.Spec.SessionAffinityAttributes = &CloudflareLoadBalancerSessionAffinityAttributes{Samesite: "Bogus"}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a random_steering weight above 1", func() {
			in := validLoadBalancer()
			in.Spec.RandomSteering = &CloudflareLoadBalancerRandomSteering{DefaultWeight: 2}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a geo pool entry with no pools", func() {
			in := validLoadBalancer()
			in.Spec.RegionPools = []*CloudflareLoadBalancerGeoPools{{Code: "WNAM"}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
