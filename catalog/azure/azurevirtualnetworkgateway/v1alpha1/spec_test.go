package azurevirtualnetworkgatewayv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureVirtualNetworkGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVirtualNetworkGatewaySpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// ref builds a StringValueOrRef carrying a value_from reference.
func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

// ipConfiguration builds a gateway ip configuration with a public IP.
func ipConfiguration(name string) *AzureVirtualNetworkGatewayIpConfiguration {
	return &AzureVirtualNetworkGatewayIpConfiguration{
		Name:              name,
		SubnetId:          ref("gateway-subnet"),
		PublicIpAddressId: ref("gateway-pip-" + name),
	}
}

// validResource returns a minimal valid route-based VPN gateway that
// individual cases mutate into the shape under test.
func validResource() *AzureVirtualNetworkGateway {
	return &AzureVirtualNetworkGateway{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVirtualNetworkGateway",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-vng",
		},
		Spec: &AzureVirtualNetworkGatewaySpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "hub-vpn-gateway",
			Sku:           AzureVirtualNetworkGatewaySku_VPN_GW_1_AZ,
			IpConfigurations: []*AzureVirtualNetworkGatewayIpConfiguration{
				ipConfiguration("primary"),
			},
		},
	}
}

var _ = ginkgo.Describe("AzureVirtualNetworkGatewaySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_virtual_network_gateway", func() {

			ginkgo.It("should not return a validation error for a minimal VPN gateway", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit route-based Generation2 VpnGw2AZ gateway", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayType_VPN
				input.Spec.VpnType = AzureVirtualNetworkGatewayVpnType_ROUTE_BASED
				input.Spec.Generation = AzureVirtualNetworkGatewayGeneration_GENERATION2
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_VPN_GW_2_AZ
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an active-active gateway with two ip configurations and BGP", func() {
				input := validResource()
				input.Spec.ActiveActive = true
				input.Spec.BgpEnabled = true
				input.Spec.IpConfigurations = append(input.Spec.IpConfigurations, ipConfiguration("secondary"))
				input.Spec.BgpSettings = &AzureVirtualNetworkGatewayBgpSettings{
					Asn: 65515,
					PeeringAddresses: []*AzureVirtualNetworkGatewayBgpPeeringAddress{
						{IpConfigurationName: "primary", ApipaAddresses: []string{"169.254.21.4"}},
						{IpConfigurationName: "secondary", ApipaAddresses: []string{"169.254.22.4"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a policy-based BASIC gateway", func() {
				input := validResource()
				input.Spec.VpnType = AzureVirtualNetworkGatewayVpnType_POLICY_BASED
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_BASIC
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an ExpressRoute gateway without public IPs", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayType_EXPRESS_ROUTE
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_ER_GW_1_AZ
				input.Spec.IpConfigurations = []*AzureVirtualNetworkGatewayIpConfiguration{
					{Name: "primary", SubnetId: ref("gateway-subnet")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an ER_GW_SCALE gateway with both scale units", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayType_EXPRESS_ROUTE
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_ER_GW_SCALE
				input.Spec.IpConfigurations = []*AzureVirtualNetworkGatewayIpConfiguration{
					{Name: "primary", SubnetId: ref("gateway-subnet")},
				}
				minUnits, maxUnits := int32(2), int32(10)
				input.Spec.MinimumScaleUnit = &minUnits
				input.Spec.MaximumScaleUnit = &maxUnits
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a point-to-site configuration with Entra ID auth", func() {
				input := validResource()
				input.Spec.VpnClientConfiguration = &AzureVirtualNetworkGatewayVpnClientConfiguration{
					AddressSpaces:      []string{"172.16.0.0/24"},
					AadTenant:          "https://login.microsoftonline.com/tenant-id",
					AadAudience:        "app-client-id",
					AadIssuer:          "https://sts.windows.net/tenant-id/",
					VpnClientProtocols: []string{"OpenVPN"},
					VpnAuthTypes:       []string{"AAD"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept NAT rules with mappings and forced tunneling", func() {
				input := validResource()
				input.Spec.DefaultLocalNetworkGatewayId = ref("hq-site")
				input.Spec.NatRules = []*AzureVirtualNetworkGatewayNatRule{
					{
						Name:             "egress-overlap",
						ExternalMappings: []*AzureVirtualNetworkGatewayNatRuleMapping{{AddressSpace: "100.64.1.0/24"}},
						InternalMappings: []*AzureVirtualNetworkGatewayNatRuleMapping{{AddressSpace: "10.0.1.0/24"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_virtual_network_gateway", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the sku is unspecified", func() {
				input := validResource()
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_azure_virtual_network_gateway_sku_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when no ip configuration exists", func() {
				input := validResource()
				input.Spec.IpConfigurations = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an ip configuration without a subnet", func() {
				input := validResource()
				input.Spec.IpConfigurations[0].SubnetId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a VPN gateway ip configuration without a public IP", func() {
				input := validResource()
				input.Spec.IpConfigurations[0].PublicIpAddressId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an ExpressRoute gateway carrying a public IP", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayType_EXPRESS_ROUTE
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_ER_GW_1_AZ
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a VPN gateway on an ExpressRoute sku", func() {
				input := validResource()
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_ER_GW_1_AZ
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a policy-based gateway above BASIC", func() {
				input := validResource()
				input.Spec.VpnType = AzureVirtualNetworkGatewayVpnType_POLICY_BASED
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_VPN_GW_1_AZ
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a Generation2 gateway on VPN_GW_1_AZ", func() {
				input := validResource()
				input.Spec.Generation = AzureVirtualNetworkGatewayGeneration_GENERATION2
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_VPN_GW_1_AZ
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a Generation1 gateway on VPN_GW_4_AZ", func() {
				input := validResource()
				input.Spec.Generation = AzureVirtualNetworkGatewayGeneration_GENERATION1
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_VPN_GW_4_AZ
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			// Azure retired the non-AZ VpnGw tiers (new creates rejected with
			// NonAzSkusNotAllowedForVPNGateway since 2025-11-01); the values are
			// removed AND reserved, so defined_only rejects their old numbers.
			ginkgo.It("should return a validation error for a retired non-AZ VpnGw sku number", func() {
				input := validResource()
				input.Spec.Sku = AzureVirtualNetworkGatewaySku(5) // formerly VPN_GW_1
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			// The legacy Standard tier remains only for ExpressRoute gateways;
			// its VPN use retired 2025-09-30.
			ginkgo.It("should return a validation error for a VPN gateway on the legacy STANDARD sku", func() {
				input := validResource()
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_STANDARD
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for active-active with one ip configuration", func() {
				input := validResource()
				input.Spec.ActiveActive = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a point-to-site block on an ExpressRoute gateway", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayType_EXPRESS_ROUTE
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_ER_GW_1_AZ
				input.Spec.IpConfigurations = []*AzureVirtualNetworkGatewayIpConfiguration{
					{Name: "primary", SubnetId: ref("gateway-subnet")},
				}
				input.Spec.VpnClientConfiguration = &AzureVirtualNetworkGatewayVpnClientConfiguration{
					AddressSpaces: []string{"172.16.0.0/24"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a partial Entra ID trio", func() {
				input := validResource()
				input.Spec.VpnClientConfiguration = &AzureVirtualNetworkGatewayVpnClientConfiguration{
					AddressSpaces: []string{"172.16.0.0/24"},
					AadTenant:     "https://login.microsoftonline.com/tenant-id",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a radius address without its secret", func() {
				input := validResource()
				input.Spec.VpnClientConfiguration = &AzureVirtualNetworkGatewayVpnClientConfiguration{
					AddressSpaces:       []string{"172.16.0.0/24"},
					RadiusServerAddress: "10.0.0.4",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid vpn client protocol", func() {
				input := validResource()
				input.Spec.VpnClientConfiguration = &AzureVirtualNetworkGatewayVpnClientConfiguration{
					AddressSpaces:      []string{"172.16.0.0/24"},
					VpnClientProtocols: []string{"PPTP"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for scale units without ER_GW_SCALE", func() {
				input := validResource()
				minUnits, maxUnits := int32(2), int32(10)
				input.Spec.MinimumScaleUnit = &minUnits
				input.Spec.MaximumScaleUnit = &maxUnits
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for ER_GW_SCALE without scale units", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayType_EXPRESS_ROUTE
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_ER_GW_SCALE
				input.Spec.IpConfigurations = []*AzureVirtualNetworkGatewayIpConfiguration{
					{Name: "primary", SubnetId: ref("gateway-subnet")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a scale-unit floor above the ceiling", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayType_EXPRESS_ROUTE
				input.Spec.Sku = AzureVirtualNetworkGatewaySku_ER_GW_SCALE
				input.Spec.IpConfigurations = []*AzureVirtualNetworkGatewayIpConfiguration{
					{Name: "primary", SubnetId: ref("gateway-subnet")},
				}
				minUnits, maxUnits := int32(10), int32(2)
				input.Spec.MinimumScaleUnit = &minUnits
				input.Spec.MaximumScaleUnit = &maxUnits
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate NAT rule names", func() {
				input := validResource()
				rule := func() *AzureVirtualNetworkGatewayNatRule {
					return &AzureVirtualNetworkGatewayNatRule{
						Name:             "overlap",
						ExternalMappings: []*AzureVirtualNetworkGatewayNatRuleMapping{{AddressSpace: "100.64.1.0/24"}},
						InternalMappings: []*AzureVirtualNetworkGatewayNatRuleMapping{{AddressSpace: "10.0.1.0/24"}},
					}
				}
				input.Spec.NatRules = []*AzureVirtualNetworkGatewayNatRule{rule(), rule()}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a NAT rule without internal mappings", func() {
				input := validResource()
				input.Spec.NatRules = []*AzureVirtualNetworkGatewayNatRule{
					{
						Name:             "half-rule",
						ExternalMappings: []*AzureVirtualNetworkGatewayNatRuleMapping{{AddressSpace: "100.64.1.0/24"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an APIPA address outside IPv4 form", func() {
				input := validResource()
				input.Spec.BgpSettings = &AzureVirtualNetworkGatewayBgpSettings{
					Asn: 65515,
					PeeringAddresses: []*AzureVirtualNetworkGatewayBgpPeeringAddress{
						{ApipaAddresses: []string{"not-an-ip"}},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
