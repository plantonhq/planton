package azureprivatednszonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzurePrivateDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePrivateDnsZoneSpec Validation Tests")
}

var _ = ginkgo.Describe("AzurePrivateDnsZoneSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_private_dns_zone", func() {

			ginkgo.It("should not return a validation error for minimal valid fields (privatelink zone)", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "pg-private-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "privatelink.postgres.database.azure.com",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full metadata", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-pg-dns",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-network-rg",
							},
						},
						Name: "privatelink.postgres.database.azure.com",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for mysql privatelink zone", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "mysql-private-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "privatelink.mysql.database.azure.com",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for custom internal DNS zone", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "internal-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "contoso.internal",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with registration_enabled set to true", func() {
				regEnabled := true
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "internal-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "contoso.internal",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
						RegistrationEnabled: &regEnabled,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with registration_enabled set to false", func() {
				regEnabled := false
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "pg-private-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "privatelink.postgres.database.azure.com",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
						RegistrationEnabled: &regEnabled,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for keyvault privatelink zone", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "kv-private-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "privatelink.vaultcore.azure.net",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with valueFrom reference for resource_group", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "pg-private-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "my-rg-resource",
								},
							},
						},
						Name: "privatelink.postgres.database.azure.com",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
								ValueFrom: &foreignkeyv1.ValueFromRef{
									Name: "my-vpc-resource",
								},
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
		ginkgo.Context("azure_private_dns_zone", func() {

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						Name: "privatelink.postgres.database.azure.com",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vnet_id is missing", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "privatelink.postgres.database.azure.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when zone name has invalid format (uppercase)", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "PrivateLink.Postgres.Database.Azure.COM",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when zone name has invalid format (leading dot)", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: ".invalid.domain.com",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "wrong.version/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "privatelink.postgres.database.azure.com",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns",
					},
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "privatelink.postgres.database.azure.com",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Spec: &AzurePrivateDnsZoneSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name: "privatelink.postgres.database.azure.com",
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzurePrivateDnsZone{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzurePrivateDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
