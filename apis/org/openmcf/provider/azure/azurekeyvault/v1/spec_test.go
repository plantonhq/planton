package azurekeyvaultv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAzureKeyVaultSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureKeyVaultSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AzureKeyVaultSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_key_vault", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-key-vault",
					},
					Spec: &AzureKeyVaultSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-resource-group"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for production configuration with premium SKU", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-key-vault",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzureKeyVaultSpec{
						Region:                  "eastus",
						ResourceGroup:           stringRef("prod-security-rg"),
						Sku:                     AzureKeyVaultSku_PREMIUM.Enum(),
						EnableRbacAuthorization: boolPtr(true),
						EnablePurgeProtection:   boolPtr(true),
						SoftDeleteRetentionDays: int32Ptr(90),
						SecretNames:             []string{"database-password", "api-key", "jwt-secret"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with network ACLs configured", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-kv-with-network",
					},
					Spec: &AzureKeyVaultSpec{
						Region:        "westus2",
						ResourceGroup: stringRef("test-rg"),
						NetworkAcls: &AzureKeyVaultNetworkAcls{
							DefaultAction:       AzureKeyVaultNetworkAction_DENY.Enum(),
							BypassAzureServices: boolPtr(true),
							IpRules:             []string{"203.0.113.0/24", "198.51.100.42/32"},
							VirtualNetworkSubnetIds: []string{
								"/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet1",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with minimum soft delete retention days", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "dev-key-vault",
					},
					Spec: &AzureKeyVaultSpec{
						Region:                  "eastus",
						ResourceGroup:           stringRef("dev-rg"),
						SoftDeleteRetentionDays: int32Ptr(7),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with maximum soft delete retention days", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-key-vault",
					},
					Spec: &AzureKeyVaultSpec{
						Region:                  "eastus",
						ResourceGroup:           stringRef("prod-rg"),
						SoftDeleteRetentionDays: int32Ptr(90),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with standard SKU", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "standard-kv",
					},
					Spec: &AzureKeyVaultSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						Sku:           AzureKeyVaultSku_STANDARD.Enum(),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with empty secret names list", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "empty-secrets-kv",
					},
					Spec: &AzureKeyVaultSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						SecretNames:   []string{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with network ACLs allowing all traffic", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "open-kv",
					},
					Spec: &AzureKeyVaultSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						NetworkAcls: &AzureKeyVaultNetworkAcls{
							DefaultAction:       AzureKeyVaultNetworkAction_ALLOW.Enum(),
							BypassAzureServices: boolPtr(false),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_key_vault", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-key-vault",
					},
					Spec: &AzureKeyVaultSpec{
						ResourceGroup: stringRef("test-resource-group"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty string", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-key-vault",
					},
					Spec: &AzureKeyVaultSpec{
						Region:        "",
						ResourceGroup: stringRef("test-resource-group"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-key-vault",
					},
					Spec: &AzureKeyVaultSpec{
						Region: "eastus",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is nil", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-key-vault",
					},
					Spec: &AzureKeyVaultSpec{
						Region:        "eastus",
						ResourceGroup: nil,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when soft_delete_retention_days is below minimum (< 7)", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "invalid-retention-kv",
					},
					Spec: &AzureKeyVaultSpec{
						Region:                  "eastus",
						ResourceGroup:           stringRef("test-rg"),
						SoftDeleteRetentionDays: int32Ptr(6),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when soft_delete_retention_days is above maximum (> 90)", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "invalid-retention-kv",
					},
					Spec: &AzureKeyVaultSpec{
						Region:                  "eastus",
						ResourceGroup:           stringRef("test-rg"),
						SoftDeleteRetentionDays: int32Ptr(91),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when secret_names exceeds maximum (> 100)", func() {
				// Create a list with 101 secret names
				secretNames := make([]string, 101)
				for i := 0; i < 101; i++ {
					secretNames[i] = "secret-" + string(rune('a'+i%26))
				}

				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "too-many-secrets-kv",
					},
					Spec: &AzureKeyVaultSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						SecretNames:   secretNames,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network ACLs has too many IP rules (> 200)", func() {
				// Create a list with 201 IP rules
				ipRules := make([]string, 201)
				for i := 0; i < 201; i++ {
					ipRules[i] = "10.0.0.1/32"
				}

				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "too-many-ips-kv",
					},
					Spec: &AzureKeyVaultSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						NetworkAcls: &AzureKeyVaultNetworkAcls{
							IpRules: ipRules,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network ACLs has too many VNet subnet IDs (> 100)", func() {
				// Create a list with 101 subnet IDs
				subnetIds := make([]string, 101)
				for i := 0; i < 101; i++ {
					subnetIds[i] = "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet1"
				}

				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "too-many-subnets-kv",
					},
					Spec: &AzureKeyVaultSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						NetworkAcls: &AzureKeyVaultNetworkAcls{
							VirtualNetworkSubnetIds: subnetIds,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &AzureKeyVault{
					ApiVersion: "wrong.version/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-key-vault",
					},
					Spec: &AzureKeyVaultSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-key-vault",
					},
					Spec: &AzureKeyVaultSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Spec: &AzureKeyVaultSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureKeyVault{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureKeyVault",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-key-vault",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

// Helper functions for pointer types
func boolPtr(b bool) *bool {
	return &b
}

func int32Ptr(i int32) *int32 {
	return &i
}

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}
