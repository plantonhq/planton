package azurefirewallpolicyrulecollectiongroupv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureFirewallPolicyRuleCollectionGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFirewallPolicyRuleCollectionGroupSpec Validation Tests")
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

// validNetworkRule returns a well-formed network rule for reuse.
func validNetworkRule() *AzureFirewallPolicyNetworkRule {
	return &AzureFirewallPolicyNetworkRule{
		Name:                 "allow-dns",
		Protocols:            []AzureFirewallPolicyRuleProtocol{AzureFirewallPolicyRuleProtocol_UDP},
		SourceAddresses:      []string{"10.0.0.0/16"},
		DestinationAddresses: []string{"8.8.8.8"},
		DestinationPorts:     []string{"53"},
	}
}

// validNatRule returns a well-formed DNAT rule for reuse.
func validNatRule() *AzureFirewallPolicyNatRule {
	return &AzureFirewallPolicyNatRule{
		Name:               "rdp-to-jumpbox",
		Protocols:          []AzureFirewallPolicyRuleProtocol{AzureFirewallPolicyRuleProtocol_TCP},
		SourceAddresses:    []string{"*"},
		DestinationAddress: "203.0.113.10",
		DestinationPorts:   []string{"3389"},
		TranslatedAddress:  "10.0.1.4",
		TranslatedPort:     3389,
	}
}

