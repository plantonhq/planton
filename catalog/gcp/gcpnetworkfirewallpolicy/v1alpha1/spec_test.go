package gcpnetworkfirewallpolicyv1alpha1

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
	ginkgo.RunSpecs(t, "GcpNetworkFirewallPolicySpec Suite")
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

var _ = ginkgo.Describe("GcpNetworkFirewallPolicySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	tcp := func(ports ...string) *GcpNetworkFirewallPolicyLayer4Config {
		return &GcpNetworkFirewallPolicyLayer4Config{IpProtocol: "tcp", Ports: ports}
	}

	rule := func(priority int32, action, direction string, l4 ...*GcpNetworkFirewallPolicyLayer4Config) *GcpNetworkFirewallPolicyRule {
		return &GcpNetworkFirewallPolicyRule{
			Priority:  proto.Int32(priority),
			Action:    action,
			Direction: direction,
			Match:     &GcpNetworkFirewallPolicyRuleMatch{Layer4Configs: l4},
		}
	}

	// A global baseline on one network: allow IAP SSH and health checks,
	// deny every other ingress.
	globalBaseline := func() *GcpNetworkFirewallPolicy {
		iap := rule(1000, "allow", "INGRESS", tcp("22"))
		iap.RuleName = "allow-iap-ssh"
		iap.Match.SrcIpRanges = []string{"35.235.240.0/20"}
		hc := rule(1100, "allow", "INGRESS", tcp())
		hc.Match.SrcIpRanges = []string{"35.191.0.0/16", "130.211.0.0/22"}
		deny := rule(65000, "deny", "INGRESS", &GcpNetworkFirewallPolicyLayer4Config{IpProtocol: "all"})
		deny.Match.SrcIpRanges = []string{"0.0.0.0/0"}
		deny.EnableLogging = true
		return &GcpNetworkFirewallPolicy{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpNetworkFirewallPolicy",
			Metadata: &shared.CloudResourceMetadata{
				Name: "baseline",
			},
			Spec: &GcpNetworkFirewallPolicySpec{
				ProjectId: nameRef("acme-net"),
				Rules:     []*GcpNetworkFirewallPolicyRule{iap, hc, deny},
				Associations: []*GcpNetworkFirewallPolicyAssociation{{
					Name:    "baseline-main",
					Network: nameRef("main-vpc"),
				}},
			},
		}
	}

	expectError := func(target *GcpNetworkFirewallPolicy, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a global baseline policy associated with a network by reference", func() {
		gomega.Expect(validator.Validate(globalBaseline())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a regional policy with an internal managed load balancer rule", func() {
		target := globalBaseline()
		target.Spec.Region = "us-central1"
		ilb := rule(2000, "allow", "INGRESS", tcp("80", "443"))
		ilb.TargetType = proto.String("INTERNAL_MANAGED_LB")
		ilb.TargetForwardingRules = []*foreignkeyv1.StringValueOrRef{litRef("projects/acme-net/regions/us-central1/forwardingRules/ilb")}
		ilb.Match.SrcIpRanges = []string{"10.0.0.0/8"}
		target.Spec.Rules = append(target.Spec.Rules, ilb)
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every policy type on a regional policy and VPC_POLICY on a global one", func() {
		target := globalBaseline()
		target.Spec.PolicyType = proto.String("VPC_POLICY")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		target.Spec.Region = "europe-west4"
		for _, pt := range []string{"VPC_POLICY", "RDMA_ROCE_POLICY", "RDMA_FALCON_POLICY", "ULL_POLICY"} {
			target.Spec.PolicyType = proto.String(pt)
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept every match selector populated, with secure tags and networks by reference", func() {
		target := globalBaseline()
		r := target.Spec.Rules[0]
		r.Match = &GcpNetworkFirewallPolicyRuleMatch{
			Layer4Configs:           []*GcpNetworkFirewallPolicyLayer4Config{tcp("443", "8000-8999"), {IpProtocol: "17"}, {IpProtocol: "icmp"}},
			SrcIpRanges:             []string{"10.0.0.0/8", "2001:db8::/32"},
			DestIpRanges:            []string{"10.10.0.0/16"},
			SrcAddressGroups:        []string{"projects/acme/locations/global/addressGroups/partners"},
			DestAddressGroups:       []string{"organizations/123456789012/locations/global/addressGroups/dbs"},
			SrcFqdns:                []string{"partner.example.com"},
			DestFqdns:               []string{"api.example.com"},
			SrcRegionCodes:          []string{"US", "DE"},
			DestRegionCodes:         []string{"IN"},
			SrcThreatIntelligences:  []string{"iplist-known-malicious-ips"},
			DestThreatIntelligences: []string{"iplist-tor-exit-nodes"},
			SrcSecureTags:           []*foreignkeyv1.StringValueOrRef{nameRef("frontend-tag"), litRef("tagValues/281474976710656")},
			SrcNetworks:             []*foreignkeyv1.StringValueOrRef{nameRef("shared-vpc")},
			SrcNetworkContext:       proto.String("VPC_NETWORKS"),
			DestNetworkContext:      proto.String("INTRA_VPC"),
		}
		r.TargetSecureTags = []*foreignkeyv1.StringValueOrRef{nameRef("backend-tag")}
		r.TargetServiceAccounts = []*foreignkeyv1.StringValueOrRef{nameRef("backend-sa")}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a security profile group rule with TLS inspection", func() {
		target := globalBaseline()
		r := rule(500, "apply_security_profile_group", "EGRESS", tcp("443"))
		r.SecurityProfileGroup = "//networksecurity.googleapis.com/projects/acme-net/locations/global/securityProfileGroups/ngfw"
		r.TlsInspect = true
		target.Spec.Rules = append(target.Spec.Rules, r)
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept priority 0 and the lowest priority 2147483647", func() {
		target := globalBaseline()
		target.Spec.Rules[0].Priority = proto.Int32(0)
		target.Spec.Rules[2].Priority = proto.Int32(2147483647)
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a policy with no rules, no associations, no project, and every deletion policy", func() {
		target := globalBaseline()
		target.Spec.ProjectId = nil
		target.Spec.Rules = nil
		target.Spec.Associations = nil
		for _, dp := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target.Spec.DeletionPolicy = dp
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a bad policy name and a bad region", func() {
		target := globalBaseline()
		target.Spec.PolicyName = "Bad_Name-"
		expectError(target, "policy_name must be 1-63 characters")
		target = globalBaseline()
		target.Spec.Region = "US-Central1"
		expectError(target, "region must be a valid GCP region name")
	})

	ginkgo.It("should reject an RDMA or ULL policy type on a global policy and an unknown type", func() {
		target := globalBaseline()
		target.Spec.PolicyType = proto.String("RDMA_ROCE_POLICY")
		expectError(target, "exist only for a regional policy")
		target = globalBaseline()
		target.Spec.PolicyType = proto.String("FIREWALL_POLICY")
		expectError(target, "policy_type must be one of")
	})

	ginkgo.It("should reject two rules with the same priority and a rule without a priority", func() {
		target := globalBaseline()
		target.Spec.Rules[1].Priority = proto.Int32(1000)
		expectError(target, "distinct priority")
		target = globalBaseline()
		target.Spec.Rules[0].Priority = nil
		expectError(target, "priority")
	})

	ginkgo.It("should reject an unknown action and direction", func() {
		target := globalBaseline()
		target.Spec.Rules[0].Action = "ALLOW"
		expectError(target, "action must be one of")
		target = globalBaseline()
		target.Spec.Rules[0].Direction = "inbound"
		expectError(target, "direction must be INGRESS or EGRESS")
	})

	ginkgo.It("should reject a match without layer4 configs and ports on a non-tcp/udp protocol", func() {
		target := globalBaseline()
		target.Spec.Rules[0].Match.Layer4Configs = nil
		expectError(target, "layer4_configs")
		target = globalBaseline()
		target.Spec.Rules[0].Match.Layer4Configs = []*GcpNetworkFirewallPolicyLayer4Config{{IpProtocol: "all", Ports: []string{"22"}}}
		expectError(target, "ports can be set only when")
	})

	ginkgo.It("should reject an unknown protocol, an out-of-range protocol number, and a malformed port", func() {
		target := globalBaseline()
		target.Spec.Rules[0].Match.Layer4Configs = []*GcpNetworkFirewallPolicyLayer4Config{{IpProtocol: "TCP"}}
		expectError(target, "ip_protocol must be")
		target = globalBaseline()
		target.Spec.Rules[0].Match.Layer4Configs = []*GcpNetworkFirewallPolicyLayer4Config{{IpProtocol: "256"}}
		expectError(target, "ip_protocol must be")
		target = globalBaseline()
		target.Spec.Rules[0].Match.Layer4Configs = []*GcpNetworkFirewallPolicyLayer4Config{tcp("http")}
		expectError(target, "ports")
	})

	ginkgo.It("should reject a bare address without a prefix length and a lowercase region code", func() {
		target := globalBaseline()
		target.Spec.Rules[0].Match.DestIpRanges = []string{"10.0.0.1"}
		expectError(target, "dest_ip_ranges")
		target = globalBaseline()
		target.Spec.Rules[0].Match.DestRegionCodes = []string{"in"}
		expectError(target, "dest_region_codes")
	})

	ginkgo.It("should reject more than a hundred FQDNs and an unknown network context", func() {
		target := globalBaseline()
		fqdns := make([]string, 101)
		for i := range fqdns {
			fqdns[i] = "a.example.com"
		}
		target.Spec.Rules[0].Match.DestFqdns = fqdns
		expectError(target, "dest_fqdns")
		target = globalBaseline()
		target.Spec.Rules[0].Match.DestNetworkContext = proto.String("EXTERNAL")
		expectError(target, "dest_network_context must be one of")
	})

	ginkgo.It("should reject a security profile group on an allow rule and its absence on an apply rule", func() {
		target := globalBaseline()
		target.Spec.Rules[0].SecurityProfileGroup = "//networksecurity.googleapis.com/projects/acme/locations/global/securityProfileGroups/ngfw"
		expectError(target, "security_profile_group is required when")
		target = globalBaseline()
		target.Spec.Rules[0].Action = "apply_security_profile_group"
		expectError(target, "security_profile_group is required when")
	})

	ginkgo.It("should reject tls_inspect on an allow rule and logging on a goto_next rule", func() {
		target := globalBaseline()
		target.Spec.Rules[0].TlsInspect = true
		expectError(target, "tls_inspect can be true only")
		target = globalBaseline()
		target.Spec.Rules[0].Action = "goto_next"
		target.Spec.Rules[0].EnableLogging = true
		expectError(target, "enable_logging cannot be set on a goto_next")
	})

	ginkgo.It("should reject forwarding rules without the load balancer target type, and the type without rules", func() {
		target := globalBaseline()
		target.Spec.Rules[0].TargetForwardingRules = []*foreignkeyv1.StringValueOrRef{nameRef("ilb")}
		expectError(target, "target_forwarding_rules is required when")
		target = globalBaseline()
		target.Spec.Rules[0].TargetType = proto.String("INTERNAL_MANAGED_LB")
		expectError(target, "target_forwarding_rules is required when")
		target = globalBaseline()
		target.Spec.Rules[0].TargetType = proto.String("LOAD_BALANCER")
		expectError(target, "target_type must be INSTANCES or INTERNAL_MANAGED_LB")
	})

	ginkgo.It("should reject an association without a network, a duplicate name, or a bad name", func() {
		target := globalBaseline()
		target.Spec.Associations[0].Network = nil
		expectError(target, "network")
		target = globalBaseline()
		target.Spec.Associations = append(target.Spec.Associations, &GcpNetworkFirewallPolicyAssociation{
			Name:    "baseline-main",
			Network: nameRef("other-vpc"),
		})
		expectError(target, "distinct name within the policy")
		target = globalBaseline()
		target.Spec.Associations[0].Name = "-bad"
		expectError(target, "name must be 1-63 characters")
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		target := globalBaseline()
		target.Spec.DeletionPolicy = "RETAIN"
		expectError(target, "deletion_policy must be one of")
	})
})
