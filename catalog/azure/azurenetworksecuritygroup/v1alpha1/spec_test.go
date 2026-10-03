package azurenetworksecuritygroupv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureNetworkSecurityGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureNetworkSecurityGroupSpec Validation Tests")
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

// strPtr returns a pointer to the given string (for optional fields).
func strPtr(value string) *string {
	return &value
}

// validRule returns a minimal valid security rule that individual cases
// then mutate into the shape under test.
func validRule() *AzureNetworkSecurityGroupRule {
	return &AzureNetworkSecurityGroupRule{
		Name:                 "allow-https-inbound",
		Priority:             100,
		Direction:            AzureNetworkSecurityGroupRuleDirection_INBOUND,
		Access:               AzureNetworkSecurityGroupRuleAccess_ALLOW,
		Protocol:             AzureNetworkSecurityGroupRuleProtocol_TCP,
		DestinationPortRange: strPtr("443"),
	}
}

// validResource returns a minimal valid AzureNetworkSecurityGroup that
// individual cases then mutate into the shape under test.
func validResource() *AzureNetworkSecurityGroup {
	return &AzureNetworkSecurityGroup{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureNetworkSecurityGroup",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-nsg",
		},
		Spec: &AzureNetworkSecurityGroupSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "web-tier",
		},
	}
}

var _ = ginkgo.Describe("AzureNetworkSecurityGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_network_security_group", func() {

			ginkgo.It("should not return a validation error for a rule-less group (Azure defaults govern)", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the resource group as a reference", func() {
				input := validResource()
				input.Spec.ResourceGroup = ref("platform-rg")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a minimal allow rule", func() {
				input := validResource()
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{validRule()}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every protocol including AH and ESP", func() {
				for _, protocol := range []AzureNetworkSecurityGroupRuleProtocol{
					AzureNetworkSecurityGroupRuleProtocol_ANY,
					AzureNetworkSecurityGroupRuleProtocol_TCP,
					AzureNetworkSecurityGroupRuleProtocol_UDP,
					AzureNetworkSecurityGroupRuleProtocol_ICMP,
					AzureNetworkSecurityGroupRuleProtocol_AH,
					AzureNetworkSecurityGroupRuleProtocol_ESP,
				} {
					input := validResource()
					rule := validRule()
					rule.Protocol = protocol
					input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept plural destination ports", func() {
				input := validResource()
				rule := validRule()
				rule.DestinationPortRange = nil
				rule.DestinationPortRanges = []string{"80", "443"}
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept plural source address prefixes", func() {
				input := validResource()
				rule := validRule()
				rule.SourceAddressPrefixes = []string{"10.1.0.0/16", "10.2.0.0/16"}
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept application security group addressing", func() {
				input := validResource()
				rule := validRule()
				rule.SourceApplicationSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/applicationSecurityGroups/web-servers"),
				}
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a service-tag source with a deny decision", func() {
				input := validResource()
				rule := validRule()
				rule.Name = "deny-internet-inbound"
				rule.Access = AzureNetworkSecurityGroupRuleAccess_DENY
				rule.SourceAddressPrefix = strPtr("Internet")
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept user tags", func() {
				input := validResource()
				input.Spec.Tags = map[string]string{"cost-center": "networking"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_network_security_group", func() {

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

			ginkgo.It("should return a validation error when name starts with a period", func() {
				input := validResource()
				input.Spec.Name = ".bad"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a rule's priority is below 100", func() {
				input := validResource()
				rule := validRule()
				rule.Priority = 99
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a rule's priority is above 4096", func() {
				input := validResource()
				rule := validRule()
				rule.Priority = 4097
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a rule's direction is unspecified", func() {
				input := validResource()
				rule := validRule()
				rule.Direction = AzureNetworkSecurityGroupRuleDirection_azure_network_security_group_rule_direction_unspecified
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a rule's access is unspecified", func() {
				input := validResource()
				rule := validRule()
				rule.Access = AzureNetworkSecurityGroupRuleAccess_azure_network_security_group_rule_access_unspecified
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a rule's protocol is unspecified", func() {
				input := validResource()
				rule := validRule()
				rule.Protocol = AzureNetworkSecurityGroupRuleProtocol_azure_network_security_group_rule_protocol_unspecified
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both destination port forms are set", func() {
				input := validResource()
				rule := validRule()
				rule.DestinationPortRanges = []string{"80"}
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when no destination port form is set", func() {
				input := validResource()
				rule := validRule()
				rule.DestinationPortRange = nil
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both source port forms are set", func() {
				input := validResource()
				rule := validRule()
				rule.SourcePortRange = strPtr("*")
				rule.SourcePortRanges = []string{"1024-65535"}
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when two source addressing styles are combined", func() {
				input := validResource()
				rule := validRule()
				rule.SourceAddressPrefix = strPtr("10.0.0.0/8")
				rule.SourceAddressPrefixes = []string{"10.1.0.0/16"}
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a prefix and an application security group are combined", func() {
				input := validResource()
				rule := validRule()
				rule.DestinationAddressPrefix = strPtr("10.0.1.0/24")
				rule.DestinationApplicationSecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
					literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/applicationSecurityGroups/db-servers"),
				}
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when more than 10 application security groups are listed", func() {
				input := validResource()
				rule := validRule()
				for i := 0; i < 11; i++ {
					rule.SourceApplicationSecurityGroupIds = append(rule.SourceApplicationSecurityGroupIds,
						literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/applicationSecurityGroups/asg"))
				}
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a rule description exceeds 140 characters", func() {
				input := validResource()
				rule := validRule()
				description := ""
				for len(description) < 141 {
					description += "d"
				}
				rule.Description = description
				input.Spec.SecurityRules = []*AzureNetworkSecurityGroupRule{rule}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := validResource()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := validResource()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := validResource()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := validResource()
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
