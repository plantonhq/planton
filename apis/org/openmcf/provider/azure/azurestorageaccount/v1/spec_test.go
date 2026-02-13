package azurestorageaccountv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAzureStorageAccountSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureStorageAccountSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AzureStorageAccountSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_storage_account", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-storage-account",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-resource-group"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for production configuration", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-storage",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzureStorageAccountSpec{
						Region:                 "eastus",
						ResourceGroup:          stringRef("prod-storage-rg"),
						AccountKind:            AzureStorageAccountKind_STORAGE_V2.Enum(),
						AccountTier:            AzureStorageAccountTier_STANDARD.Enum(),
						ReplicationType:        AzureStorageReplicationType_GRS.Enum(),
						AccessTier:             AzureStorageAccessTier_HOT.Enum(),
						EnableHttpsTrafficOnly: boolPtr(true),
						MinTlsVersion:          AzureTlsVersion_TLS1_2.Enum(),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with blob containers", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "storage-with-containers",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "westus2",
						ResourceGroup: stringRef("test-rg"),
						Containers: []*AzureStorageContainer{
							{
								Name:       "data",
								AccessType: AzureStorageContainerAccess_PRIVATE.Enum(),
							},
							{
								Name:       "public-assets",
								AccessType: AzureStorageContainerAccess_BLOB.Enum(),
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with network rules configured", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "storage-with-network-rules",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						NetworkRules: &AzureStorageNetworkRules{
							DefaultAction:       AzureStorageNetworkAction_DENY.Enum(),
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

			ginkgo.It("should not return a validation error with blob properties configured", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "storage-with-blob-props",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						BlobProperties: &AzureStorageBlobProperties{
							EnableVersioning:                 boolPtr(true),
							SoftDeleteRetentionDays:          int32Ptr(30),
							ContainerSoftDeleteRetentionDays: int32Ptr(14),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with premium tier and block blob storage", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "premium-storage",
					},
					Spec: &AzureStorageAccountSpec{
						Region:          "eastus",
						ResourceGroup:   stringRef("test-rg"),
						AccountKind:     AzureStorageAccountKind_BLOCK_BLOB_STORAGE.Enum(),
						AccountTier:     AzureStorageAccountTier_PREMIUM.Enum(),
						ReplicationType: AzureStorageReplicationType_LRS.Enum(),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with cool access tier", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "cool-storage",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						AccessTier:    AzureStorageAccessTier_COOL.Enum(),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with geo-zone-redundant storage", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "gzrs-storage",
					},
					Spec: &AzureStorageAccountSpec{
						Region:          "eastus",
						ResourceGroup:   stringRef("test-rg"),
						ReplicationType: AzureStorageReplicationType_GZRS.Enum(),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_storage_account", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-storage-account",
					},
					Spec: &AzureStorageAccountSpec{
						ResourceGroup: stringRef("test-resource-group"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty string", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-storage-account",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "",
						ResourceGroup: stringRef("test-resource-group"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-storage-account",
					},
					Spec: &AzureStorageAccountSpec{
						Region: "eastus",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is nil", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-storage-account",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: nil,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when container name is too short", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-storage",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						Containers: []*AzureStorageContainer{
							{
								Name: "ab", // Too short, needs at least 3 characters
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when container name is too long", func() {
				longName := "this-is-a-very-long-container-name-that-exceeds-the-sixty-three-character-limit"
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-storage",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						Containers: []*AzureStorageContainer{
							{
								Name: longName,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when soft_delete_retention_days exceeds maximum (> 365)", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "invalid-retention-storage",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						BlobProperties: &AzureStorageBlobProperties{
							SoftDeleteRetentionDays: int32Ptr(366),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when container_soft_delete_retention_days exceeds maximum (> 365)", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "invalid-container-retention-storage",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						BlobProperties: &AzureStorageBlobProperties{
							ContainerSoftDeleteRetentionDays: int32Ptr(366),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when containers exceeds maximum (> 100)", func() {
				// Create a list with 101 containers
				containers := make([]*AzureStorageContainer, 101)
				for i := 0; i < 101; i++ {
					containers[i] = &AzureStorageContainer{
						Name: "container" + string(rune('a'+i%26)),
					}
				}

				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "too-many-containers-storage",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						Containers:    containers,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network rules has too many IP rules (> 200)", func() {
				// Create a list with 201 IP rules
				ipRules := make([]string, 201)
				for i := 0; i < 201; i++ {
					ipRules[i] = "10.0.0.1/32"
				}

				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "too-many-ips-storage",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						NetworkRules: &AzureStorageNetworkRules{
							IpRules: ipRules,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when network rules has too many VNet subnet IDs (> 100)", func() {
				// Create a list with 101 subnet IDs
				subnetIds := make([]string, 101)
				for i := 0; i < 101; i++ {
					subnetIds[i] = "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet1"
				}

				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "too-many-subnets-storage",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						NetworkRules: &AzureStorageNetworkRules{
							VirtualNetworkSubnetIds: subnetIds,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &AzureStorageAccount{
					ApiVersion: "wrong.version/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-storage",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-storage",
					},
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Spec: &AzureStorageAccountSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureStorageAccount{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureStorageAccount",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-storage",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}

// Helper functions for pointer types
func boolPtr(b bool) *bool {
	return &b
}

func int32Ptr(i int32) *int32 {
	return &i
}
