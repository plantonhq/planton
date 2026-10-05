package azurevirtualhubv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureVirtualHubSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVirtualHubSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// skuPtr returns a pointer to the given sku enum value.
func skuPtr(value AzureVirtualHubSku) *AzureVirtualHubSku {
	return &value
}

// prefPtr returns a pointer to the given routing-preference enum value.
func prefPtr(value AzureVirtualHubRoutingPreference) *AzureVirtualHubRoutingPreference {
	return &value
}

// stepPtr returns a pointer to the given next-step enum value.
func stepPtr(value AzureVirtualHubRouteMapNextStep) *AzureVirtualHubRouteMapNextStep {
	return &value
}

// int32Ptr returns a pointer to the given int32 (for optional fields).
func int32Ptr(value int32) *int32 {
	return &value
}

const testFirewallId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/azureFirewalls/hub-fw"

// validResource returns a minimal valid WAN-attached Standard hub that
// individual cases mutate into the shape under test.
func validResource() *AzureVirtualHub {
	return &AzureVirtualHub{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVirtualHub",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-vhub",
		},
		Spec: &AzureVirtualHubSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "hub-eastus",
			VirtualWanId:  literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualWans/global-wan"),
			AddressPrefix: "10.100.0.0/23",
		},
	}
}

// validRouteTable returns a route table with one CIDR route at a
// firewall next hop.
func validRouteTable(name string) *AzureVirtualHubRouteTable {
	return &AzureVirtualHubRouteTable{
		Name:   name,
		Labels: []string{"prod"},
		Routes: []*AzureVirtualHubRouteTableRoute{
			{
				Name:             "to-firewall",
				DestinationsType: AzureVirtualHubRouteDestinationsType_CIDR,
				Destinations:     []string{"0.0.0.0/0"},
				NextHop:          literal(testFirewallId),
			},
		},
	}
}

