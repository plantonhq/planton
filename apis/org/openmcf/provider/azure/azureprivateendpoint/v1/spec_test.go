package azureprivateendpointv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAzurePrivateEndpointSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePrivateEndpointSpec Validation Tests")
}

var _ = ginkgo.Describe("AzurePrivateEndpointSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_private_endpoint", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "pg-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "pg-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-pg",
							},
						},
						SubresourceNames: []string{"postgresqlServer"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full metadata", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-pg-pe",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "westeurope",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-network-rg",
							},
						},
						Name: "prod-pg-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/prod-pg",
							},
						},
						SubresourceNames: []string{"postgresqlServer"},
						PrivateDnsZoneId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Network/privateDnsZones/privatelink.postgres.database.azure.com",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for key vault private endpoint", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "kv-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "kv-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.KeyVault/vaults/my-kv",
							},
						},
						SubresourceNames: []string{"vault"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for storage blob private endpoint", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "blob-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "blob-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Storage/storageAccounts/mystorage",
							},
						},
						SubresourceNames: []string{"blob"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error without subresource names", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "custom-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "custom-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/privateLinkServices/my-pls",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error without private DNS zone", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "pg-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "pg-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-pg",
							},
						},
						SubresourceNames: []string{"postgresqlServer"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom references", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "pg-pe",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "prod-rg",
								},
							},
						},
						Name: "pg-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "pe-subnet",
								},
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "prod-postgresql",
								},
							},
						},
						SubresourceNames: []string{"postgresqlServer"},
						PrivateDnsZoneId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "pg-dns-zone",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with max length name", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "max-name-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij12",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-pg",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_private_endpoint", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-pg",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						Name:   "test-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-pg",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-pg",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 80 characters", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-too-long",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-pg",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subnet_id is missing", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-private-endpoint",
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-pg",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when private_connection_resource_id is missing", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "wrong.version/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-pg",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pe",
					},
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-pg",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Spec: &AzurePrivateEndpointSpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "test-private-endpoint",
						SubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pe-subnet",
							},
						},
						PrivateConnectionResourceId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.DBforPostgreSQL/flexibleServers/my-pg",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzurePrivateEndpoint{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzurePrivateEndpoint",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pe",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
