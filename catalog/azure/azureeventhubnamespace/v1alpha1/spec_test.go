package azureeventhubnamespacev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureEventHubNamespaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventHubNamespaceSpec Validation Tests")
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

// helper to create a minimal valid namespace (STANDARD tier by default)
func minimalNamespace() *AzureEventHubNamespace {
	return &AzureEventHubNamespace{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventHubNamespace",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-eh",
		},
		Spec: &AzureEventHubNamespaceSpec{
			Region:        "eastus",
			ResourceGroup: literal("my-rg"),
			NamespaceName: "myapp-eventhubs",
		},
	}
}

// helper to create a valid rule set (ALLOW; no admitted sources needed)
func allowRuleSet() *AzureEventHubNamespaceNetworkRuleSets {
	return &AzureEventHubNamespaceNetworkRuleSets{
		DefaultAction: AzureEventHubNetworkRuleSetDefaultAction_ALLOW,
	}
}

var _ = ginkgo.Describe("AzureEventHubNamespaceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_event_hub_namespace", func() {

			ginkgo.It("should accept a minimal STANDARD namespace", func() {
				err := protovalidate.Validate(minimalNamespace())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every declared sku", func() {
				for _, sku := range []AzureEventHubNamespaceSku{
					AzureEventHubNamespaceSku_BASIC,
					AzureEventHubNamespaceSku_STANDARD,
					AzureEventHubNamespaceSku_PREMIUM,
				} {
					input := minimalNamespace()
					input.Spec.Sku = sku
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept throughput capacity up to 40 on STANDARD", func() {
				capacity := int32(40)
				input := minimalNamespace()
				input.Spec.Sku = AzureEventHubNamespaceSku_STANDARD
				input.Spec.Capacity = &capacity
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept every allowed PREMIUM processing-unit count", func() {
				for _, c := range []int32{1, 2, 4, 8, 16} {
					capacity := c
					input := minimalNamespace()
					input.Spec.Sku = AzureEventHubNamespaceSku_PREMIUM
					input.Spec.Capacity = &capacity
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept auto-inflate with a throughput ceiling", func() {
				autoInflate := true
				ceiling := int32(20)
				input := minimalNamespace()
				input.Spec.AutoInflateEnabled = &autoInflate
				input.Spec.MaximumThroughputUnits = &ceiling
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a dedicated cluster reference", func() {
				input := minimalNamespace()
				input.Spec.DedicatedClusterId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.EventHub/clusters/my-cluster")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := minimalNamespace()
				input.Spec.Identity = &AzureEventHubNamespaceIdentity{
					Type: AzureEventHubNamespaceIdentityType_SYSTEM_ASSIGNED,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with identity ids", func() {
				input := minimalNamespace()
				input.Spec.Identity = &AzureEventHubNamespaceIdentity{
					Type: AzureEventHubNamespaceIdentityType_USER_ASSIGNED,
					UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
						literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept the keyless posture dial", func() {
				localAuth := false
				input := minimalNamespace()
				input.Spec.LocalAuthenticationEnabled = &localAuth
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an ALLOW rule set with no admitted sources", func() {
				input := minimalNamespace()
				input.Spec.NetworkRuleSets = allowRuleSet()
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a DENY rule set with admitted ip rules", func() {
				input := minimalNamespace()
				input.Spec.NetworkRuleSets = &AzureEventHubNamespaceNetworkRuleSets{
					DefaultAction: AzureEventHubNetworkRuleSetDefaultAction_DENY,
					IpRules:       []string{"203.0.113.0/24"},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a DENY rule set with admitted subnets", func() {
				input := minimalNamespace()
				input.Spec.NetworkRuleSets = &AzureEventHubNamespaceNetworkRuleSets{
					DefaultAction: AzureEventHubNetworkRuleSetDefaultAction_DENY,
					VirtualNetworkRules: []*AzureEventHubNamespaceVirtualNetworkRule{
						{SubnetId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/app")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a private posture with matching rule-set dial", func() {
				publicAccess := false
				rulesetAccess := false
				input := minimalNamespace()
				input.Spec.PublicNetworkAccessEnabled = &publicAccess
				input.Spec.NetworkRuleSets = allowRuleSet()
				input.Spec.NetworkRuleSets.PublicNetworkAccessEnabled = &rulesetAccess
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept user tags", func() {
				input := minimalNamespace()
				input.Spec.Tags = map[string]string{"cost-center": "streaming", "team": "data-platform"}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("namespace_name", func() {

			ginkgo.It("should reject a missing namespace_name", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name shorter than 6 characters", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = "abc"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name starting with a number", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = "1myeventhubs"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name ending with a hyphen", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = "myeventhubs-"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with invalid characters", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = "my_event_hubs"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("required scalars", func() {

			ginkgo.It("should reject a missing region", func() {
				input := minimalNamespace()
				input.Spec.Region = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing resource_group", func() {
				input := minimalNamespace()
				input.Spec.ResourceGroup = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("capacity", func() {

			ginkgo.It("should reject zero capacity", func() {
				capacity := int32(0)
				input := minimalNamespace()
				input.Spec.Capacity = &capacity
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a PREMIUM capacity that is not a sold PU count", func() {
				capacity := int32(3)
				input := minimalNamespace()
				input.Spec.Sku = AzureEventHubNamespaceSku_PREMIUM
				input.Spec.Capacity = &capacity
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a throughput capacity above 40 on STANDARD", func() {
				capacity := int32(41)
				input := minimalNamespace()
				input.Spec.Sku = AzureEventHubNamespaceSku_STANDARD
				input.Spec.Capacity = &capacity
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("maximum_throughput_units", func() {

			ginkgo.It("should reject a ceiling above 40", func() {
				ceiling := int32(41)
				input := minimalNamespace()
				input.Spec.MaximumThroughputUnits = &ceiling
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a negative ceiling", func() {
				ceiling := int32(-1)
				input := minimalNamespace()
				input.Spec.MaximumThroughputUnits = &ceiling
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("identity", func() {

			ginkgo.It("should reject an unspecified identity type", func() {
				input := minimalNamespace()
				input.Spec.Identity = &AzureEventHubNamespaceIdentity{}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject USER_ASSIGNED without identity ids", func() {
				input := minimalNamespace()
				input.Spec.Identity = &AzureEventHubNamespaceIdentity{
					Type: AzureEventHubNamespaceIdentityType_USER_ASSIGNED,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject SYSTEM_ASSIGNED with identity ids", func() {
				input := minimalNamespace()
				input.Spec.Identity = &AzureEventHubNamespaceIdentity{
					Type: AzureEventHubNamespaceIdentityType_SYSTEM_ASSIGNED,
					UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
						literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("network_rule_sets", func() {

			ginkgo.It("should reject a rule set on the BASIC tier", func() {
				input := minimalNamespace()
				input.Spec.Sku = AzureEventHubNamespaceSku_BASIC
				input.Spec.NetworkRuleSets = allowRuleSet()
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unspecified default_action", func() {
				input := minimalNamespace()
				input.Spec.NetworkRuleSets = &AzureEventHubNamespaceNetworkRuleSets{}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject DENY with no admitted sources", func() {
				input := minimalNamespace()
				input.Spec.NetworkRuleSets = &AzureEventHubNamespaceNetworkRuleSets{
					DefaultAction: AzureEventHubNetworkRuleSetDefaultAction_DENY,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a rule-set public-access dial that disagrees with the namespace", func() {
				rulesetAccess := false
				input := minimalNamespace()
				input.Spec.NetworkRuleSets = allowRuleSet()
				input.Spec.NetworkRuleSets.PublicNetworkAccessEnabled = &rulesetAccess
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing subnet_id on a virtual network rule", func() {
				input := minimalNamespace()
				input.Spec.NetworkRuleSets = &AzureEventHubNamespaceNetworkRuleSets{
					DefaultAction:       AzureEventHubNetworkRuleSetDefaultAction_DENY,
					VirtualNetworkRules: []*AzureEventHubNamespaceVirtualNetworkRule{{}},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
