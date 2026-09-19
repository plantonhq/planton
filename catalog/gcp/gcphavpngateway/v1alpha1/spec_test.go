package gcphavpngatewayv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpHaVpnGatewaySpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

var _ = ginkgo.Describe("GcpHaVpnGatewaySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpHaVpnGateway {
		return &GcpHaVpnGateway{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpHaVpnGateway",
			Metadata: &shared.CloudResourceMetadata{
				Name: "hub-vpn",
			},
			Spec: &GcpHaVpnGatewaySpec{
				Region:  "us-central1",
				Network: litRef("projects/acme-net/global/networks/hub"),
				Router: &GcpHaVpnGatewayRouter{
					Bgp: &GcpHaVpnGatewayRouterBgp{Asn: 64514},
				},
			},
		}
	}

	expectError := func(target *GcpHaVpnGateway, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept the minimal internet-facing gateway with a 16-bit private ASN", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every knob set: references, dual-stack, custom advertisement, labels, tags", func() {
		target := minimal()
		target.Spec.ProjectId = nameRef("network-host")
		target.Spec.Network = nameRef("hub")
		target.Spec.GatewayName = "hub-vpn-gw"
		target.Spec.Description = "Hub VPN to every branch"
		target.Spec.GatewayIpVersion = "IPV4"
		target.Spec.StackType = "IPV4_IPV6"
		target.Spec.Labels = map[string]string{"team": "network"}
		target.Spec.ResourceManagerTags = map[string]string{"tagKeys/1": "tagValues/2"}
		target.Spec.Router = &GcpHaVpnGatewayRouter{
			Name:        "hub-vpn-router",
			Description: "BGP for the hub VPN",
			Bgp: &GcpHaVpnGatewayRouterBgp{
				Asn:              4200000001,
				AdvertiseMode:    "CUSTOM",
				AdvertisedGroups: []string{"ALL_SUBNETS"},
				AdvertisedIpRanges: []*GcpHaVpnGatewayRouterBgpAdvertisedIpRange{
					{Range: "10.10.0.0/16", Description: "shared services"},
				},
				KeepaliveInterval: 30,
				IdentifierRange:   "169.254.8.0/30",
			},
			ResourceManagerTags: map[string]string{"tagKeys/1": "tagValues/2"},
		}
		target.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept HA VPN over Interconnect: pinned interfaces on an encrypted router", func() {
		target := minimal()
		target.Spec.Router.EncryptedInterconnectRouter = true
		target.Spec.VpnInterfaces = []*GcpHaVpnGatewayVpnInterface{
			{Id: 0, InterconnectAttachment: "projects/acme-net/regions/us-central1/interconnectAttachments/attach-a"},
			{Id: 1, InterconnectAttachment: "projects/acme-net/regions/us-central1/interconnectAttachments/attach-b"},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every enum value and the empty default", func() {
		for _, v := range []string{"", "IPV4", "IPV6"} {
			target := minimal()
			target.Spec.GatewayIpVersion = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), v)
		}
		for _, v := range []string{"", "IPV4_ONLY", "IPV4_IPV6", "IPV6_ONLY"} {
			target := minimal()
			target.Spec.StackType = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), v)
		}
		for _, v := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), v)
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a gateway without its region, network, or router", func() {
		target := minimal()
		target.Spec.Region = ""
		expectError(target, "region")

		target = minimal()
		target.Spec.Network = nil
		expectError(target, "network")

		target = minimal()
		target.Spec.Router = nil
		expectError(target, "router")

		target = minimal()
		target.Spec.Router.Bgp = nil
		expectError(target, "bgp")
	})

	ginkgo.It("should reject an ASN outside the private ranges", func() {
		for _, asn := range []uint32{0, 1, 64511, 65535, 4199999999} {
			target := minimal()
			target.Spec.Router.Bgp.Asn = asn
			if asn == 0 {
				expectError(target, "asn")
			} else {
				expectError(target, "asn must be a private ASN")
			}
		}
	})

	ginkgo.It("should reject custom advertisement in DEFAULT mode and a keepalive outside 20-60", func() {
		target := minimal()
		target.Spec.Router.Bgp.AdvertisedGroups = []string{"ALL_SUBNETS"}
		expectError(target, "advertised_groups can only be set when advertise_mode is CUSTOM")

		target = minimal()
		target.Spec.Router.Bgp.AdvertisedIpRanges = []*GcpHaVpnGatewayRouterBgpAdvertisedIpRange{{Range: "10.0.0.0/8"}}
		expectError(target, "advertised_ip_ranges can only be set when advertise_mode is CUSTOM")

		target = minimal()
		target.Spec.Router.Bgp.KeepaliveInterval = 10
		expectError(target, "keepalive_interval must be between 20 and 60")
	})

	ginkgo.It("should reject Interconnect interfaces on a non-encrypted router, a duplicate interface id, or three interfaces", func() {
		target := minimal()
		target.Spec.VpnInterfaces = []*GcpHaVpnGatewayVpnInterface{{Id: 0, InterconnectAttachment: "attach-a"}}
		expectError(target, "requires router.encrypted_interconnect_router = true")

		target = minimal()
		target.Spec.Router.EncryptedInterconnectRouter = true
		target.Spec.VpnInterfaces = []*GcpHaVpnGatewayVpnInterface{
			{Id: 0, InterconnectAttachment: "attach-a"},
			{Id: 0, InterconnectAttachment: "attach-b"},
		}
		expectError(target, "each vpn_interfaces entry must name a different interface id")

		target = minimal()
		target.Spec.Router.EncryptedInterconnectRouter = true
		target.Spec.VpnInterfaces = []*GcpHaVpnGatewayVpnInterface{
			{Id: 0, InterconnectAttachment: "a"}, {Id: 1, InterconnectAttachment: "b"}, {Id: 1, InterconnectAttachment: "c"},
		}
		expectError(target, "vpn_interfaces")
	})

	ginkgo.It("should reject malformed names and unknown enum values", func() {
		target := minimal()
		target.Spec.GatewayName = "Hub-VPN"
		expectError(target, "gateway_name must be 1-63 characters")

		target = minimal()
		target.Spec.Router.Name = "router_1"
		expectError(target, "router name must be 1-63 characters")

		target = minimal()
		target.Spec.GatewayIpVersion = "DUAL"
		expectError(target, "gateway_ip_version must be IPV4 or IPV6")

		target = minimal()
		target.Spec.StackType = "IPV4"
		expectError(target, "stack_type must be IPV4_ONLY, IPV4_IPV6, or IPV6_ONLY")

		target = minimal()
		target.Spec.DeletionPolicy = "RETAIN"
		expectError(target, "deletion_policy must be one of")
	})
})
