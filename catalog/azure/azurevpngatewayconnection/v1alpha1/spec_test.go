package azurevpngatewayconnectionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureVpnGatewayConnectionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVpnGatewayConnectionSpec Validation Tests")
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

const (
	testGatewayId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/vpnGateways/hub-vpn-gateway"
	testSiteId     = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/vpnSites/branch-london"
	testSiteLinkId = testSiteId + "/vpnSiteLinks/primary-isp"
	testHubId      = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/virtualHubs/hub-eastus"
)

// validLink returns a minimal valid tunnel pinned to the test site
// link.
func validLink(name string) *AzureVpnGatewayConnectionLink {
	return &AzureVpnGatewayConnectionLink{
		Name:          name,
		VpnSiteLinkId: literal(testSiteLinkId),
	}
}

// validIpsecPolicy returns a complete pinned proposal (the provider
// requires every field of a configured policy).
func validIpsecPolicy() *AzureVpnGatewayConnectionIpsecPolicy {
	return &AzureVpnGatewayConnectionIpsecPolicy{
		SaLifetimeSec:          3600,
		SaDataSizeKb:           102400000,
		EncryptionAlgorithm:    "AES256",
		IntegrityAlgorithm:     "SHA256",
		IkeEncryptionAlgorithm: "AES256",
		IkeIntegrityAlgorithm:  "SHA256",
		DhGroup:                "DHGroup14",
		PfsGroup:               "PFS2048",
	}
}

// validResource returns a minimal valid connection that individual
// cases mutate into the shape under test.
func validResource() *AzureVpnGatewayConnection {
	return &AzureVpnGatewayConnection{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVpnGatewayConnection",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-vpn-connection",
		},
		Spec: &AzureVpnGatewayConnectionSpec{
			Name:            "branch-london",
			VpnGatewayId:    literal(testGatewayId),
			RemoteVpnSiteId: literal(testSiteId),
			VpnLinks:        []*AzureVpnGatewayConnectionLink{validLink("primary-isp")},
		},
	}
}

