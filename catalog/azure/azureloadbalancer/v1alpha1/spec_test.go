package azureloadbalancerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureLoadBalancerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureLoadBalancerSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testSubnetId   = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/default"
	testPublicIpId = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/publicIPAddresses/pip"
	testPrefixId   = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/publicIPPrefixes/prefix"
	testVnetId     = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet"
)

// baseLoadBalancer returns a minimal valid public load balancer that
// individual cases mutate to exercise one contract at a time.
func baseLoadBalancer() *AzureLoadBalancer {
	return &AzureLoadBalancer{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureLoadBalancer",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-lb",
		},
		Spec: &AzureLoadBalancerSpec{
			Region:        "eastus",
			ResourceGroup: literal("my-rg"),
			Name:          "test-lb",
			FrontendIpConfigurations: []*AzureLoadBalancerFrontendIpConfiguration{
				{
					Name:              "public",
					PublicIpAddressId: literal(testPublicIpId),
				},
			},
			BackendPools: []*AzureLoadBalancerBackendPool{
				{Name: "web"},
			},
			HealthProbes: []*AzureLoadBalancerHealthProbe{
				{
					Name:     "tcp-health",
					Protocol: AzureLoadBalancerProbeProtocol_PROBE_TCP,
					Port:     80,
				},
			},
			Rules: []*AzureLoadBalancerRule{
				{
					Name:             "http",
					Protocol:         AzureLoadBalancerTransportProtocol_TCP,
					FrontendPort:     80,
					BackendPort:      80,
					BackendPoolNames: []string{"web"},
					ProbeName:        "tcp-health",
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureLoadBalancerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal public load balancer", func() {
			err := protovalidate.Validate(baseLoadBalancer())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a frontend-only load balancer (no pools, probes, or rules)", func() {
			input := baseLoadBalancer()
			input.Spec.BackendPools = nil
			input.Spec.HealthProbes = nil
			input.Spec.Rules = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a rule without a probe", func() {
			input := baseLoadBalancer()
			input.Spec.Rules[0].ProbeName = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an internal frontend with a pinned static address and zones", func() {
			input := baseLoadBalancer()
			input.Spec.FrontendIpConfigurations = []*AzureLoadBalancerFrontendIpConfiguration{
				{
					Name:             "internal",
					SubnetId:         literal(testSubnetId),
					PrivateIpAddress: "10.0.1.100",
					Zones:            []string{"1", "2", "3"},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a public-IP-prefix frontend", func() {
			input := baseLoadBalancer()
			input.Spec.FrontendIpConfigurations = []*AzureLoadBalancerFrontendIpConfiguration{
				{
					Name:             "prefix-frontend",
					PublicIpPrefixId: literal(testPrefixId),
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept multiple frontends with rules naming their frontend", func() {
			input := baseLoadBalancer()
			input.Spec.FrontendIpConfigurations = append(input.Spec.FrontendIpConfigurations,
				&AzureLoadBalancerFrontendIpConfiguration{
					Name:     "internal",
					SubnetId: literal(testSubnetId),
				})
			input.Spec.Rules[0].FrontendIpConfigurationName = "public"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an HA-ports rule (protocol ALL, ports 0) ", func() {
			input := baseLoadBalancer()
			input.Spec.FrontendIpConfigurations[0] = &AzureLoadBalancerFrontendIpConfiguration{
				Name:     "internal",
				SubnetId: literal(testSubnetId),
			}
			input.Spec.Rules[0].Protocol = AzureLoadBalancerTransportProtocol_ALL
			input.Spec.Rules[0].FrontendPort = 0
			input.Spec.Rules[0].BackendPort = 0
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept HTTP and HTTPS probes with request paths and a probe threshold", func() {
			threshold := int32(3)
			input := baseLoadBalancer()
			input.Spec.HealthProbes = []*AzureLoadBalancerHealthProbe{
				{
					Name:           "http-health",
					Protocol:       AzureLoadBalancerProbeProtocol_PROBE_HTTP,
					Port:           80,
					RequestPath:    "/healthz",
					ProbeThreshold: &threshold,
				},
				{
					Name:        "https-health",
					Protocol:    AzureLoadBalancerProbeProtocol_PROBE_HTTPS,
					Port:        443,
					RequestPath: "/healthz",
				},
			}
			input.Spec.Rules[0].ProbeName = "http-health"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a GATEWAY SKU load balancer with tunnel interfaces", func() {
			input := baseLoadBalancer()
			input.Spec.Sku = AzureLoadBalancerSku_GATEWAY
			input.Spec.FrontendIpConfigurations[0] = &AzureLoadBalancerFrontendIpConfiguration{
				Name:     "gateway-frontend",
				SubnetId: literal(testSubnetId),
			}
			input.Spec.BackendPools = []*AzureLoadBalancerBackendPool{
				{
					Name: "nva-pool",
					TunnelInterfaces: []*AzureLoadBalancerBackendPoolTunnelInterface{
						{
							Identifier: 800,
							Port:       10800,
							Protocol:   AzureLoadBalancerTunnelProtocol_VXLAN,
							Type:       AzureLoadBalancerTunnelType_INTERNAL,
						},
						{
							Identifier: 801,
							Port:       10801,
							Protocol:   AzureLoadBalancerTunnelProtocol_VXLAN,
							Type:       AzureLoadBalancerTunnelType_EXTERNAL,
						},
					},
				},
			}
			input.Spec.Rules = []*AzureLoadBalancerRule{
				{
					Name:             "chain",
					Protocol:         AzureLoadBalancerTransportProtocol_ALL,
					FrontendPort:     0,
					BackendPort:      0,
					BackendPoolNames: []string{"nva-pool"},
				},
			}
			input.Spec.HealthProbes = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept IP-based pool addresses on a vnet-scoped pool", func() {
			input := baseLoadBalancer()
			input.Spec.BackendPools[0].VirtualNetworkId = literal(testVnetId)
			input.Spec.BackendPools[0].Addresses = []*AzureLoadBalancerBackendPoolAddress{
				{Name: "appliance-1", IpAddress: "10.0.1.10"},
				{Name: "appliance-2", IpAddress: "10.0.1.11"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a GLOBAL-tier pool of regional load balancer frontends", func() {
			input := baseLoadBalancer()
			input.Spec.SkuTier = AzureLoadBalancerSkuTier_GLOBAL
			input.Spec.BackendPools[0].Addresses = []*AzureLoadBalancerBackendPoolAddress{
				{
					Name:                                  "eastus-lb",
					LoadBalancerFrontendIpConfigurationId: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/regional/frontendIPConfigurations/public",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a single-target NAT rule", func() {
			input := baseLoadBalancer()
			input.Spec.NatRules = []*AzureLoadBalancerNatRule{
				{
					Name:         "ssh-admin",
					Protocol:     AzureLoadBalancerTransportProtocol_TCP,
					FrontendPort: 2222,
					BackendPort:  22,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a pool-style NAT rule with a port range", func() {
			input := baseLoadBalancer()
			input.Spec.NatRules = []*AzureLoadBalancerNatRule{
				{
					Name:              "per-instance-ssh",
					Protocol:          AzureLoadBalancerTransportProtocol_TCP,
					BackendPort:       22,
					BackendPoolName:   "web",
					FrontendPortStart: 50000,
					FrontendPortEnd:   50099,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an outbound rule against a declared pool and frontend", func() {
			ports := int32(2048)
			disableSnat := true
			input := baseLoadBalancer()
			input.Spec.Rules[0].DisableOutboundSnat = &disableSnat
			input.Spec.OutboundRules = []*AzureLoadBalancerOutboundRule{
				{
					Name:                         "egress",
					FrontendIpConfigurationNames: []string{"public"},
					BackendPoolName:              "web",
					Protocol:                     AzureLoadBalancerTransportProtocol_ALL,
					AllocatedOutboundPorts:       &ports,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept user tags and load distribution", func() {
			input := baseLoadBalancer()
			input.Spec.Tags = map[string]string{"cost-center": "networking"}
			input.Spec.Rules[0].LoadDistribution = AzureLoadBalancerLoadDistribution_SOURCE_IP
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing region", func() {
			input := baseLoadBalancer()
			input.Spec.Region = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a missing resource group", func() {
			input := baseLoadBalancer()
			input.Spec.ResourceGroup = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a missing name", func() {
			input := baseLoadBalancer()
			input.Spec.Name = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an empty frontend list", func() {
			input := baseLoadBalancer()
			input.Spec.FrontendIpConfigurations = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a frontend with two address sources", func() {
			input := baseLoadBalancer()
			input.Spec.FrontendIpConfigurations[0].SubnetId = literal(testSubnetId)
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a pinned private address on a public frontend", func() {
			input := baseLoadBalancer()
			input.Spec.FrontendIpConfigurations[0].PrivateIpAddress = "10.0.1.100"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject zones on a public frontend", func() {
			input := baseLoadBalancer()
			input.Spec.FrontendIpConfigurations[0].Zones = []string{"1"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject GLOBAL tier on the GATEWAY SKU", func() {
			input := baseLoadBalancer()
			input.Spec.Sku = AzureLoadBalancerSku_GATEWAY
			input.Spec.SkuTier = AzureLoadBalancerSkuTier_GLOBAL
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a GATEWAY load balancer whose pool lacks tunnel interfaces", func() {
			input := baseLoadBalancer()
			input.Spec.Sku = AzureLoadBalancerSku_GATEWAY
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject tunnel interfaces on the STANDARD SKU", func() {
			input := baseLoadBalancer()
			input.Spec.BackendPools[0].TunnelInterfaces = []*AzureLoadBalancerBackendPoolTunnelInterface{
				{
					Identifier: 800,
					Port:       10800,
					Protocol:   AzureLoadBalancerTunnelProtocol_VXLAN,
					Type:       AzureLoadBalancerTunnelType_INTERNAL,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a two-pool rule on the STANDARD SKU", func() {
			input := baseLoadBalancer()
			input.Spec.BackendPools = append(input.Spec.BackendPools, &AzureLoadBalancerBackendPool{Name: "api"})
			input.Spec.Rules[0].BackendPoolNames = []string{"web", "api"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a rule referencing an undeclared pool", func() {
			input := baseLoadBalancer()
			input.Spec.Rules[0].BackendPoolNames = []string{"missing"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a rule referencing an undeclared probe", func() {
			input := baseLoadBalancer()
			input.Spec.Rules[0].ProbeName = "missing"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a rule omitting the frontend when several frontends exist", func() {
			input := baseLoadBalancer()
			input.Spec.FrontendIpConfigurations = append(input.Spec.FrontendIpConfigurations,
				&AzureLoadBalancerFrontendIpConfiguration{
					Name:     "internal",
					SubnetId: literal(testSubnetId),
				})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a rule naming an undeclared frontend", func() {
			input := baseLoadBalancer()
			input.Spec.Rules[0].FrontendIpConfigurationName = "missing"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an HA-ports rule with non-zero ports", func() {
			input := baseLoadBalancer()
			input.Spec.Rules[0].Protocol = AzureLoadBalancerTransportProtocol_ALL
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a non-ALL rule with port 0", func() {
			input := baseLoadBalancer()
			input.Spec.Rules[0].FrontendPort = 0
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a rule without a protocol", func() {
			input := baseLoadBalancer()
			input.Spec.Rules[0].Protocol = AzureLoadBalancerTransportProtocol_azure_load_balancer_transport_protocol_unspecified
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an HTTP probe without a request path", func() {
			input := baseLoadBalancer()
			input.Spec.HealthProbes[0].Protocol = AzureLoadBalancerProbeProtocol_PROBE_HTTP
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a TCP probe with a request path", func() {
			input := baseLoadBalancer()
			input.Spec.HealthProbes[0].RequestPath = "/healthz"
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a probe interval below 5 seconds", func() {
			interval := int32(3)
			input := baseLoadBalancer()
			input.Spec.HealthProbes[0].IntervalInSeconds = &interval
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a probe threshold above 100", func() {
			threshold := int32(101)
			input := baseLoadBalancer()
			input.Spec.HealthProbes[0].ProbeThreshold = &threshold
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a rule idle timeout outside 4-100", func() {
			timeout := int32(101)
			input := baseLoadBalancer()
			input.Spec.Rules[0].IdleTimeoutInMinutes = &timeout
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject synchronous_mode without a pool virtual network", func() {
			input := baseLoadBalancer()
			input.Spec.BackendPools[0].SynchronousMode = AzureLoadBalancerBackendPoolSyncMode_AUTOMATIC
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject IP-based addresses without the pool virtual network", func() {
			input := baseLoadBalancer()
			input.Spec.BackendPools[0].Addresses = []*AzureLoadBalancerBackendPoolAddress{
				{Name: "appliance-1", IpAddress: "10.0.1.10"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a pool address with both an IP and a regional frontend", func() {
			input := baseLoadBalancer()
			input.Spec.BackendPools[0].VirtualNetworkId = literal(testVnetId)
			input.Spec.BackendPools[0].Addresses = []*AzureLoadBalancerBackendPoolAddress{
				{
					Name:                                  "bad",
					IpAddress:                             "10.0.1.10",
					LoadBalancerFrontendIpConfigurationId: "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/loadBalancers/r/frontendIPConfigurations/f",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a NAT rule setting both modes", func() {
			input := baseLoadBalancer()
			input.Spec.NatRules = []*AzureLoadBalancerNatRule{
				{
					Name:              "bad",
					Protocol:          AzureLoadBalancerTransportProtocol_TCP,
					FrontendPort:      2222,
					BackendPort:       22,
					BackendPoolName:   "web",
					FrontendPortStart: 50000,
					FrontendPortEnd:   50099,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a NAT rule with neither mode", func() {
			input := baseLoadBalancer()
			input.Spec.NatRules = []*AzureLoadBalancerNatRule{
				{
					Name:        "bad",
					Protocol:    AzureLoadBalancerTransportProtocol_TCP,
					BackendPort: 22,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a pool-style NAT rule missing its port range", func() {
			input := baseLoadBalancer()
			input.Spec.NatRules = []*AzureLoadBalancerNatRule{
				{
					Name:            "bad",
					Protocol:        AzureLoadBalancerTransportProtocol_TCP,
					BackendPort:     22,
					BackendPoolName: "web",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a pool-style NAT rule referencing an undeclared pool", func() {
			input := baseLoadBalancer()
			input.Spec.NatRules = []*AzureLoadBalancerNatRule{
				{
					Name:              "bad",
					Protocol:          AzureLoadBalancerTransportProtocol_TCP,
					BackendPort:       22,
					BackendPoolName:   "missing",
					FrontendPortStart: 50000,
					FrontendPortEnd:   50099,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a NAT rule idle timeout above 30 minutes", func() {
			timeout := int32(31)
			input := baseLoadBalancer()
			input.Spec.NatRules = []*AzureLoadBalancerNatRule{
				{
					Name:                 "bad",
					Protocol:             AzureLoadBalancerTransportProtocol_TCP,
					FrontendPort:         2222,
					BackendPort:          22,
					IdleTimeoutInMinutes: &timeout,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an outbound rule referencing an undeclared pool", func() {
			input := baseLoadBalancer()
			input.Spec.OutboundRules = []*AzureLoadBalancerOutboundRule{
				{
					Name:                         "bad",
					FrontendIpConfigurationNames: []string{"public"},
					BackendPoolName:              "missing",
					Protocol:                     AzureLoadBalancerTransportProtocol_ALL,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an outbound rule with no frontends", func() {
			input := baseLoadBalancer()
			input.Spec.OutboundRules = []*AzureLoadBalancerOutboundRule{
				{
					Name:            "bad",
					BackendPoolName: "web",
					Protocol:        AzureLoadBalancerTransportProtocol_ALL,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an outbound rule naming an undeclared frontend", func() {
			input := baseLoadBalancer()
			input.Spec.OutboundRules = []*AzureLoadBalancerOutboundRule{
				{
					Name:                         "bad",
					FrontendIpConfigurationNames: []string{"missing"},
					BackendPoolName:              "web",
					Protocol:                     AzureLoadBalancerTransportProtocol_ALL,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
		})
	})
})