// validRouteMap returns a route map with one transform rule and one
// drop rule.
func validRouteMap(name string) *AzureVirtualHubRouteMap {
	return &AzureVirtualHubRouteMap{
		Name: name,
		Rules: []*AzureVirtualHubRouteMapRule{
			{
				Name: "tag-on-prem",
				MatchCriteria: []*AzureVirtualHubRouteMapMatchCriterion{
					{
						MatchCondition: AzureVirtualHubRouteMapMatchCondition_CONTAINS,
						RoutePrefix:    []string{"10.0.0.0/8"},
					},
				},
				Actions: []*AzureVirtualHubRouteMapAction{
					{
						Type: AzureVirtualHubRouteMapActionType_ADD,
						Parameters: []*AzureVirtualHubRouteMapActionParameter{
							{Community: []string{"65001:100"}},
						},
					},
				},
				NextStepIfMatched: stepPtr(AzureVirtualHubRouteMapNextStep_CONTINUE),
			},
			{
				Name: "drop-test-ranges",
				MatchCriteria: []*AzureVirtualHubRouteMapMatchCriterion{
					{
						MatchCondition: AzureVirtualHubRouteMapMatchCondition_EQUALS,
						RoutePrefix:    []string{"192.0.2.0/24"},
					},
				},
				Actions: []*AzureVirtualHubRouteMapAction{
					{Type: AzureVirtualHubRouteMapActionType_DROP},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureVirtualHubSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_virtual_hub", func() {

			ginkgo.It("should not return a validation error for a minimal WAN-attached hub", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a fully-tuned hub", func() {
				input := validResource()
				input.Spec.Sku = skuPtr(AzureVirtualHubSku_BASIC)
				input.Spec.HubRoutingPreference = prefPtr(AzureVirtualHubRoutingPreference_AS_PATH)
				input.Spec.BranchToBranchTrafficEnabled = true
				input.Spec.VirtualRouterAutoScaleMinCapacity = int32Ptr(5)
				input.Spec.Tags = map[string]string{"cost-center": "networking"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept inline hub routes", func() {
				input := validResource()
				input.Spec.Routes = []*AzureVirtualHubRoute{
					{
						AddressPrefixes:  []string{"10.20.0.0/16", "10.30.0.0/16"},
						NextHopIpAddress: "10.100.0.68",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept custom route tables with routes", func() {
				input := validResource()
				input.Spec.RouteTables = []*AzureVirtualHubRouteTable{
					validRouteTable("prod-isolated"),
					validRouteTable("shared-services"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept route maps including a parameterless DROP action", func() {
				input := validResource()
				input.Spec.RouteMaps = []*AzureVirtualHubRouteMap{validRouteMap("ingress-policy")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a BGP peering with an NVA", func() {
				input := validResource()
				input.Spec.BgpConnections = []*AzureVirtualHubBgpConnection{
					{
						Name:    "nva-peering",
						PeerAsn: 65010,
						PeerIp:  "10.20.1.4",
						VirtualNetworkConnectionId: literal(
							"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus/hubVirtualNetworkConnections/nva-spoke"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a routing intent steering both categories", func() {
				input := validResource()
				input.Spec.RoutingIntent = &AzureVirtualHubRoutingIntent{
					Name: "hubRoutingIntent",
					RoutingPolicies: []*AzureVirtualHubRoutingPolicy{
						{
							Name:         "InternetTraffic",
							Destinations: []AzureVirtualHubRoutingPolicyDestination{AzureVirtualHubRoutingPolicyDestination_INTERNET},
							NextHop:      literal(testFirewallId),
						},
						{
							Name:         "PrivateTraffic",
							Destinations: []AzureVirtualHubRoutingPolicyDestination{AzureVirtualHubRoutingPolicyDestination_PRIVATE_TRAFFIC},
							NextHop:      literal(testFirewallId),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_virtual_hub", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when virtual_wan_id is missing", func() {
				input := validResource()
				input.Spec.VirtualWanId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when address_prefix is missing", func() {
				input := validResource()
				input.Spec.AddressPrefix = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an address_prefix without a mask", func() {
				input := validResource()
				input.Spec.AddressPrefix = "10.100.0.0"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an undefined sku", func() {
				input := validResource()
				input.Spec.Sku = skuPtr(AzureVirtualHubSku(99))
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an auto-scale capacity below the ARM floor of 2", func() {
				input := validResource()
				input.Spec.VirtualRouterAutoScaleMinCapacity = int32Ptr(1)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an inline route with a malformed next hop", func() {
				input := validResource()
				input.Spec.Routes = []*AzureVirtualHubRoute{
					{AddressPrefixes: []string{"10.20.0.0/16"}, NextHopIpAddress: "not-an-ip"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an inline route with no prefixes", func() {
				input := validResource()
				input.Spec.Routes = []*AzureVirtualHubRoute{
					{NextHopIpAddress: "10.100.0.68"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a route table name with forbidden characters", func() {
				input := validResource()
				table := validRouteTable("bad?name")
				input.Spec.RouteTables = []*AzureVirtualHubRouteTable{table}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate route table names", func() {
				input := validResource()
				input.Spec.RouteTables = []*AzureVirtualHubRouteTable{
					validRouteTable("prod"),
					validRouteTable("prod"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a route without an explicit destinations type", func() {
				input := validResource()
				table := validRouteTable("prod")
				table.Routes[0].DestinationsType = AzureVirtualHubRouteDestinationsType_azure_virtual_hub_route_destinations_type_unspecified
				input.Spec.RouteTables = []*AzureVirtualHubRouteTable{table}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a route with no destinations", func() {
				input := validResource()
				table := validRouteTable("prod")
				table.Routes[0].Destinations = nil
				input.Spec.RouteTables = []*AzureVirtualHubRouteTable{table}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a route without a next hop", func() {
				input := validResource()
				table := validRouteTable("prod")
				table.Routes[0].NextHop = nil
				input.Spec.RouteTables = []*AzureVirtualHubRouteTable{table}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate route names within a table", func() {
				input := validResource()
				table := validRouteTable("prod")
				table.Routes = append(table.Routes, table.Routes[0])
				input.Spec.RouteTables = []*AzureVirtualHubRouteTable{table}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a route map name starting with a hyphen", func() {
				input := validResource()
				input.Spec.RouteMaps = []*AzureVirtualHubRouteMap{validRouteMap("-bad")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate route map names", func() {
				input := validResource()
				input.Spec.RouteMaps = []*AzureVirtualHubRouteMap{
					validRouteMap("policy"),
					validRouteMap("policy"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an ADD action with no parameters", func() {
				input := validResource()
				routeMap := validRouteMap("policy")
				routeMap.Rules[0].Actions[0].Parameters = nil
				input.Spec.RouteMaps = []*AzureVirtualHubRouteMap{routeMap}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an action without an explicit type", func() {
				input := validResource()
				routeMap := validRouteMap("policy")
				routeMap.Rules[0].Actions[0].Type = AzureVirtualHubRouteMapActionType_azure_virtual_hub_route_map_action_type_unspecified
				input.Spec.RouteMaps = []*AzureVirtualHubRouteMap{routeMap}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a criterion without an explicit condition", func() {
				input := validResource()
				routeMap := validRouteMap("policy")
				routeMap.Rules[0].MatchCriteria[0].MatchCondition = AzureVirtualHubRouteMapMatchCondition_azure_virtual_hub_route_map_match_condition_unspecified
				input.Spec.RouteMaps = []*AzureVirtualHubRouteMap{routeMap}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a criterion with nothing to match", func() {
				input := validResource()
				routeMap := validRouteMap("policy")
				routeMap.Rules[0].MatchCriteria[0].RoutePrefix = nil
				input.Spec.RouteMaps = []*AzureVirtualHubRouteMap{routeMap}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate rule names within a map", func() {
				input := validResource()
				routeMap := validRouteMap("policy")
				routeMap.Rules[1].Name = routeMap.Rules[0].Name
				input.Spec.RouteMaps = []*AzureVirtualHubRouteMap{routeMap}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a BGP peering with a negative ASN", func() {
				input := validResource()
				input.Spec.BgpConnections = []*AzureVirtualHubBgpConnection{
					{Name: "nva", PeerAsn: -1, PeerIp: "10.20.1.4"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a BGP peering with a malformed peer IP", func() {
				input := validResource()
				input.Spec.BgpConnections = []*AzureVirtualHubBgpConnection{
					{Name: "nva", PeerAsn: 65010, PeerIp: "peer.example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate BGP peering names", func() {
				input := validResource()
				input.Spec.BgpConnections = []*AzureVirtualHubBgpConnection{
					{Name: "nva", PeerAsn: 65010, PeerIp: "10.20.1.4"},
					{Name: "nva", PeerAsn: 65011, PeerIp: "10.20.1.5"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a routing intent with no policies", func() {
				input := validResource()
				input.Spec.RoutingIntent = &AzureVirtualHubRoutingIntent{Name: "hubRoutingIntent"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a routing policy with no destinations", func() {
				input := validResource()
				input.Spec.RoutingIntent = &AzureVirtualHubRoutingIntent{
					Name: "hubRoutingIntent",
					RoutingPolicies: []*AzureVirtualHubRoutingPolicy{
						{Name: "InternetTraffic", NextHop: literal(testFirewallId)},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a routing policy without a next hop", func() {
				input := validResource()
				input.Spec.RoutingIntent = &AzureVirtualHubRoutingIntent{
					Name: "hubRoutingIntent",
					RoutingPolicies: []*AzureVirtualHubRoutingPolicy{
						{
							Name:         "InternetTraffic",
							Destinations: []AzureVirtualHubRoutingPolicyDestination{AzureVirtualHubRoutingPolicyDestination_INTERNET},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate routing policy names", func() {
				input := validResource()
				input.Spec.RoutingIntent = &AzureVirtualHubRoutingIntent{
					Name: "hubRoutingIntent",
					RoutingPolicies: []*AzureVirtualHubRoutingPolicy{
						{
							Name:         "policy",
							Destinations: []AzureVirtualHubRoutingPolicyDestination{AzureVirtualHubRoutingPolicyDestination_INTERNET},
							NextHop:      literal(testFirewallId),
						},
						{
							Name:         "policy",
							Destinations: []AzureVirtualHubRoutingPolicyDestination{AzureVirtualHubRoutingPolicyDestination_PRIVATE_TRAFFIC},
							NextHop:      literal(testFirewallId),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
