package gcphierarchicalfirewallpolicyv1alpha1

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
	ginkgo.RunSpecs(t, "GcpHierarchicalFirewallPolicySpec Suite")
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

var _ = ginkgo.Describe("GcpHierarchicalFirewallPolicySpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	tcp := func(ports ...string) *GcpHierarchicalFirewallPolicyLayer4Config {
		return &GcpHierarchicalFirewallPolicyLayer4Config{IpProtocol: "tcp", Ports: ports}
	}

	rule := func(priority int32, action, direction string, l4 ...*GcpHierarchicalFirewallPolicyLayer4Config) *GcpHierarchicalFirewallPolicyRule {
		return &GcpHierarchicalFirewallPolicyRule{
			Priority:  proto.Int32(priority),
			Action:    action,
			Direction: direction,
			Match:     &GcpHierarchicalFirewallPolicyRuleMatch{Layer4Configs: l4},
		}
	}

	// An organization-wide baseline: deny SSH from the internet, delegate
	// the rest, enforced on the organization itself.
	orgBaseline := func() *GcpHierarchicalFirewallPolicy {
		deny := rule(1000, "deny", "INGRESS", tcp("22"))
		deny.Match.SrcIpRanges = []string{"0.0.0.0/0"}
		deny.EnableLogging = true
		next := rule(2000, "goto_next", "INGRESS", &GcpHierarchicalFirewallPolicyLayer4Config{IpProtocol: "all"})
		return &GcpHierarchicalFirewallPolicy{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpHierarchicalFirewallPolicy",
			Metadata: &shared.CloudResourceMetadata{
				Name: "org-baseline",
			},
			Spec: &GcpHierarchicalFirewallPolicySpec{
				Parent: &GcpHierarchicalFirewallPolicyParent{OrganizationId: "123456789012"},
				Rules:  []*GcpHierarchicalFirewallPolicyRule{deny, next},
				Associations: []*GcpHierarchicalFirewallPolicyAssociation{{
					Name:   "org-baseline-org",
					Target: &GcpHierarchicalFirewallPolicyAttachmentTarget{OrganizationId: "123456789012"},
				}},
			},
		}
	}

	expectError := func(target *GcpHierarchicalFirewallPolicy, substring string) {
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring(substring))
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept an organization baseline with a logged deny and a goto_next", func() {
		gomega.Expect(validator.Validate(orgBaseline())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a folder-scoped policy whose parent and association target are folder references", func() {
		target := orgBaseline()
		target.Spec.Parent = &GcpHierarchicalFirewallPolicyParent{FolderId: nameRef("prod-folder")}
		target.Spec.Associations[0].Target = &GcpHierarchicalFirewallPolicyAttachmentTarget{FolderId: nameRef("prod-folder")}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every match selector populated, with secure tags and networks by reference", func() {
		target := orgBaseline()
		r := target.Spec.Rules[0]
		r.Match = &GcpHierarchicalFirewallPolicyRuleMatch{
			Layer4Configs:           []*GcpHierarchicalFirewallPolicyLayer4Config{tcp("443", "8000-8999"), {IpProtocol: "17"}, {IpProtocol: "icmp"}},
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
		r.TargetResources = []*foreignkeyv1.StringValueOrRef{nameRef("shared-vpc")}
		r.TargetSecureTags = []*foreignkeyv1.StringValueOrRef{nameRef("backend-tag")}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a security profile group rule with TLS inspection", func() {
		target := orgBaseline()
		r := rule(500, "apply_security_profile_group", "EGRESS", tcp("443"))
		r.SecurityProfileGroup = "//networksecurity.googleapis.com/organizations/123456789012/locations/global/securityProfileGroups/ngfw"
		r.TlsInspect = true
		target.Spec.Rules = append(target.Spec.Rules, r)
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept priority 0 as the highest priority", func() {
		target := orgBaseline()
		target.Spec.Rules[0].Priority = proto.Int32(0)
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a policy with no rules and no associations, and every deletion policy", func() {
		target := orgBaseline()
		target.Spec.Rules = nil
		target.Spec.Associations = nil
		for _, dp := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			target.Spec.DeletionPolicy = dp
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept target_service_accounts by reference when no secure tags are set", func() {
		target := orgBaseline()
		target.Spec.Rules[0].TargetServiceAccounts = []*foreignkeyv1.StringValueOrRef{nameRef("web-sa"), litRef("web@acme.iam.gserviceaccount.com")}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a parent with both arms or neither", func() {
		target := orgBaseline()
		target.Spec.Parent.FolderId = litRef("987654321098")
		expectError(target, "set exactly one of organization_id or folder_id")
		target.Spec.Parent = &GcpHierarchicalFirewallPolicyParent{}
		expectError(target, "set exactly one of organization_id or folder_id")
	})

	ginkgo.It("should reject a non-numeric organization id and a bad short name", func() {
		target := orgBaseline()
		target.Spec.Parent.OrganizationId = "organizations/123"
		expectError(target, "numeric organization ID")
		target = orgBaseline()
		target.Spec.ShortName = "Bad_Name-"
		expectError(target, "short_name must be 1-63 characters")
	})

	ginkgo.It("should reject two rules with the same priority", func() {
		target := orgBaseline()
		target.Spec.Rules[1].Priority = proto.Int32(1000)
		expectError(target, "distinct priority")
	})

	ginkgo.It("should reject a rule without a priority or with a reserved implied priority", func() {
		target := orgBaseline()
		target.Spec.Rules[0].Priority = nil
		expectError(target, "priority")
		target = orgBaseline()
		target.Spec.Rules[0].Priority = proto.Int32(2147483646)
		expectError(target, "priority")
	})

	ginkgo.It("should reject an unknown action, an uppercase action, and an unknown direction", func() {
		target := orgBaseline()
		target.Spec.Rules[0].Action = "ALLOW"
		expectError(target, "action must be one of")
		target = orgBaseline()
		target.Spec.Rules[0].Direction = "inbound"
		expectError(target, "direction must be INGRESS or EGRESS")
	})

	ginkgo.It("should reject a match without layer4 configs and ports on a non-tcp/udp protocol", func() {
		target := orgBaseline()
		target.Spec.Rules[0].Match.Layer4Configs = nil
		expectError(target, "layer4_configs")
		target = orgBaseline()
		target.Spec.Rules[0].Match.Layer4Configs = []*GcpHierarchicalFirewallPolicyLayer4Config{{IpProtocol: "icmp", Ports: []string{"22"}}}
		expectError(target, "ports can be set only when")
	})

	ginkgo.It("should reject an unknown protocol, an out-of-range protocol number, and a malformed port", func() {
		target := orgBaseline()
		target.Spec.Rules[0].Match.Layer4Configs = []*GcpHierarchicalFirewallPolicyLayer4Config{{IpProtocol: "TCP"}}
		expectError(target, "ip_protocol must be")
		target = orgBaseline()
		target.Spec.Rules[0].Match.Layer4Configs = []*GcpHierarchicalFirewallPolicyLayer4Config{{IpProtocol: "256"}}
		expectError(target, "ip_protocol must be")
		target = orgBaseline()
		target.Spec.Rules[0].Match.Layer4Configs = []*GcpHierarchicalFirewallPolicyLayer4Config{tcp("http")}
		expectError(target, "ports")
	})

	ginkgo.It("should reject a bare address without a prefix length and a lowercase region code", func() {
		target := orgBaseline()
		target.Spec.Rules[0].Match.SrcIpRanges = []string{"10.0.0.1"}
		expectError(target, "src_ip_ranges")
		target = orgBaseline()
		target.Spec.Rules[0].Match.SrcRegionCodes = []string{"us"}
		expectError(target, "src_region_codes")
	})

	ginkgo.It("should reject more than ten address groups and an unknown network context", func() {
		target := orgBaseline()
		groups := make([]string, 11)
		for i := range groups {
			groups[i] = "projects/acme/locations/global/addressGroups/g"
		}
		target.Spec.Rules[0].Match.SrcAddressGroups = groups
		expectError(target, "src_address_groups")
		target = orgBaseline()
		target.Spec.Rules[0].Match.SrcNetworkContext = proto.String("EXTERNAL")
		expectError(target, "src_network_context must be one of")
	})

	ginkgo.It("should reject a security profile group on an allow rule and its absence on an apply rule", func() {
		target := orgBaseline()
		target.Spec.Rules[0].SecurityProfileGroup = "//networksecurity.googleapis.com/projects/acme/locations/global/securityProfileGroups/ngfw"
		expectError(target, "security_profile_group is required when")
		target = orgBaseline()
		target.Spec.Rules[0].Action = "apply_security_profile_group"
		expectError(target, "security_profile_group is required when")
	})

	ginkgo.It("should reject tls_inspect on a deny rule and logging on a goto_next rule", func() {
		target := orgBaseline()
		target.Spec.Rules[0].TlsInspect = true
		expectError(target, "tls_inspect can be true only")
		target = orgBaseline()
		target.Spec.Rules[1].EnableLogging = true
		expectError(target, "enable_logging cannot be set on a goto_next")
	})

	ginkgo.It("should reject target secure tags together with target service accounts", func() {
		target := orgBaseline()
		target.Spec.Rules[0].TargetSecureTags = []*foreignkeyv1.StringValueOrRef{nameRef("web-tag")}
		target.Spec.Rules[0].TargetServiceAccounts = []*foreignkeyv1.StringValueOrRef{nameRef("web-sa")}
		expectError(target, "cannot both be set on one rule")
	})

	ginkgo.It("should reject an association with both target arms, a duplicate name, or a bad name", func() {
		target := orgBaseline()
		target.Spec.Associations[0].Target.FolderId = litRef("987654321098")
		expectError(target, "attaches the policy to the organization or to one folder")
		target = orgBaseline()
		target.Spec.Associations = append(target.Spec.Associations, &GcpHierarchicalFirewallPolicyAssociation{
			Name:   "org-baseline-org",
			Target: &GcpHierarchicalFirewallPolicyAttachmentTarget{FolderId: litRef("987654321098")},
		})
		expectError(target, "distinct name within the policy")
		target = orgBaseline()
		target.Spec.Associations[0].Name = "-bad"
		expectError(target, "name must be 1-63 characters")
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		target := orgBaseline()
		target.Spec.DeletionPolicy = "RETAIN"
		expectError(target, "deletion_policy must be one of")
	})
})
