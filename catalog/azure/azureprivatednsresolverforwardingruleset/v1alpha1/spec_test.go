package azureprivatednsresolverforwardingrulesetv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzurePrivateDnsResolverForwardingRulesetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePrivateDnsResolverForwardingRulesetSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

func boolPtr(v bool) *bool { return &v }

const testOutboundEndpointId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/dnsResolvers/platform-resolver/outboundEndpoints/outbound"

// validRule returns a well-formed forwarding rule that individual
// cases mutate into the shape under test.
func validRule() *AzurePrivateDnsResolverForwardingRule {
	return &AzurePrivateDnsResolverForwardingRule{
		Name:       "corp-domain",
		DomainName: "corp.contoso.com.",
		TargetDnsServers: []*AzurePrivateDnsResolverForwardingRuleTargetDnsServer{
			{IpAddress: "10.0.0.4"},
		},
	}
}

// validResource returns a minimal valid ruleset (one endpoint, no
// rules) that individual cases mutate into the shape under test.
func validResource() *AzurePrivateDnsResolverForwardingRuleset {
	return &AzurePrivateDnsResolverForwardingRuleset{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePrivateDnsResolverForwardingRuleset",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ruleset",
		},
		Spec: &AzurePrivateDnsResolverForwardingRulesetSpec{
			Region:              "eastus",
			ResourceGroup:       literal("platform-rg"),
			Name:                "platform-ruleset",
			OutboundEndpointIds: []*foreignkeyv1.StringValueOrRef{literal(testOutboundEndpointId)},
		},
	}
}

var _ = ginkgo.Describe("AzurePrivateDnsResolverForwardingRulesetSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_private_dns_resolver_forwarding_ruleset", func() {

			ginkgo.It("should not return a validation error for the minimal ruleset (no rules)", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a ruleset with one rule at defaults (port and enabled unset)", func() {
				input := validResource()
				input.Spec.ForwardingRules = []*AzurePrivateDnsResolverForwardingRule{validRule()}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the full rule shape: multiple targets with ports, disabled, metadata", func() {
				input := validResource()
				rule := validRule()
				rule.TargetDnsServers = []*AzurePrivateDnsResolverForwardingRuleTargetDnsServer{
					{IpAddress: "10.0.0.4", Port: int32Ptr(53)},
					{IpAddress: "10.0.0.5", Port: int32Ptr(5353)},
				}
				rule.Enabled = boolPtr(false)
				rule.Metadata = map[string]string{"owner": "network-team"}
				input.Spec.ForwardingRules = []*AzurePrivateDnsResolverForwardingRule{rule}
				input.Spec.Tags = map[string]string{"team": "platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple rules with unique names and two outbound endpoints", func() {
				input := validResource()
				input.Spec.OutboundEndpointIds = []*foreignkeyv1.StringValueOrRef{
					literal(testOutboundEndpointId),
					literal(testOutboundEndpointId + "-2"),
				}
				second := validRule()
				second.Name = "second-domain"
				second.DomainName = "internal.contoso.com."
				input.Spec.ForwardingRules = []*AzurePrivateDnsResolverForwardingRule{validRule(), second}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_private_dns_resolver_forwarding_ruleset", func() {

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing ruleset name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty outbound endpoint list (the provider requires at least one)", func() {
				input := validResource()
				input.Spec.OutboundEndpointIds = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate rule names", func() {
				input := validResource()
				input.Spec.ForwardingRules = []*AzurePrivateDnsResolverForwardingRule{validRule(), validRule()}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "unique")).To(gomega.BeTrue())
			})

			ginkgo.It("should reject a rule without a name", func() {
				input := validResource()
				rule := validRule()
				rule.Name = ""
				input.Spec.ForwardingRules = []*AzurePrivateDnsResolverForwardingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a rule without a domain name", func() {
				input := validResource()
				rule := validRule()
				rule.DomainName = ""
				input.Spec.ForwardingRules = []*AzurePrivateDnsResolverForwardingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a rule with no target DNS servers", func() {
				input := validResource()
				rule := validRule()
				rule.TargetDnsServers = nil
				input.Spec.ForwardingRules = []*AzurePrivateDnsResolverForwardingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a target that is not an IPv4 address", func() {
				input := validResource()
				rule := validRule()
				rule.TargetDnsServers = []*AzurePrivateDnsResolverForwardingRuleTargetDnsServer{
					{IpAddress: "dns.onprem.local"},
				}
				input.Spec.ForwardingRules = []*AzurePrivateDnsResolverForwardingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject port 0 (below the valid range)", func() {
				input := validResource()
				rule := validRule()
				rule.TargetDnsServers = []*AzurePrivateDnsResolverForwardingRuleTargetDnsServer{
					{IpAddress: "10.0.0.4", Port: int32Ptr(0)},
				}
				input.Spec.ForwardingRules = []*AzurePrivateDnsResolverForwardingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a port above 65535", func() {
				input := validResource()
				rule := validRule()
				rule.TargetDnsServers = []*AzurePrivateDnsResolverForwardingRuleTargetDnsServer{
					{IpAddress: "10.0.0.4", Port: int32Ptr(65536)},
				}
				input.Spec.ForwardingRules = []*AzurePrivateDnsResolverForwardingRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a wrong api version", func() {
				input := validResource()
				input.ApiVersion = "azure.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a wrong kind", func() {
				input := validResource()
				input.Kind = "AzureDnsForwardingRuleset"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
