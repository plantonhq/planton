package gcprouternatv1alpha1

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
	ginkgo.RunSpecs(t, "GcpRouterNatSpec Suite")
}

func value(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpRouterNatSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpRouterNat.
	minimal := func() *GcpRouterNat {
		return &GcpRouterNat{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpRouterNat",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-router-nat",
			},
			Spec: &GcpRouterNatSpec{
				RouterName:  "test-router",
				NatName:     "test-nat",
				Region:      "us-central1",
				VpcSelfLink: value("projects/test-project-123/global/networks/test-vpc"),
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec (auto-allocate, all subnets)", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept omitted project_id (ambient provider project)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an explicit project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = value("test-project-123")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a full HTTPS VPC self link", func() {
		msg := minimal()
		msg.Spec.VpcSelfLink = value("https://www.googleapis.com/compute/v1/projects/p/global/networks/n")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a VPC reference (valueFrom)", func() {
		msg := minimal()
		msg.Spec.VpcSelfLink = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
				ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-vpc"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept manual NAT IPs with a drain entry", func() {
		msg := minimal()
		msg.Spec.NatIps = []*foreignkeyv1.StringValueOrRef{
			value("projects/p/regions/us-central1/addresses/ip-a"),
			value("projects/p/regions/us-central1/addresses/ip-b"),
		}
		msg.Spec.DrainNatIps = []*foreignkeyv1.StringValueOrRef{
			value("projects/p/regions/us-central1/addresses/ip-b"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept subnetwork scoping with secondary range selection", func() {
		msg := minimal()
		msg.Spec.SourceSubnetworkIpRangesToNat = "LIST_OF_SUBNETWORKS"
		msg.Spec.Subnetworks = []*GcpRouterNatSubnetwork{
			{
				Subnetwork:            value("projects/p/regions/us-central1/subnetworks/s"),
				SourceIpRangesToNat:   []string{"PRIMARY_IP_RANGE", "LIST_OF_SECONDARY_IP_RANGES"},
				SecondaryIpRangeNames: []string{"pods"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept listed subnetworks with the mode left empty (implied)", func() {
		msg := minimal()
		msg.Spec.Subnetworks = []*GcpRouterNatSubnetwork{
			{Subnetwork: value("projects/p/regions/us-central1/subnetworks/s")},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES", func() {
		msg := minimal()
		msg.Spec.SourceSubnetworkIpRangesToNat = "ALL_SUBNETWORKS_ALL_PRIMARY_IP_RANGES"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept dynamic port allocation with power-of-two bounds", func() {
		msg := minimal()
		msg.Spec.EnableDynamicPortAllocation = true
		msg.Spec.MinPortsPerVm = 64
		msg.Spec.MaxPortsPerVm = 4096
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a non-power-of-two min_ports_per_vm with static allocation", func() {
		msg := minimal()
		msg.Spec.MinPortsPerVm = 100
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept endpoint-independent mapping without dynamic ports", func() {
		msg := minimal()
		msg.Spec.EnableEndpointIndependentMapping = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept all timeouts and auto_network_tier", func() {
		msg := minimal()
		msg.Spec.UdpIdleTimeoutSec = 60
		msg.Spec.IcmpIdleTimeoutSec = 45
		msg.Spec.TcpEstablishedIdleTimeoutSec = 600
		msg.Spec.TcpTransitoryIdleTimeoutSec = 15
		msg.Spec.TcpTimeWaitTimeoutSec = 60
		msg.Spec.AutoNetworkTier = "STANDARD"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each valid endpoint type", func() {
		for _, endpointType := range []string{"ENDPOINT_TYPE_VM", "ENDPOINT_TYPE_SWG", "ENDPOINT_TYPE_MANAGED_PROXY_LB"} {
			msg := minimal()
			msg.Spec.EndpointTypes = []string{endpointType}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should accept a public NAT rule with dedicated active IPs", func() {
		msg := minimal()
		msg.Spec.NatIps = []*foreignkeyv1.StringValueOrRef{
			value("projects/p/regions/us-central1/addresses/pool-ip"),
		}
		msg.Spec.Rules = []*GcpRouterNatRule{
			{
				RuleNumber:  100,
				Match:       "destination.ip == '203.0.113.10'",
				Description: "partner egress",
				Action: &GcpRouterNatRuleAction{
					SourceNatActiveIps: []*foreignkeyv1.StringValueOrRef{
						value("projects/p/regions/us-central1/addresses/partner-ip"),
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a private NAT with subnetwork-range rules", func() {
		msg := minimal()
		msg.Spec.Type = "PRIVATE"
		msg.Spec.Rules = []*GcpRouterNatRule{
			{
				RuleNumber: 200,
				Match:      "nexthop.hub == '//networkconnectivity.googleapis.com/projects/p/locations/global/hubs/h'",
				Action: &GcpRouterNatRuleAction{
					SourceNatActiveRanges: []*foreignkeyv1.StringValueOrRef{
						value("projects/p/regions/us-central1/subnetworks/nat-range"),
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a router BGP arm with private ASN and keepalive", func() {
		msg := minimal()
		msg.Spec.RouterBgp = &GcpRouterNatRouterBgp{
			Asn:               64514,
			KeepaliveInterval: 30,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a 32-bit private ASN", func() {
		msg := minimal()
		msg.Spec.RouterBgp = &GcpRouterNatRouterBgp{Asn: 4200000100}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept CUSTOM advertisement with groups and ranges", func() {
		msg := minimal()
		msg.Spec.RouterBgp = &GcpRouterNatRouterBgp{
			Asn:              64514,
			AdvertiseMode:    "CUSTOM",
			AdvertisedGroups: []string{"ALL_SUBNETS"},
			AdvertisedIpRanges: []*GcpRouterNatRouterBgpAdvertisedIpRange{
				{Range: "10.10.0.0/16", Description: "on-prem reachable services"},
			},
			IdentifierRange: "169.254.8.0/30",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept DEFAULT advertisement with no custom entries", func() {
		msg := minimal()
		msg.Spec.RouterBgp = &GcpRouterNatRouterBgp{Asn: 64514, AdvertiseMode: "DEFAULT"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a router description and encrypted-interconnect dedication", func() {
		msg := minimal()
		msg.Spec.RouterDescription = "egress router for the prod fleet"
		msg.Spec.EncryptedInterconnectRouter = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept resource manager tags on the router", func() {
		msg := minimal()
		msg.Spec.ResourceManagerTags = map[string]string{"tagKeys/123": "tagValues/456"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept every deletion_policy value", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = policy
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should accept NAT64 for all IPv6 subnetworks", func() {
		msg := minimal()
		msg.Spec.SourceSubnetworkIpRangesToNat64 = "ALL_IPV6_SUBNETWORKS"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept NAT64 scoped to listed subnetworks", func() {
		msg := minimal()
		msg.Spec.SourceSubnetworkIpRangesToNat64 = "LIST_OF_IPV6_SUBNETWORKS"
		msg.Spec.Nat64Subnetworks = []*foreignkeyv1.StringValueOrRef{
			value("projects/p/regions/us-central1/subnetworks/dual-stack"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept listed nat64_subnetworks with the mode left empty (implied)", func() {
		msg := minimal()
		msg.Spec.Nat64Subnetworks = []*foreignkeyv1.StringValueOrRef{
			value("projects/p/regions/us-central1/subnetworks/dual-stack"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept every log filter value", func() {
		for _, filter := range []GcpRouterNatLogFilter{
			GcpRouterNatLogFilter_DISABLED,
			GcpRouterNatLogFilter_ERRORS_ONLY,
			GcpRouterNatLogFilter_ALL,
			GcpRouterNatLogFilter_TRANSLATIONS_ONLY,
		} {
			msg := minimal()
			logFilter := filter
			msg.Spec.LogFilter = &logFilter
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject when metadata is missing", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when spec is missing", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when router_name is empty", func() {
		msg := minimal()
		msg.Spec.RouterName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid router_name format", func() {
		msg := minimal()
		msg.Spec.RouterName = "Invalid_Router"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when nat_name is empty", func() {
		msg := minimal()
		msg.Spec.NatName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid nat_name format", func() {
		msg := minimal()
		msg.Spec.NatName = "nat-ending-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when region is empty", func() {
		msg := minimal()
		msg.Spec.Region = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when vpc_self_link is missing", func() {
		msg := minimal()
		msg.Spec.VpcSelfLink = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid type", func() {
		msg := minimal()
		msg.Spec.Type = "HYBRID"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid source_subnetwork_ip_ranges_to_nat", func() {
		msg := minimal()
		msg.Spec.SourceSubnetworkIpRangesToNat = "SOME_SUBNETWORKS"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject LIST_OF_SUBNETWORKS with no subnetworks listed", func() {
		msg := minimal()
		msg.Spec.SourceSubnetworkIpRangesToNat = "LIST_OF_SUBNETWORKS"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject listed subnetworks with an ALL_SUBNETWORKS mode", func() {
		msg := minimal()
		msg.Spec.SourceSubnetworkIpRangesToNat = "ALL_SUBNETWORKS_ALL_IP_RANGES"
		msg.Spec.Subnetworks = []*GcpRouterNatSubnetwork{
			{Subnetwork: value("projects/p/regions/us-central1/subnetworks/s")},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a subnetwork entry without the subnetwork ref", func() {
		msg := minimal()
		msg.Spec.Subnetworks = []*GcpRouterNatSubnetwork{
			{SourceIpRangesToNat: []string{"ALL_IP_RANGES"}},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid source_ip_ranges_to_nat value", func() {
		msg := minimal()
		msg.Spec.Subnetworks = []*GcpRouterNatSubnetwork{
			{
				Subnetwork:          value("projects/p/regions/us-central1/subnetworks/s"),
				SourceIpRangesToNat: []string{"EVERYTHING"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject LIST_OF_SECONDARY_IP_RANGES without secondary names", func() {
		msg := minimal()
		msg.Spec.Subnetworks = []*GcpRouterNatSubnetwork{
			{
				Subnetwork:          value("projects/p/regions/us-central1/subnetworks/s"),
				SourceIpRangesToNat: []string{"LIST_OF_SECONDARY_IP_RANGES"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject secondary names without the secondary-range mode", func() {
		msg := minimal()
		msg.Spec.Subnetworks = []*GcpRouterNatSubnetwork{
			{
				Subnetwork:            value("projects/p/regions/us-central1/subnetworks/s"),
				SourceIpRangesToNat:   []string{"PRIMARY_IP_RANGE"},
				SecondaryIpRangeNames: []string{"pods"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject dynamic port allocation combined with EIM", func() {
		msg := minimal()
		msg.Spec.EnableDynamicPortAllocation = true
		msg.Spec.EnableEndpointIndependentMapping = true
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject max_ports_per_vm without dynamic allocation", func() {
		msg := minimal()
		msg.Spec.MaxPortsPerVm = 4096
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject max_ports_per_vm below min_ports_per_vm", func() {
		msg := minimal()
		msg.Spec.EnableDynamicPortAllocation = true
		msg.Spec.MinPortsPerVm = 4096
		msg.Spec.MaxPortsPerVm = 64
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject non-power-of-two ports with dynamic allocation", func() {
		msg := minimal()
		msg.Spec.EnableDynamicPortAllocation = true
		msg.Spec.MinPortsPerVm = 100
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject more than one endpoint type", func() {
		msg := minimal()
		msg.Spec.EndpointTypes = []string{"ENDPOINT_TYPE_VM", "ENDPOINT_TYPE_SWG"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid endpoint type", func() {
		msg := minimal()
		msg.Spec.EndpointTypes = []string{"ENDPOINT_TYPE_GKE"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid auto_network_tier", func() {
		msg := minimal()
		msg.Spec.AutoNetworkTier = "BASIC"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject private NAT with manual nat_ips", func() {
		msg := minimal()
		msg.Spec.Type = "PRIVATE"
		msg.Spec.NatIps = []*foreignkeyv1.StringValueOrRef{
			value("projects/p/regions/us-central1/addresses/ip-a"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject private NAT with auto_network_tier", func() {
		msg := minimal()
		msg.Spec.Type = "PRIVATE"
		msg.Spec.AutoNetworkTier = "PREMIUM"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject drain_nat_ips without manual nat_ips", func() {
		msg := minimal()
		msg.Spec.DrainNatIps = []*foreignkeyv1.StringValueOrRef{
			value("projects/p/regions/us-central1/addresses/ip-a"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a rule without a match expression", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpRouterNatRule{
			{RuleNumber: 100},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a rule_number above 65000", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpRouterNatRule{
			{RuleNumber: 65001, Match: "destination.ip == '203.0.113.10'"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject IP-based rule actions on private NAT", func() {
		msg := minimal()
		msg.Spec.Type = "PRIVATE"
		msg.Spec.Rules = []*GcpRouterNatRule{
			{
				RuleNumber: 100,
				Match:      "destination.ip == '203.0.113.10'",
				Action: &GcpRouterNatRuleAction{
					SourceNatActiveIps: []*foreignkeyv1.StringValueOrRef{
						value("projects/p/regions/us-central1/addresses/ip-a"),
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject range-based rule actions on public NAT", func() {
		msg := minimal()
		msg.Spec.Rules = []*GcpRouterNatRule{
			{
				RuleNumber: 100,
				Match:      "destination.ip == '203.0.113.10'",
				Action: &GcpRouterNatRuleAction{
					SourceNatActiveRanges: []*foreignkeyv1.StringValueOrRef{
						value("projects/p/regions/us-central1/subnetworks/s"),
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a non-private router ASN", func() {
		msg := minimal()
		msg.Spec.RouterBgp = &GcpRouterNatRouterBgp{Asn: 15169}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unset ASN when router_bgp is present", func() {
		msg := minimal()
		msg.Spec.RouterBgp = &GcpRouterNatRouterBgp{KeepaliveInterval: 30}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a router keepalive outside 20-60", func() {
		msg := minimal()
		msg.Spec.RouterBgp = &GcpRouterNatRouterBgp{Asn: 64514, KeepaliveInterval: 10}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid advertise_mode", func() {
		msg := minimal()
		msg.Spec.RouterBgp = &GcpRouterNatRouterBgp{Asn: 64514, AdvertiseMode: "SELECTIVE"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject advertised_groups without CUSTOM mode", func() {
		msg := minimal()
		msg.Spec.RouterBgp = &GcpRouterNatRouterBgp{
			Asn:              64514,
			AdvertisedGroups: []string{"ALL_SUBNETS"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject advertised_ip_ranges without CUSTOM mode", func() {
		msg := minimal()
		msg.Spec.RouterBgp = &GcpRouterNatRouterBgp{
			Asn:           64514,
			AdvertiseMode: "DEFAULT",
			AdvertisedIpRanges: []*GcpRouterNatRouterBgpAdvertisedIpRange{
				{Range: "10.10.0.0/16"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an advertised range with no CIDR", func() {
		msg := minimal()
		msg.Spec.RouterBgp = &GcpRouterNatRouterBgp{
			Asn:                64514,
			AdvertiseMode:      "CUSTOM",
			AdvertisedIpRanges: []*GcpRouterNatRouterBgpAdvertisedIpRange{{Description: "empty"}},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an unknown advertised group value", func() {
		msg := minimal()
		msg.Spec.RouterBgp = &GcpRouterNatRouterBgp{
			Asn:              64514,
			AdvertiseMode:    "CUSTOM",
			AdvertisedGroups: []string{"ALL_VPC_SUBNETS"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid NAT64 mode", func() {
		msg := minimal()
		msg.Spec.SourceSubnetworkIpRangesToNat64 = "SOME_IPV6_SUBNETWORKS"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject NAT64 list mode with no subnetworks listed", func() {
		msg := minimal()
		msg.Spec.SourceSubnetworkIpRangesToNat64 = "LIST_OF_IPV6_SUBNETWORKS"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject listed nat64_subnetworks with an ALL_IPV6 mode", func() {
		msg := minimal()
		msg.Spec.SourceSubnetworkIpRangesToNat64 = "ALL_IPV6_SUBNETWORKS"
		msg.Spec.Nat64Subnetworks = []*foreignkeyv1.StringValueOrRef{
			value("projects/p/regions/us-central1/subnetworks/dual-stack"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
