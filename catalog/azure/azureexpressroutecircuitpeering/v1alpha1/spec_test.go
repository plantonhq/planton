package azureexpressroutecircuitpeeringv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureExpressRouteCircuitPeeringSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureExpressRouteCircuitPeeringSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// validResource returns a minimal valid PRIVATE peering that individual
// cases mutate into the shape under test.
func validResource() *AzureExpressRouteCircuitPeering {
	return &AzureExpressRouteCircuitPeering{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureExpressRouteCircuitPeering",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ercp",
		},
		Spec: &AzureExpressRouteCircuitPeeringSpec{
			ResourceGroup:              literal("test-rg"),
			ExpressRouteCircuitName:    literal("hq-circuit"),
			PeeringType:                AzureExpressRouteCircuitPeeringType_AZURE_PRIVATE_PEERING,
			VlanId:                     100,
			PrimaryPeerAddressPrefix:   "192.168.16.0/30",
			SecondaryPeerAddressPrefix: "192.168.16.4/30",
			PeerAsn:                    65010,
		},
	}
}

// validMicrosoftPeering returns a valid MICROSOFT peering with the
// mandatory advertisement contract.
func validMicrosoftPeering() *AzureExpressRouteCircuitPeering {
	input := validResource()
	input.Spec.PeeringType = AzureExpressRouteCircuitPeeringType_MICROSOFT_PEERING
	input.Spec.MicrosoftPeeringConfig = &AzureExpressRouteCircuitPeeringMicrosoftConfig{
		AdvertisedPublicPrefixes: []string{"203.0.113.0/24"},
	}
	return input
}

var _ = ginkgo.Describe("AzureExpressRouteCircuitPeeringSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_express_route_circuit_peering", func() {

			ginkgo.It("should not return a validation error for a minimal private peering", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a Microsoft peering with its advertisement contract", func() {
				err := protovalidate.Validate(validMicrosoftPeering())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a Microsoft peering with a route filter and registry details", func() {
				input := validMicrosoftPeering()
				input.Spec.RouteFilterId = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/routeFilters/o365"
				input.Spec.MicrosoftPeeringConfig.CustomerAsn = 65100
				registry := "ARIN"
				input.Spec.MicrosoftPeeringConfig.RoutingRegistryName = &registry
				input.Spec.MicrosoftPeeringConfig.AdvertisedCommunities = []string{"12076:20000"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an IPv6 block on a private peering", func() {
				input := validResource()
				input.Spec.Ipv6 = &AzureExpressRouteCircuitPeeringIpv6{
					PrimaryPeerAddressPrefix:   "2001:db8::/126",
					SecondaryPeerAddressPrefix: "2001:db8::4/126",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept Global Reach connections on a private peering", func() {
				input := validResource()
				input.Spec.Connections = []*AzureExpressRouteCircuitPeeringConnection{
					{
						Name:              "hq-to-branch",
						PeerPeeringId:     literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/expressRouteCircuits/branch/peerings/AzurePrivatePeering"),
						AddressPrefixIpv4: "172.16.0.0/29",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a shared key reference and disabled ipv4", func() {
				input := validResource()
				input.Spec.SharedKey = literal("bgp-md5-key")
				disabled := false
				input.Spec.Ipv4Enabled = &disabled
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_express_route_circuit_peering", func() {

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the circuit name is missing", func() {
				input := validResource()
				input.Spec.ExpressRouteCircuitName = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the peering type is unspecified", func() {
				input := validResource()
				input.Spec.PeeringType = AzureExpressRouteCircuitPeeringType_azure_express_route_circuit_peering_type_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for VLAN id zero", func() {
				input := validResource()
				input.Spec.VlanId = 0
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for VLAN id 4095", func() {
				input := validResource()
				input.Spec.VlanId = 4095
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an unpaired primary prefix", func() {
				input := validResource()
				input.Spec.SecondaryPeerAddressPrefix = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a route filter on a private peering", func() {
				input := validResource()
				input.Spec.RouteFilterId = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/routeFilters/o365"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a Microsoft config on a private peering", func() {
				input := validResource()
				input.Spec.MicrosoftPeeringConfig = &AzureExpressRouteCircuitPeeringMicrosoftConfig{
					AdvertisedPublicPrefixes: []string{"203.0.113.0/24"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a Microsoft peering with prefixes but no config", func() {
				input := validMicrosoftPeering()
				input.Spec.MicrosoftPeeringConfig = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a Microsoft config without session prefixes", func() {
				input := validMicrosoftPeering()
				input.Spec.PrimaryPeerAddressPrefix = ""
				input.Spec.SecondaryPeerAddressPrefix = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an empty advertised-prefix list", func() {
				input := validMicrosoftPeering()
				input.Spec.MicrosoftPeeringConfig.AdvertisedPublicPrefixes = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for ipv6 on the deprecated public peering", func() {
				input := validResource()
				input.Spec.PeeringType = AzureExpressRouteCircuitPeeringType_AZURE_PUBLIC_PEERING
				input.Spec.Ipv6 = &AzureExpressRouteCircuitPeeringIpv6{
					PrimaryPeerAddressPrefix:   "2001:db8::/126",
					SecondaryPeerAddressPrefix: "2001:db8::4/126",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an ipv6 block missing its secondary prefix", func() {
				input := validResource()
				input.Spec.Ipv6 = &AzureExpressRouteCircuitPeeringIpv6{
					PrimaryPeerAddressPrefix: "2001:db8::/126",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for connections on a Microsoft peering", func() {
				input := validMicrosoftPeering()
				input.Spec.Connections = []*AzureExpressRouteCircuitPeeringConnection{
					{
						Name:              "hq-to-branch",
						PeerPeeringId:     literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/expressRouteCircuits/branch/peerings/AzurePrivatePeering"),
						AddressPrefixIpv4: "172.16.0.0/29",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate connection names", func() {
				input := validResource()
				connection := &AzureExpressRouteCircuitPeeringConnection{
					Name:              "hq-to-branch",
					PeerPeeringId:     literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/expressRouteCircuits/branch/peerings/AzurePrivatePeering"),
					AddressPrefixIpv4: "172.16.0.0/29",
				}
				input.Spec.Connections = []*AzureExpressRouteCircuitPeeringConnection{connection, connection}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a connection without a far side", func() {
				input := validResource()
				input.Spec.Connections = []*AzureExpressRouteCircuitPeeringConnection{
					{
						Name:              "hq-to-branch",
						AddressPrefixIpv4: "172.16.0.0/29",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a connection without an address prefix", func() {
				input := validResource()
				input.Spec.Connections = []*AzureExpressRouteCircuitPeeringConnection{
					{
						Name:          "hq-to-branch",
						PeerPeeringId: literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/expressRouteCircuits/branch/peerings/AzurePrivatePeering"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a connection name ending with a hyphen", func() {
				input := validResource()
				input.Spec.Connections = []*AzureExpressRouteCircuitPeeringConnection{
					{
						Name:              "hq-to-branch-",
						PeerPeeringId:     literal("/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/expressRouteCircuits/branch/peerings/AzurePrivatePeering"),
						AddressPrefixIpv4: "172.16.0.0/29",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
