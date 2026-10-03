package azureservicebusnamespacev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureServiceBusNamespaceSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureServiceBusNamespaceSpec Validation Tests")
}

// helper to create a minimal valid namespace (STANDARD tier by default)
func minimalNamespace() *AzureServiceBusNamespace {
	return &AzureServiceBusNamespace{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureServiceBusNamespace",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-sb",
		},
		Spec: &AzureServiceBusNamespaceSpec{
			Region: "eastus",
			ResourceGroup: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "my-rg",
				},
			},
			NamespaceName: "myapp-servicebus",
		},
	}
}

// helper to create a valid PREMIUM namespace (capacity + partitions required)
func premiumNamespace() *AzureServiceBusNamespace {
	capacity := int32(1)
	partitions := int32(1)
	input := minimalNamespace()
	input.Spec.Sku = AzureServiceBusNamespaceSku_PREMIUM
	input.Spec.Capacity = &capacity
	input.Spec.PremiumMessagingPartitions = &partitions
	return input
}

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("AzureServiceBusNamespaceSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_service_bus_namespace", func() {

			ginkgo.It("should accept a minimal STANDARD namespace", func() {
				err := protovalidate.Validate(minimalNamespace())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every declared sku with its required pairings", func() {
				basic := minimalNamespace()
				basic.Spec.Sku = AzureServiceBusNamespaceSku_BASIC
				gomega.Expect(protovalidate.Validate(basic)).To(gomega.BeNil())

				standard := minimalNamespace()
				standard.Spec.Sku = AzureServiceBusNamespaceSku_STANDARD
				gomega.Expect(protovalidate.Validate(standard)).To(gomega.BeNil())

				gomega.Expect(protovalidate.Validate(premiumNamespace())).To(gomega.BeNil())
			})

			ginkgo.It("should accept every allowed PREMIUM capacity", func() {
				for _, c := range []int32{1, 2, 4, 8, 16} {
					capacity := c
					input := premiumNamespace()
					input.Spec.Capacity = &capacity
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept every allowed partition count", func() {
				for _, p := range []int32{1, 2, 4} {
					partitions := p
					input := premiumNamespace()
					input.Spec.PremiumMessagingPartitions = &partitions
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := minimalNamespace()
				input.Spec.Identity = &AzureServiceBusNamespaceIdentity{
					Type: AzureServiceBusNamespaceIdentityType_SYSTEM_ASSIGNED,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with identity ids", func() {
				input := minimalNamespace()
				input.Spec.Identity = &AzureServiceBusNamespaceIdentity{
					Type: AzureServiceBusNamespaceIdentityType_USER_ASSIGNED,
					UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
						literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept CMK on PREMIUM with a user-assigned identity", func() {
				input := premiumNamespace()
				input.Spec.Identity = &AzureServiceBusNamespaceIdentity{
					Type: AzureServiceBusNamespaceIdentityType_USER_ASSIGNED,
					UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
						literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
					},
				}
				infra := true
				input.Spec.CustomerManagedKey = &AzureServiceBusNamespaceCustomerManagedKey{
					KeyVaultKeyId:                   literal("https://vault.vault.azure.net/keys/cmk"),
					UserAssignedIdentityId:          literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
					InfrastructureEncryptionEnabled: &infra,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a PREMIUM network rule set with DENY and admitted sources", func() {
				input := premiumNamespace()
				trusted := true
				input.Spec.NetworkRuleSet = &AzureServiceBusNamespaceNetworkRuleSet{
					DefaultAction:          AzureServiceBusNetworkDefaultAction_DENY,
					TrustedServicesAllowed: &trusted,
					IpRules:                []string{"203.0.113.0/24"},
					NetworkRules: []*AzureServiceBusNamespaceNetworkRule{
						{SubnetId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/app")},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an ALLOW network rule set without admitted sources", func() {
				input := premiumNamespace()
				input.Spec.NetworkRuleSet = &AzureServiceBusNamespaceNetworkRuleSet{
					DefaultAction: AzureServiceBusNetworkDefaultAction_ALLOW,
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a keyless posture (local auth disabled)", func() {
				localAuth := false
				input := minimalNamespace()
				input.Spec.LocalAuthEnabled = &localAuth
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept user tags", func() {
				input := minimalNamespace()
				input.Spec.Tags = map[string]string{"team": "payments", "cost-center": "cc-42"}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a valueFrom reference for resource_group", func() {
				input := minimalNamespace()
				input.Spec.ResourceGroup = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureResourceGroup,
							Name:      "shared-rg",
							FieldPath: "status.outputs.resource_group_name",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a namespace name at min length (6)", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = "abcde1"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a namespace name at max length (50)", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = "a-very-long-service-bus-namespace-name-for-test-01"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_service_bus_namespace", func() {

			ginkgo.It("should reject a missing region", func() {
				input := minimalNamespace()
				input.Spec.Region = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing resource_group", func() {
				input := minimalNamespace()
				input.Spec.ResourceGroup = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing namespace_name", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a namespace name shorter than 6 characters", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = "abcde"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a namespace name starting with a number", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = "1-invalid-name"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a namespace name ending with a hyphen", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = "invalid-name-"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject the reserved '-sb' suffix", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = "myapp-bus-sb"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject the reserved '-mgmt' suffix", func() {
				input := minimalNamespace()
				input.Spec.NamespaceName = "myapp-bus-mgmt"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an undeclared sku value", func() {
				input := minimalNamespace()
				input.Spec.Sku = AzureServiceBusNamespaceSku(99)
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject capacity on a non-PREMIUM namespace", func() {
				capacity := int32(2)
				input := minimalNamespace()
				input.Spec.Capacity = &capacity
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject PREMIUM without capacity", func() {
				input := premiumNamespace()
				input.Spec.Capacity = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a capacity outside the allowed set", func() {
				for _, c := range []int32{0, 3, 17} {
					capacity := c
					input := premiumNamespace()
					input.Spec.Capacity = &capacity
					gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
				}
			})

			ginkgo.It("should reject partitions on a non-PREMIUM namespace", func() {
				partitions := int32(1)
				input := minimalNamespace()
				input.Spec.PremiumMessagingPartitions = &partitions
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject PREMIUM without partitions", func() {
				input := premiumNamespace()
				input.Spec.PremiumMessagingPartitions = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a partition count outside the allowed set", func() {
				for _, p := range []int32{0, 3, 5} {
					partitions := p
					input := premiumNamespace()
					input.Spec.PremiumMessagingPartitions = &partitions
					gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
				}
			})

			ginkgo.It("should reject an identity block without a type", func() {
				input := minimalNamespace()
				input.Spec.Identity = &AzureServiceBusNamespaceIdentity{}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject USER_ASSIGNED identity without identity ids", func() {
				input := minimalNamespace()
				input.Spec.Identity = &AzureServiceBusNamespaceIdentity{
					Type: AzureServiceBusNamespaceIdentityType_USER_ASSIGNED,
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject SYSTEM_ASSIGNED identity carrying identity ids", func() {
				input := minimalNamespace()
				input.Spec.Identity = &AzureServiceBusNamespaceIdentity{
					Type: AzureServiceBusNamespaceIdentityType_SYSTEM_ASSIGNED,
					UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
						literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject CMK on a non-PREMIUM namespace", func() {
				input := minimalNamespace()
				input.Spec.Identity = &AzureServiceBusNamespaceIdentity{
					Type: AzureServiceBusNamespaceIdentityType_USER_ASSIGNED,
					UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
						literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
					},
				}
				input.Spec.CustomerManagedKey = &AzureServiceBusNamespaceCustomerManagedKey{
					KeyVaultKeyId:          literal("https://vault.vault.azure.net/keys/cmk"),
					UserAssignedIdentityId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject CMK without a user-assigned identity model", func() {
				input := premiumNamespace()
				input.Spec.Identity = &AzureServiceBusNamespaceIdentity{
					Type: AzureServiceBusNamespaceIdentityType_SYSTEM_ASSIGNED,
				}
				input.Spec.CustomerManagedKey = &AzureServiceBusNamespaceCustomerManagedKey{
					KeyVaultKeyId:          literal("https://vault.vault.azure.net/keys/cmk"),
					UserAssignedIdentityId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject CMK without the identity block at all", func() {
				input := premiumNamespace()
				input.Spec.CustomerManagedKey = &AzureServiceBusNamespaceCustomerManagedKey{
					KeyVaultKeyId:          literal("https://vault.vault.azure.net/keys/cmk"),
					UserAssignedIdentityId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject CMK missing its key reference", func() {
				input := premiumNamespace()
				input.Spec.Identity = &AzureServiceBusNamespaceIdentity{
					Type: AzureServiceBusNamespaceIdentityType_USER_ASSIGNED,
					UserAssignedIdentityIds: []*foreignkeyv1.StringValueOrRef{
						literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
					},
				}
				input.Spec.CustomerManagedKey = &AzureServiceBusNamespaceCustomerManagedKey{
					UserAssignedIdentityId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai"),
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject network_rule_set on a non-PREMIUM namespace", func() {
				input := minimalNamespace()
				input.Spec.NetworkRuleSet = &AzureServiceBusNamespaceNetworkRuleSet{
					DefaultAction: AzureServiceBusNetworkDefaultAction_ALLOW,
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject DENY with no admitted sources", func() {
				input := premiumNamespace()
				input.Spec.NetworkRuleSet = &AzureServiceBusNamespaceNetworkRuleSet{
					DefaultAction: AzureServiceBusNetworkDefaultAction_DENY,
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a network rule without its subnet", func() {
				input := premiumNamespace()
				input.Spec.NetworkRuleSet = &AzureServiceBusNamespaceNetworkRuleSet{
					DefaultAction: AzureServiceBusNetworkDefaultAction_DENY,
					NetworkRules:  []*AzureServiceBusNamespaceNetworkRule{{}},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing metadata block", func() {
				input := minimalNamespace()
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing spec block", func() {
				input := &AzureServiceBusNamespace{
					ApiVersion: "azure.planton.dev/v1alpha1",
					Kind:       "AzureServiceBusNamespace",
					Metadata: &shared.CatalogObjectMetadata{
						Name: "test-sb",
					},
				}
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an incorrect api_version", func() {
				input := minimalNamespace()
				input.ApiVersion = "wrong.version/v1"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an incorrect kind", func() {
				input := minimalNamespace()
				input.Kind = "WrongKind"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
