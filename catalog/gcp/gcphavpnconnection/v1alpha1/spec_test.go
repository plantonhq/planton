package gcphavpnconnectionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpHaVpnConnectionSpec Suite")
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

var _ = ginkgo.Describe("GcpHaVpnConnectionSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	tunnel := func(name string, gwIface int32, extIface *int32, cidr string) *GcpHaVpnConnectionTunnel {
		return &GcpHaVpnConnectionTunnel{
			Name:                         name,
			VpnGatewayInterface:          gwIface,
			PeerExternalGatewayInterface: extIface,
			SharedSecret:                 "correct-horse-battery-staple",
			BgpSession: &GcpHaVpnConnectionBgpSession{
				InterfaceIpRange: cidr,
				PeerAsn:          65001,
			},
		}
	}

	// Two tunnels to a two-address on-premises device -- Google's 99.99% shape.
	toOnprem := func() *GcpHaVpnConnection {
		return &GcpHaVpnConnection{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpHaVpnConnection",
			Metadata: &shared.CloudResourceMetadata{
				Name: "hq",
			},
			Spec: &GcpHaVpnConnectionSpec{
				Gateway: litRef("projects/acme-net/regions/us-central1/vpnGateways/hub-vpn"),
				Router:  litRef("hub-vpn"),
				Region:  litRef("us-central1"),
				Peer: &GcpHaVpnConnectionPeer{
					ExternalGateway: &GcpHaVpnConnectionExternalGateway{
						RedundancyType: "TWO_IPS_REDUNDANCY",
						Interfaces: []*GcpHaVpnConnectionExternalGatewayInterface{
							{Id: 0, IpAddress: "203.0.113.10"},
							{Id: 1, IpAddress: "203.0.113.11"},
						},
					},
				},
				Tunnels: []*GcpHaVpnConnectionTunnel{
					tunnel("hq-tunnel-0", 0, proto.Int32(0), "169.254.10.1/30"),
					tunnel("hq-tunnel-1", 1, proto.Int32(1), "169.254.11.1/30"),
				},
			},
		}
	}

	// Two tunnels to another Google Cloud gateway: interfaces pair themselves.
	toGcp := func() *GcpHaVpnConnection {
		target := toOnprem()
		target.Metadata.Name = "to-spoke"
		target.Spec.Peer = &GcpHaVpnConnectionPeer{GcpGateway: nameRef("spoke-vpn")}
		target.Spec.Tunnels = []*GcpHaVpnConnectionTunnel{
			tunnel("to-spoke-0", 0, nil, "169.254.20.1/30"),
			tunnel("to-spoke-1", 1, nil, "169.254.21.1/30"),
		}
		return target
	}

	expectError := func(target *GcpHaVpnConnection, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept two tunnels to a two-address on-premises device", func() {
		gomega.Expect(validator.Validate(toOnprem())).To(gomega.Succeed())
	})

	ginkgo.It("should accept two tunnels to another Google Cloud gateway with the gateway trio by reference", func() {
		target := toGcp()
		target.Spec.Gateway = nameRef("hub-vpn")
		target.Spec.Router = nameRef("hub-vpn")
		target.Spec.Region = nameRef("hub-vpn")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept one tunnel to a single-address device and four to a four-address device", func() {
		target := toOnprem()
		target.Spec.Peer.ExternalGateway.RedundancyType = "SINGLE_IP_INTERNALLY_REDUNDANT"
		target.Spec.Peer.ExternalGateway.Interfaces = target.Spec.Peer.ExternalGateway.Interfaces[:1]
		target.Spec.Tunnels = target.Spec.Tunnels[:1]
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())

		target = toOnprem()
		target.Spec.Peer.ExternalGateway.RedundancyType = "FOUR_IPS_REDUNDANCY"
		target.Spec.Peer.ExternalGateway.Interfaces = []*GcpHaVpnConnectionExternalGatewayInterface{
			{Id: 0, IpAddress: "203.0.113.10"}, {Id: 1, IpAddress: "203.0.113.11"},
			{Id: 2, IpAddress: "203.0.113.12"}, {Id: 3, IpAddress: "203.0.113.13"},
		}
		target.Spec.Tunnels = []*GcpHaVpnConnectionTunnel{
			tunnel("t0", 0, proto.Int32(0), "169.254.10.1/30"),
			tunnel("t1", 0, proto.Int32(1), "169.254.11.1/30"),
			tunnel("t2", 1, proto.Int32(2), "169.254.12.1/30"),
			tunnel("t3", 1, proto.Int32(3), "169.254.13.1/30"),
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every tunnel and session knob set", func() {
		target := toOnprem()
		target.Spec.ProjectId = nameRef("network-host")
		target.Spec.ResourceManagerTags = map[string]string{"tagKeys/1": "tagValues/2"}
		target.Spec.DeletionPolicy = "PREVENT"
		target.Spec.Peer.ExternalGateway.Name = "hq-device"
		target.Spec.Peer.ExternalGateway.Description = "HQ Cisco ASA pair"
		target.Spec.Peer.ExternalGateway.Labels = map[string]string{"site": "hq"}
		t0 := target.Spec.Tunnels[0]
		t0.IkeVersion = proto.Int32(2)
		t0.LocalTrafficSelector = []string{"10.0.0.0/8"}
		t0.RemoteTrafficSelector = []string{"192.168.0.0/16"}
		t0.CipherSuite = &GcpHaVpnConnectionCipherSuite{
			Phase1: &GcpHaVpnConnectionCipherSuitePhase1{Encryption: []string{"AES-GCM-16-256"}, Integrity: []string{"HMAC-SHA2-256-128"}, Prf: []string{"PRF-HMAC-SHA2-256"}, Dh: []string{"GROUP-14"}},
			Phase2: &GcpHaVpnConnectionCipherSuitePhase2{Encryption: []string{"AES-GCM-16-256"}, Integrity: []string{"HMAC-SHA2-256-128"}, Pfs: []string{"GROUP-14"}},
		}
		t0.Labels = map[string]string{"tunnel": "primary"}
		t0.BgpSession = &GcpHaVpnConnectionBgpSession{
			Name:                        "hq-session-0",
			InterfaceIpRange:            "169.254.10.1/30",
			IpVersion:                   proto.String("IPV4"),
			PeerAsn:                     65001,
			PeerIpAddress:               proto.String("169.254.10.2"),
			AdvertisedRoutePriority:     proto.Int32(100),
			AdvertiseMode:               "CUSTOM",
			AdvertisedGroups:            []string{"ALL_SUBNETS"},
			AdvertisedIpRanges:          []*GcpHaVpnConnectionBgpAdvertisedIpRange{{Range: "10.10.0.0/16", Description: "shared"}},
			Enable:                      proto.Bool(true),
			EnableIpv4:                  proto.Bool(true),
			EnableIpv6:                  false,
			CustomLearnedIpRanges:       []*GcpHaVpnConnectionBgpCustomLearnedIpRange{{Range: "192.168.50.0/24"}},
			CustomLearnedRoutePriority:  proto.Int32(0),
			Bfd:                         &GcpHaVpnConnectionBgpBfd{SessionInitializationMode: "ACTIVE", MinReceiveInterval: 1000, MinTransmitInterval: 1000, Multiplier: 5},
			Md5AuthenticationKey:        &GcpHaVpnConnectionBgpMd5AuthenticationKey{Name: "hq-key-0", Key: "s3cret"},
			ImportPolicies:              []string{"accept-hq"},
			ExportPolicies:              []string{"export-shared"},
			Ipv4NexthopAddress:          proto.String("169.254.10.1"),
			PeerIpv4NexthopAddress:      proto.String("169.254.10.2"),
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a dual-stack session with IPv6 next hops", func() {
		target := toGcp()
		s := target.Spec.Tunnels[0].BgpSession
		s.EnableIpv6 = true
		s.Ipv6NexthopAddress = proto.String("2600:2d00:0:2::1")
		s.PeerIpv6NexthopAddress = proto.String("2600:2d00:0:2::2")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every deletion_policy value and the empty default", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target := toGcp()
			target.Spec.DeletionPolicy = policy
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed(), policy)
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a connection missing its gateway, router, region, peer, or tunnels", func() {
		target := toGcp()
		target.Spec.Gateway = nil
		expectError(target, "gateway")

		target = toGcp()
		target.Spec.Router = nil
		expectError(target, "router")

		target = toGcp()
		target.Spec.Region = nil
		expectError(target, "region")

		target = toGcp()
		target.Spec.Peer = nil
		expectError(target, "peer")

		target = toGcp()
		target.Spec.Tunnels = nil
		expectError(target, "tunnels")
	})

	ginkgo.It("should reject a peer with both arms or neither", func() {
		target := toOnprem()
		target.Spec.Peer.GcpGateway = nameRef("spoke-vpn")
		expectError(target, "set exactly one of external_gateway")

		target = toGcp()
		target.Spec.Peer = &GcpHaVpnConnectionPeer{}
		expectError(target, "set exactly one of external_gateway")
	})

	ginkgo.It("should reject an external gateway whose interface count does not match its redundancy type", func() {
		target := toOnprem()
		target.Spec.Peer.ExternalGateway.Interfaces = target.Spec.Peer.ExternalGateway.Interfaces[:1]
		target.Spec.Tunnels = target.Spec.Tunnels[:1]
		expectError(target, "interfaces must match redundancy_type")
	})

	ginkgo.It("should reject duplicate or out-of-range external interface ids and an interface with no or two addresses", func() {
		target := toOnprem()
		target.Spec.Peer.ExternalGateway.Interfaces[1].Id = 0
		expectError(target, "interface ids must be unique")

		target = toOnprem()
		target.Spec.Peer.ExternalGateway.Interfaces[1].Id = 3
		expectError(target, "interface ids must be unique and below the interface count")

		target = toOnprem()
		target.Spec.Peer.ExternalGateway.Interfaces[0].IpAddress = ""
		expectError(target, "exactly one of ip_address or ipv6_address")

		target = toOnprem()
		target.Spec.Peer.ExternalGateway.Interfaces[0].Ipv6Address = "2001:db8::1"
		expectError(target, "exactly one of ip_address or ipv6_address")
	})

	ginkgo.It("should reject tunnels that name an external interface with a Google peer, or omit it with an external peer", func() {
		target := toGcp()
		target.Spec.Tunnels[0].PeerExternalGatewayInterface = proto.Int32(0)
		expectError(target, "with peer.gcp_gateway none does")

		target = toOnprem()
		target.Spec.Tunnels[0].PeerExternalGatewayInterface = nil
		expectError(target, "with peer.external_gateway every tunnel sets peer_external_gateway_interface")
	})

	ginkgo.It("should reject a tunnel naming an external interface the gateway does not declare", func() {
		target := toOnprem()
		target.Spec.Tunnels[1].PeerExternalGatewayInterface = proto.Int32(3)
		expectError(target, "must name an interface declared in peer.external_gateway.interfaces")
	})

	ginkgo.It("should reject duplicate tunnel names and duplicate session ranges", func() {
		target := toGcp()
		target.Spec.Tunnels[1].Name = target.Spec.Tunnels[0].Name
		expectError(target, "each tunnel must have a distinct name")

		target = toGcp()
		target.Spec.Tunnels[1].BgpSession.InterfaceIpRange = target.Spec.Tunnels[0].BgpSession.InterfaceIpRange
		expectError(target, "must be a different /30")
	})

	ginkgo.It("should reject a tunnel without a secret, a session, or a valid name", func() {
		target := toGcp()
		target.Spec.Tunnels[0].SharedSecret = ""
		expectError(target, "shared_secret")

		target = toGcp()
		target.Spec.Tunnels[0].BgpSession = nil
		expectError(target, "bgp_session")

		target = toGcp()
		target.Spec.Tunnels[0].Name = "Tunnel_0"
		expectError(target, "name")
	})

	ginkgo.It("should reject a session without a peer ASN, a bad interface range, or a non-link-local peer address", func() {
		target := toGcp()
		target.Spec.Tunnels[0].BgpSession.PeerAsn = 0
		expectError(target, "peer_asn")

		target = toGcp()
		target.Spec.Tunnels[0].BgpSession.InterfaceIpRange = "10.0.0.1/30"
		expectError(target, "interface_ip_range must be a link-local IPv4 CIDR")

		target = toGcp()
		target.Spec.Tunnels[0].BgpSession.InterfaceIpRange = "169.254.10.1/29"
		expectError(target, "interface_ip_range must be a link-local IPv4 CIDR")

		target = toGcp()
		target.Spec.Tunnels[0].BgpSession.PeerIpAddress = proto.String("10.0.0.2")
		expectError(target, "peer_ip_address, when set, must be a link-local address")
	})

	ginkgo.It("should reject custom advertisement in DEFAULT mode and out-of-range priorities, BFD, IKE, and interface ids", func() {
		target := toGcp()
		target.Spec.Tunnels[0].BgpSession.AdvertisedGroups = []string{"ALL_SUBNETS"}
		expectError(target, "advertised_groups can only be set when advertise_mode is CUSTOM")

		target = toGcp()
		target.Spec.Tunnels[0].BgpSession.AdvertisedRoutePriority = proto.Int32(70000)
		expectError(target, "advertised_route_priority")

		target = toGcp()
		target.Spec.Tunnels[0].BgpSession.CustomLearnedRoutePriority = proto.Int32(65336)
		expectError(target, "custom_learned_route_priority")

		target = toGcp()
		target.Spec.Tunnels[0].BgpSession.Bfd = &GcpHaVpnConnectionBgpBfd{SessionInitializationMode: "ACTIVE", Multiplier: 4}
		expectError(target, "multiplier must be between 5 and 16")

		target = toGcp()
		target.Spec.Tunnels[0].BgpSession.Bfd = &GcpHaVpnConnectionBgpBfd{SessionInitializationMode: "ON"}
		expectError(target, "session_initialization_mode must be ACTIVE, PASSIVE, or DISABLED")

		target = toGcp()
		target.Spec.Tunnels[0].BgpSession.Bfd = &GcpHaVpnConnectionBgpBfd{SessionInitializationMode: "ACTIVE", MinReceiveInterval: 500}
		expectError(target, "min_receive_interval must be between 1000 and 30000")

		target = toGcp()
		target.Spec.Tunnels[0].IkeVersion = proto.Int32(3)
		expectError(target, "ike_version")

		target = toGcp()
		target.Spec.Tunnels[0].VpnGatewayInterface = 2
		expectError(target, "vpn_gateway_interface")
	})

	ginkgo.It("should reject an MD5 key without material and malformed traffic selectors", func() {
		target := toGcp()
		target.Spec.Tunnels[0].BgpSession.Md5AuthenticationKey = &GcpHaVpnConnectionBgpMd5AuthenticationKey{Name: "k"}
		expectError(target, "key")

		target = toGcp()
		target.Spec.Tunnels[0].LocalTrafficSelector = []string{"10.0.0.0"}
		expectError(target, "each traffic selector must be an IPv4 CIDR")
	})

	ginkgo.It("should reject unknown enum values", func() {
		target := toOnprem()
		target.Spec.Peer.ExternalGateway.RedundancyType = "THREE_IPS"
		expectError(target, "redundancy_type must be")

		target = toGcp()
		target.Spec.Tunnels[0].BgpSession.IpVersion = proto.String("DUAL")
		expectError(target, "ip_version must be IPV4 or IPV6")

		target = toGcp()
		target.Spec.DeletionPolicy = "RETAIN"
		expectError(target, "deletion_policy must be one of")
	})
})