// validResource returns a minimal valid group that individual cases then
// mutate into the shape under test.
func validResource() *AzureFirewallPolicyRuleCollectionGroup {
	return &AzureFirewallPolicyRuleCollectionGroup{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFirewallPolicyRuleCollectionGroup",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-rcg",
		},
		Spec: &AzureFirewallPolicyRuleCollectionGroupSpec{
			FirewallPolicyId: ref("egress-baseline"),
			Name:             "platform-baseline",
			Priority:         500,
			NetworkRuleCollections: []*AzureFirewallPolicyNetworkRuleCollection{
				{
					Name:     "core-egress",
					Priority: 200,
					Action:   AzureFirewallPolicyFilterAction_ALLOW,
					Rules:    []*AzureFirewallPolicyNetworkRule{validNetworkRule()},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureFirewallPolicyRuleCollectionGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_firewall_policy_rule_collection_group", func() {

			ginkgo.It("should not return a validation error for a minimal network-rule group", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an empty group (no collections)", func() {
				input := validResource()
				input.Spec.NetworkRuleCollections = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a full application rule with TLS termination and headers", func() {
				input := validResource()
				input.Spec.ApplicationRuleCollections = []*AzureFirewallPolicyApplicationRuleCollection{
					{
						Name:     "web-egress",
						Priority: 300,
						Action:   AzureFirewallPolicyFilterAction_ALLOW,
						Rules: []*AzureFirewallPolicyApplicationRule{
							{
								Name: "allow-github",
								Protocols: []*AzureFirewallPolicyApplicationProtocol{
									{Type: AzureFirewallPolicyApplicationProtocolType_HTTPS, Port: 443},
								},
								SourceAddresses: []string{"10.0.0.0/16"},
								DestinationUrls: []string{"github.com/planton/*"},
								TerminateTls:    true,
								HttpHeaders:     []*AzureFirewallPolicyHttpHeader{{Name: "X-Org", Value: "planton"}},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a network rule with ip group references", func() {
				input := validResource()
				rule := validNetworkRule()
				rule.SourceAddresses = nil
				rule.SourceIpGroups = []*foreignkeyv1.StringValueOrRef{ref("branch-offices")}
				rule.DestinationAddresses = nil
				rule.DestinationIpGroups = []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/ipGroups/dc")}
				input.Spec.NetworkRuleCollections[0].Rules = []*AzureFirewallPolicyNetworkRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a DNAT rule with a translated address", func() {
				input := validResource()
				input.Spec.NatRuleCollections = []*AzureFirewallPolicyNatRuleCollection{
					{
						Name:     "inbound-dnat",
						Priority: 100,
						Rules:    []*AzureFirewallPolicyNatRule{validNatRule()},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a DNAT rule with a translated fqdn instead of an address", func() {
				input := validResource()
				rule := validNatRule()
				rule.TranslatedAddress = ""
				rule.TranslatedFqdn = "jumpbox.internal.example.com"
				input.Spec.NatRuleCollections = []*AzureFirewallPolicyNatRuleCollection{
					{Name: "inbound-dnat", Priority: 100, Rules: []*AzureFirewallPolicyNatRule{rule}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept boundary priorities 100 and 65000", func() {
				input := validResource()
				input.Spec.Priority = 65000
				input.Spec.NetworkRuleCollections[0].Priority = 100
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_firewall_policy_rule_collection_group", func() {

			ginkgo.It("should return a validation error when firewall_policy_id is missing", func() {
				input := validResource()
				input.Spec.FirewallPolicyId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when priority is below 100", func() {
				input := validResource()
				input.Spec.Priority = 99
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when priority is above 65000", func() {
				input := validResource()
				input.Spec.Priority = 65001
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a collection without rules", func() {
				input := validResource()
				input.Spec.NetworkRuleCollections[0].Rules = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a filter collection without an action", func() {
				input := validResource()
				input.Spec.NetworkRuleCollections[0].Action = AzureFirewallPolicyFilterAction_azure_firewall_policy_filter_action_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a network rule without protocols", func() {
				input := validResource()
				input.Spec.NetworkRuleCollections[0].Rules[0].Protocols = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a network rule without destination ports", func() {
				input := validResource()
				input.Spec.NetworkRuleCollections[0].Rules[0].DestinationPorts = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a DNAT rule with an ANY protocol", func() {
				input := validResource()
				rule := validNatRule()
				rule.Protocols = []AzureFirewallPolicyRuleProtocol{AzureFirewallPolicyRuleProtocol_ANY}
				input.Spec.NatRuleCollections = []*AzureFirewallPolicyNatRuleCollection{
					{Name: "inbound-dnat", Priority: 100, Rules: []*AzureFirewallPolicyNatRule{rule}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a DNAT rule with both translation targets", func() {
				input := validResource()
				rule := validNatRule()
				rule.TranslatedFqdn = "jumpbox.internal.example.com"
				input.Spec.NatRuleCollections = []*AzureFirewallPolicyNatRuleCollection{
					{Name: "inbound-dnat", Priority: 100, Rules: []*AzureFirewallPolicyNatRule{rule}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a DNAT rule with neither translation target", func() {
				input := validResource()
				rule := validNatRule()
				rule.TranslatedAddress = ""
				input.Spec.NatRuleCollections = []*AzureFirewallPolicyNatRuleCollection{
					{Name: "inbound-dnat", Priority: 100, Rules: []*AzureFirewallPolicyNatRule{rule}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a DNAT rule with two destination port entries", func() {
				input := validResource()
				rule := validNatRule()
				rule.DestinationPorts = []string{"3389", "3390"}
				input.Spec.NatRuleCollections = []*AzureFirewallPolicyNatRuleCollection{
					{Name: "inbound-dnat", Priority: 100, Rules: []*AzureFirewallPolicyNatRule{rule}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a DNAT rule with translated_port zero", func() {
				input := validResource()
				rule := validNatRule()
				rule.TranslatedPort = 0
				input.Spec.NatRuleCollections = []*AzureFirewallPolicyNatRuleCollection{
					{Name: "inbound-dnat", Priority: 100, Rules: []*AzureFirewallPolicyNatRule{rule}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an application protocol port above 64000", func() {
				input := validResource()
				input.Spec.ApplicationRuleCollections = []*AzureFirewallPolicyApplicationRuleCollection{
					{
						Name:     "web-egress",
						Priority: 300,
						Action:   AzureFirewallPolicyFilterAction_ALLOW,
						Rules: []*AzureFirewallPolicyApplicationRule{
							{
								Name: "bad-port",
								Protocols: []*AzureFirewallPolicyApplicationProtocol{
									{Type: AzureFirewallPolicyApplicationProtocolType_HTTP, Port: 64001},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
