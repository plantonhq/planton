package cloudflareloadbalancerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func ref(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func validLoadBalancer() *CloudflareLoadBalancer {
	return &CloudflareLoadBalancer{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareLoadBalancer",
		Metadata:   &shared.CatalogObjectMetadata{Name: "test-load-balancer"},
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

		ginkgo.It("accepts traffic rules with a fixed response and steering overrides", func() {
			in := validLoadBalancer()
			affinity := CloudflareLoadBalancerSessionAffinity_none
			policy := CloudflareLoadBalancerSteeringPolicy_off
			var priority int32 = 5
			in.Spec.Networks = []string{"network-a"}
			in.Spec.Rules = []*CloudflareLoadBalancerRule{
				{
					Name:      "maintenance-page",
					Condition: `http.request.uri.path contains "/maintenance"`,
					Priority:  &priority,
					FixedResponse: &CloudflareLoadBalancerRuleFixedResponse{
						ContentType: "text/html",
						MessageBody: "<h1>Down for maintenance</h1>",
						StatusCode:  503,
					},
				},
				{
					Name:      "api-steering",
					Condition: `http.request.uri.path contains "/api"`,
					Overrides: &CloudflareLoadBalancerRuleOverrides{
						SessionAffinity: &affinity,
						SteeringPolicy:  &policy,
						DefaultPools:    []*foreignkeyv1.StringValueOrRef{ref("pool-api")},
						FallbackPool:    ref("pool-api-fallback"),
						SessionAffinityAttributes: &CloudflareLoadBalancerSessionAffinityAttributes{
							Samesite: "None", Secure: "Always",
						},
						Ttl: 30,
					},
				},
			}
			gomega.Expect(protovalidate.Validate(in)).To(gomega.BeNil())
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

		ginkgo.It("rejects a negative rule priority", func() {
			in := validLoadBalancer()
			var priority int32 = -1
			in.Spec.Rules = []*CloudflareLoadBalancerRule{{Priority: &priority}}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid fixed_response status code", func() {
			in := validLoadBalancer()
			in.Spec.Rules = []*CloudflareLoadBalancerRule{
				{FixedResponse: &CloudflareLoadBalancerRuleFixedResponse{StatusCode: 99}},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects samesite None combined with secure Never", func() {
			in := validLoadBalancer()
			in.Spec.SessionAffinityAttributes = &CloudflareLoadBalancerSessionAffinityAttributes{
				Samesite: "None", Secure: "Never",
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects samesite None with secure Never inside rule overrides", func() {
			in := validLoadBalancer()
			in.Spec.Rules = []*CloudflareLoadBalancerRule{
				{Overrides: &CloudflareLoadBalancerRuleOverrides{
					SessionAffinityAttributes: &CloudflareLoadBalancerSessionAffinityAttributes{
						Samesite: "None", Secure: "Never",
					},
				}},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a negative ttl override", func() {
			in := validLoadBalancer()
			in.Spec.Rules = []*CloudflareLoadBalancerRule{
				{Overrides: &CloudflareLoadBalancerRuleOverrides{Ttl: -1}},
			}
			gomega.Expect(protovalidate.Validate(in)).ToNot(gomega.BeNil())
		})
	})
})
