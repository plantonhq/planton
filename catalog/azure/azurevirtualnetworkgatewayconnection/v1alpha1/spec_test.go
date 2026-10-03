package azurevirtualnetworkgatewayconnectionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureVirtualNetworkGatewayConnectionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVirtualNetworkGatewayConnectionSpec Validation Tests")
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

// validIpsecPolicy returns a complete custom IPsec proposal.
func validIpsecPolicy() *AzureVirtualNetworkGatewayConnectionIpsecPolicy {
	return &AzureVirtualNetworkGatewayConnectionIpsecPolicy{
		DhGroup:         "DHGroup14",
		IkeEncryption:   "AES256",
		IkeIntegrity:    "SHA256",
		IpsecEncryption: "AES256",
		IpsecIntegrity:  "SHA256",
		PfsGroup:        "PFS2048",
	}
}

// validResource returns a minimal valid site-to-site (IPSEC) connection
// that individual cases mutate into the shape under test.
func validResource() *AzureVirtualNetworkGatewayConnection {
	return &AzureVirtualNetworkGatewayConnection{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVirtualNetworkGatewayConnection",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-vngc",
		},
		Spec: &AzureVirtualNetworkGatewayConnectionSpec{
			Region:                  "eastus",
			ResourceGroup:           literal("test-rg"),
			Name:                    "hq-to-azure",
			Type:                    AzureVirtualNetworkGatewayConnectionType_IPSEC,
			VirtualNetworkGatewayId: ref("hub-vpn-gateway"),
			LocalNetworkGatewayId:   ref("hq-site"),
			SharedKey:               literal("test-pre-shared-key"),
		},
	}
}

var _ = ginkgo.Describe("AzureVirtualNetworkGatewayConnectionSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_virtual_network_gateway_connection", func() {

			ginkgo.It("should not return a validation error for a minimal site-to-site connection", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a connection without a shared key (Azure generates one)", func() {
				input := validResource()
				input.Spec.SharedKey = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a custom IPsec policy with DPD and BGP", func() {
				input := validResource()
				input.Spec.BgpEnabled = true
				dpd := int32(45)
				input.Spec.DpdTimeoutSeconds = &dpd
				input.Spec.IpsecPolicy = validIpsecPolicy()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept policy-based traffic selectors with a custom policy", func() {
				input := validResource()
				input.Spec.UsePolicyBasedTrafficSelectors = true
				input.Spec.IpsecPolicy = validIpsecPolicy()
				input.Spec.TrafficSelectorPolicies = []*AzureVirtualNetworkGatewayConnectionTrafficSelectorPolicy{
					{
						LocalAddressCidrs:  []string{"10.0.0.0/16"},
						RemoteAddressCidrs: []string{"192.168.100.0/24"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a VNet-to-VNet connection with a peer gateway", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayConnectionType_VNET_TO_VNET
				input.Spec.LocalNetworkGatewayId = nil
				input.Spec.PeerVirtualNetworkGatewayId = ref("spoke-vpn-gateway")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an ExpressRoute connection with FastPath", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayConnectionType_EXPRESS_ROUTE
				input.Spec.LocalNetworkGatewayId = nil
				input.Spec.SharedKey = nil
				input.Spec.ExpressRouteCircuitId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/expressRouteCircuits/er1")
				input.Spec.ExpressRouteGatewayBypass = true
				input.Spec.PrivateLinkFastPathEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept custom APIPA BGP addresses on a BGP IPsec connection", func() {
				input := validResource()
				input.Spec.BgpEnabled = true
				input.Spec.CustomBgpAddresses = &AzureVirtualNetworkGatewayConnectionCustomBgpAddresses{
					Primary:   "169.254.21.4",
					Secondary: "169.254.22.4",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept NAT rule references", func() {
				input := validResource()
				input.Spec.EgressNatRuleIds = []*foreignkeyv1.StringValueOrRef{
					literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworkGateways/gw/natRules/egress-overlap"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_virtual_network_gateway_connection", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the type is unspecified", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayConnectionType_azure_virtual_network_gateway_connection_type_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the gateway reference is missing", func() {
				input := validResource()
				input.Spec.VirtualNetworkGatewayId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for IPSEC without a local network gateway", func() {
				input := validResource()
				input.Spec.LocalNetworkGatewayId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for VNet-to-VNet without a peer gateway", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayConnectionType_VNET_TO_VNET
				input.Spec.LocalNetworkGatewayId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for ExpressRoute without a circuit", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayConnectionType_EXPRESS_ROUTE
				input.Spec.LocalNetworkGatewayId = nil
				input.Spec.SharedKey = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a shared key on an ExpressRoute connection", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayConnectionType_EXPRESS_ROUTE
				input.Spec.LocalNetworkGatewayId = nil
				input.Spec.ExpressRouteCircuitId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/expressRouteCircuits/er1")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an authorization key on an IPsec connection", func() {
				input := validResource()
				input.Spec.AuthorizationKey = literal("auth-key")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for custom BGP addresses without BGP", func() {
				input := validResource()
				input.Spec.CustomBgpAddresses = &AzureVirtualNetworkGatewayConnectionCustomBgpAddresses{
					Primary: "169.254.21.4",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for custom BGP addresses on a VNet-to-VNet connection", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayConnectionType_VNET_TO_VNET
				input.Spec.LocalNetworkGatewayId = nil
				input.Spec.PeerVirtualNetworkGatewayId = ref("spoke-vpn-gateway")
				input.Spec.BgpEnabled = true
				input.Spec.CustomBgpAddresses = &AzureVirtualNetworkGatewayConnectionCustomBgpAddresses{
					Primary: "169.254.21.4",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for FastPath without gateway bypass", func() {
				input := validResource()
				input.Spec.Type = AzureVirtualNetworkGatewayConnectionType_EXPRESS_ROUTE
				input.Spec.LocalNetworkGatewayId = nil
				input.Spec.SharedKey = nil
				input.Spec.ExpressRouteCircuitId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/expressRouteCircuits/er1")
				input.Spec.PrivateLinkFastPathEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for policy-based selectors without an ipsec policy", func() {
				input := validResource()
				input.Spec.UsePolicyBasedTrafficSelectors = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an unknown IKE encryption algorithm", func() {
				input := validResource()
				policy := validIpsecPolicy()
				policy.IkeEncryption = "ROT13"
				input.Spec.IpsecPolicy = policy
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a too-small sa_datasize", func() {
				input := validResource()
				policy := validIpsecPolicy()
				size := int32(512)
				policy.SaDatasize = &size
				input.Spec.IpsecPolicy = policy
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a traffic selector without remote CIDRs", func() {
				input := validResource()
				input.Spec.TrafficSelectorPolicies = []*AzureVirtualNetworkGatewayConnectionTrafficSelectorPolicy{
					{LocalAddressCidrs: []string{"10.0.0.0/16"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range routing weight", func() {
				input := validResource()
				weight := int32(50000)
				input.Spec.RoutingWeight = &weight
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