var _ = ginkgo.Describe("AzureVpnGatewayConnectionSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_vpn_gateway_connection", func() {

			ginkgo.It("should not return a validation error for a minimal connection", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept two links (the active-active shape) with per-link parameters", func() {
				input := validResource()
				second := validLink("backup-isp")
				second.VpnSiteLinkId = literal(testSiteId + "/vpnSiteLinks/backup-isp")
				second.BandwidthMbps = int32Ptr(50)
				second.Protocol = AzureVpnGatewayConnectionProtocol_IKE_V1
				second.ConnectionMode = AzureVpnGatewayConnectionMode_RESPONDER_ONLY
				second.RouteWeight = 10
				input.Spec.VpnLinks = append(input.Spec.VpnLinks, second)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a BGP link with custom APIPA addresses and DPD", func() {
				input := validResource()
				link := input.Spec.VpnLinks[0]
				link.BgpEnabled = true
				link.DpdTimeoutSeconds = int32Ptr(45)
				link.CustomBgpAddresses = []*AzureVpnGatewayConnectionCustomBgpAddress{
					{IpAddress: "169.254.21.5", IpConfigurationId: "Instance0"},
					{IpAddress: "169.254.22.5", IpConfigurationId: "Instance1"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a pinned IPsec proposal with a shared key reference", func() {
				input := validResource()
				link := input.Spec.VpnLinks[0]
				link.SharedKey = literal("s3cret-psk")
				link.IpsecPolicies = []*AzureVpnGatewayConnectionIpsecPolicy{validIpsecPolicy()}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept policy-based traffic selectors when a proposal is pinned", func() {
				input := validResource()
				link := input.Spec.VpnLinks[0]
				link.PolicyBasedTrafficSelectorEnabled = true
				link.IpsecPolicies = []*AzureVpnGatewayConnectionIpsecPolicy{validIpsecPolicy()}
				input.Spec.TrafficSelectorPolicies = []*AzureVpnGatewayConnectionTrafficSelectorPolicy{
					{
						LocalAddressCidrs:  []string{"10.60.0.0/16"},
						RemoteAddressCidrs: []string{"192.168.10.0/24"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a routing block with association, maps, and propagation", func() {
				input := validResource()
				input.Spec.Routing = &AzureVpnGatewayConnectionRouting{
					AssociatedRouteTableId: literal(testHubId + "/hubRouteTables/defaultRouteTable"),
					InboundRouteMapId:      literal(testHubId + "/routeMaps/ingress"),
					OutboundRouteMapId:     literal(testHubId + "/routeMaps/egress"),
					PropagatedRouteTable: &AzureVpnGatewayConnectionPropagatedRouteTable{
						RouteTableIds: []*foreignkeyv1.StringValueOrRef{
							literal(testHubId + "/hubRouteTables/defaultRouteTable"),
						},
						Labels: []string{"default"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept NAT rule references and internet security", func() {
				input := validResource()
				input.Spec.InternetSecurityEnabled = true
				link := input.Spec.VpnLinks[0]
				link.EgressNatRuleIds = []*foreignkeyv1.StringValueOrRef{
					literal(testGatewayId + "/natRules/branch-overlap"),
				}
				link.RatelimitEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_vpn_gateway_connection", func() {

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vpn_gateway_id is missing", func() {
				input := validResource()
				input.Spec.VpnGatewayId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when remote_vpn_site_id is missing", func() {
				input := validResource()
				input.Spec.RemoteVpnSiteId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vpn_links is empty", func() {
				input := validResource()
				input.Spec.VpnLinks = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a link without a site link reference", func() {
				input := validResource()
				input.Spec.VpnLinks[0].VpnSiteLinkId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for duplicate link names", func() {
				input := validResource()
				input.Spec.VpnLinks = []*AzureVpnGatewayConnectionLink{
					validLink("primary-isp"),
					validLink("primary-isp"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for policy-based selectors without a pinned proposal", func() {
				input := validResource()
				input.Spec.VpnLinks[0].PolicyBasedTrafficSelectorEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a routing block without an associated route table", func() {
				input := validResource()
				input.Spec.Routing = &AzureVpnGatewayConnectionRouting{
					InboundRouteMapId: literal(testHubId + "/routeMaps/ingress"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for propagation without route table ids", func() {
				input := validResource()
				input.Spec.Routing = &AzureVpnGatewayConnectionRouting{
					AssociatedRouteTableId: literal(testHubId + "/hubRouteTables/defaultRouteTable"),
					PropagatedRouteTable:   &AzureVpnGatewayConnectionPropagatedRouteTable{Labels: []string{"default"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a DPD timeout below 9 seconds", func() {
				input := validResource()
				input.Spec.VpnLinks[0].DpdTimeoutSeconds = int32Ptr(5)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a zero bandwidth", func() {
				input := validResource()
				input.Spec.VpnLinks[0].BandwidthMbps = int32Ptr(0)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an IPsec proposal with an unknown cipher", func() {
				input := validResource()
				policy := validIpsecPolicy()
				policy.EncryptionAlgorithm = "AES512"
				input.Spec.VpnLinks[0].IpsecPolicies = []*AzureVpnGatewayConnectionIpsecPolicy{policy}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an IPsec proposal with a lifetime below 300s", func() {
				input := validResource()
				policy := validIpsecPolicy()
				policy.SaLifetimeSec = 60
				input.Spec.VpnLinks[0].IpsecPolicies = []*AzureVpnGatewayConnectionIpsecPolicy{policy}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a custom BGP address on an unknown instance", func() {
				input := validResource()
				input.Spec.VpnLinks[0].CustomBgpAddresses = []*AzureVpnGatewayConnectionCustomBgpAddress{
					{IpAddress: "169.254.21.5", IpConfigurationId: "Instance2"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a traffic selector without remote ranges", func() {
				input := validResource()
				input.Spec.TrafficSelectorPolicies = []*AzureVpnGatewayConnectionTrafficSelectorPolicy{
					{LocalAddressCidrs: []string{"10.60.0.0/16"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
