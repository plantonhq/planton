package azurepointtositevpngatewayv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzurePointToSiteVpnGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePointToSiteVpnGatewaySpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testHubId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/test-hub"

	testVpnServerConfigurationId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/vpnServerConfigurations/remote-workforce"

	testRouteTableId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/test-hub/hubRouteTables/defaultRouteTable"
)

// validConnectionConfiguration returns a minimal client pool.
func validConnectionConfiguration(name string) *AzurePointToSiteVpnGatewayConnectionConfiguration {
	return &AzurePointToSiteVpnGatewayConnectionConfiguration{
		Name:            name,
		AddressPrefixes: []string{"172.16.201.0/24"},
	}
}

// validResource returns a minimal valid gateway that individual cases
// mutate into the shape under test.
func validResource() *AzurePointToSiteVpnGateway {
	return &AzurePointToSiteVpnGateway{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePointToSiteVpnGateway",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-p2s-gateway",
		},
		Spec: &AzurePointToSiteVpnGatewaySpec{
			Region:                   "eastus",
			ResourceGroup:            literal("test-rg"),
			Name:                     "remote-users-gw",
			VirtualHubId:             literal(testHubId),
			VpnServerConfigurationId: literal(testVpnServerConfigurationId),
			ConnectionConfigurations: []*AzurePointToSiteVpnGatewayConnectionConfiguration{
				validConnectionConfiguration("default-clients"),
			},
		},
	}
}

var _ = ginkgo.Describe("AzurePointToSiteVpnGatewaySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_point_to_site_vpn_gateway", func() {

			ginkgo.It("should not return a validation error for a minimal gateway", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit scale unit and DNS servers", func() {
				input := validResource()
				scaleUnit := int32(2)
				input.Spec.ScaleUnit = &scaleUnit
				input.Spec.DnsServers = []string{"10.0.0.4", "10.0.0.5"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a configured route block with propagation", func() {
				input := validResource()
				input.Spec.ConnectionConfigurations[0].Route = &AzurePointToSiteVpnGatewayRoute{
					AssociatedRouteTableId: literal(testRouteTableId),
					PropagatedRouteTable: &AzurePointToSiteVpnGatewayPropagatedRouteTable{
						RouteTableIds: []*foreignkeyv1.StringValueOrRef{literal(testRouteTableId)},
						Labels:        []string{"default"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept route maps on the route block", func() {
				input := validResource()
				routeMapId := testHubId + "/routeMaps/strip-communities"
				input.Spec.ConnectionConfigurations[0].Route = &AzurePointToSiteVpnGatewayRoute{
					AssociatedRouteTableId: literal(testRouteTableId),
					InboundRouteMapId:      literal(routeMapId),
					OutboundRouteMapId:     literal(routeMapId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept two connection configurations with distinct names", func() {
				input := validResource()
				second := validConnectionConfiguration("contractors")
				second.AddressPrefixes = []string{"172.16.202.0/24"}
				second.InternetSecurityEnabled = true
				input.Spec.ConnectionConfigurations = append(input.Spec.ConnectionConfigurations, second)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the internet routing preference", func() {
				input := validResource()
				input.Spec.RoutingPreferenceInternetEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_point_to_site_vpn_gateway", func() {

			ginkgo.It("should reject a missing virtual hub reference", func() {
				input := validResource()
				input.Spec.VirtualHubId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing VPN server configuration reference", func() {
				input := validResource()
				input.Spec.VpnServerConfigurationId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a gateway without connection configurations", func() {
				input := validResource()
				input.Spec.ConnectionConfigurations = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a connection configuration without an address pool", func() {
				input := validResource()
				input.Spec.ConnectionConfigurations[0].AddressPrefixes = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate connection configuration names", func() {
				input := validResource()
				input.Spec.ConnectionConfigurations = []*AzurePointToSiteVpnGatewayConnectionConfiguration{
					validConnectionConfiguration("default-clients"),
					validConnectionConfiguration("default-clients"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a route block without its associated route table", func() {
				input := validResource()
				input.Spec.ConnectionConfigurations[0].Route = &AzurePointToSiteVpnGatewayRoute{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a propagated route table without targets", func() {
				input := validResource()
				input.Spec.ConnectionConfigurations[0].Route = &AzurePointToSiteVpnGatewayRoute{
					AssociatedRouteTableId: literal(testRouteTableId),
					PropagatedRouteTable:   &AzurePointToSiteVpnGatewayPropagatedRouteTable{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a non-IPv4 DNS server", func() {
				input := validResource()
				input.Spec.DnsServers = []string{"dns.example.com"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a negative scale unit", func() {
				input := validResource()
				scaleUnit := int32(-1)
				input.Spec.ScaleUnit = &scaleUnit
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
