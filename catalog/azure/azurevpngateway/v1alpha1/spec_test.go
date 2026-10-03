package azurevpngatewayv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureVpnGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVpnGatewaySpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// int32Ptr returns a pointer to the given int32 (for optional fields).
func int32Ptr(value int32) *int32 {
	return &value
}

// routingPreferencePtr returns a pointer to the given routing
// preference (the field is optional, so the generated Go type is a
// pointer).
func routingPreferencePtr(value AzureVpnGatewayRoutingPreference) *AzureVpnGatewayRoutingPreference {
	return &value
}

const testHubId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus"

// validResource returns a minimal valid VPN gateway that individual
// cases mutate into the shape under test.
func validResource() *AzureVpnGateway {
	return &AzureVpnGateway{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVpnGateway",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-vpn-gateway",
		},
		Spec: &AzureVpnGatewaySpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "hub-vpn-gateway",
			VirtualHubId:  literal(testHubId),
		},
	}
}

// validNatRule returns a complete static egress NAT rule.
func validNatRule(name string) *AzureVpnGatewayNatRule {
	return &AzureVpnGatewayNatRule{
		Name:             name,
		ExternalMappings: []*AzureVpnGatewayNatRuleMapping{{AddressSpace: "192.168.100.0/24"}},
		InternalMappings: []*AzureVpnGatewayNatRuleMapping{{AddressSpace: "10.60.0.0/24"}},
	}
}

var _ = ginkgo.Describe("AzureVpnGatewaySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_vpn_gateway", func() {

			ginkgo.It("should not return a validation error for a minimal gateway", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit routing preference and scale unit", func() {
				input := validResource()
				input.Spec.RoutingPreference = routingPreferencePtr(AzureVpnGatewayRoutingPreference_INTERNET)
				input.Spec.ScaleUnit = int32Ptr(3)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept BGP settings with custom APIPA addresses on both instances", func() {
				input := validResource()
				input.Spec.BgpSettings = &AzureVpnGatewayBgpSettings{
					Asn:        65515,
					PeerWeight: 10,
					Instance_0BgpPeeringAddress: &AzureVpnGatewayInstanceBgpPeeringAddress{
						CustomIps: []string{"169.254.21.5"},
					},
					Instance_1BgpPeeringAddress: &AzureVpnGatewayInstanceBgpPeeringAddress{
						CustomIps: []string{"169.254.22.5"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept NAT rules of both modes and types", func() {
				input := validResource()
				ingress := validNatRule("branch-ingress")
				ingress.Mode = AzureVpnGatewayNatRuleMode_INGRESS_SNAT
				dynamic := validNatRule("dynamic-egress")
				dynamic.Type = AzureVpnGatewayNatRuleType_DYNAMIC_NAT
				dynamic.ExternalMappings = []*AzureVpnGatewayNatRuleMapping{
					{AddressSpace: "192.168.101.0/26", PortRange: "1024-65535"},
				}
				input.Spec.NatRules = []*AzureVpnGatewayNatRule{validNatRule("static-egress"), ingress, dynamic}
				input.Spec.BgpRouteTranslationForNatEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a NAT rule pinned to one instance", func() {
				input := validResource()
				rule := validNatRule("instance-pinned")
				rule.IpConfiguration = AzureVpnGatewayNatRuleIpConfiguration_INSTANCE_1
				input.Spec.NatRules = []*AzureVpnGatewayNatRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept scale unit 0 (the provider's floor)", func() {
				input := validResource()
				input.Spec.ScaleUnit = int32Ptr(0)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_vpn_gateway", func() {

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

			ginkgo.It("should return a validation error when virtual_hub_id is missing", func() {
				input := validResource()
				input.Spec.VirtualHubId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an undefined routing preference", func() {
				input := validResource()
				input.Spec.RoutingPreference = routingPreferencePtr(AzureVpnGatewayRoutingPreference(99))
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a negative scale unit", func() {
				input := validResource()
				input.Spec.ScaleUnit = int32Ptr(-1)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a peer weight above 100", func() {
				input := validResource()
				input.Spec.BgpSettings = &AzureVpnGatewayBgpSettings{Asn: 65515, PeerWeight: 101}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for ASN 0", func() {
				input := validResource()
				input.Spec.BgpSettings = &AzureVpnGatewayBgpSettings{Asn: 0, PeerWeight: 0}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an instance address block with no custom IPs", func() {
				input := validResource()
				input.Spec.BgpSettings = &AzureVpnGatewayBgpSettings{
					Asn:                         65515,
					PeerWeight:                  0,
					Instance_0BgpPeeringAddress: &AzureVpnGatewayInstanceBgpPeeringAddress{},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a custom IP that is not IPv4", func() {
				input := validResource()
				input.Spec.BgpSettings = &AzureVpnGatewayBgpSettings{
					Asn:        65515,
					PeerWeight: 0,
					Instance_0BgpPeeringAddress: &AzureVpnGatewayInstanceBgpPeeringAddress{
						CustomIps: []string{"2001:db8::1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate NAT rule names", func() {
				input := validResource()
				input.Spec.NatRules = []*AzureVpnGatewayNatRule{
					validNatRule("overlap"),
					validNatRule("overlap"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a NAT rule without external mappings", func() {
				input := validResource()
				rule := validNatRule("half-mapped")
				rule.ExternalMappings = nil
				input.Spec.NatRules = []*AzureVpnGatewayNatRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a NAT rule without internal mappings", func() {
				input := validResource()
				rule := validNatRule("half-mapped")
				rule.InternalMappings = nil
				input.Spec.NatRules = []*AzureVpnGatewayNatRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a NAT mapping without an address space", func() {
				input := validResource()
				rule := validNatRule("empty-mapping")
				rule.ExternalMappings = []*AzureVpnGatewayNatRuleMapping{{}}
				input.Spec.NatRules = []*AzureVpnGatewayNatRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an undefined NAT rule mode", func() {
				input := validResource()
				rule := validNatRule("bad-mode")
				rule.Mode = AzureVpnGatewayNatRuleMode(99)
				input.Spec.NatRules = []*AzureVpnGatewayNatRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
