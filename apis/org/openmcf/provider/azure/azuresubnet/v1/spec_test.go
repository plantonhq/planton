package azuresubnetv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestAzureSubnetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureSubnetSpec Validation Tests")
}

var _ = ginkgo.Describe("AzureSubnetSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_subnet", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:          "my-subnet",
						AddressPrefix: "10.0.1.0/24",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full metadata", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-db-subnet",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzureSubnetSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-network-rg",
							},
						},
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet",
							},
						},
						Name:          "prod-db-subnet",
						AddressPrefix: "10.0.2.0/24",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with service endpoints", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:             "my-subnet",
						AddressPrefix:    "10.0.1.0/24",
						ServiceEndpoints: []string{"Microsoft.Sql", "Microsoft.Storage", "Microsoft.KeyVault"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with delegation", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:          "pg-delegated-subnet",
						AddressPrefix: "10.0.3.0/24",
						Delegation: &AzureSubnetDelegation{
							Name:        "postgresql",
							ServiceName: "Microsoft.DBforPostgreSQL/flexibleServers",
							Actions:     []string{"Microsoft.Network/virtualNetworks/subnets/action"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with delegation without actions", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:          "cae-subnet",
						AddressPrefix: "10.0.4.0/23",
						Delegation: &AzureSubnetDelegation{
							Name:        "container-apps",
							ServiceName: "Microsoft.App/environments",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with private_endpoint_network_policies set to Enabled", func() {
				penp := "Enabled"
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:                           "pe-subnet",
						AddressPrefix:                  "10.0.5.0/24",
						PrivateEndpointNetworkPolicies: &penp,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with private_endpoint_network_policies set to NetworkSecurityGroupEnabled", func() {
				penp := "NetworkSecurityGroupEnabled"
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:                           "nsg-pe-subnet",
						AddressPrefix:                  "10.0.6.0/24",
						PrivateEndpointNetworkPolicies: &penp,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with private_link_service_network_policies_enabled set to false", func() {
				plsnpe := false
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:                                     "pls-subnet",
						AddressPrefix:                            "10.0.7.0/24",
						PrivateLinkServiceNetworkPoliciesEnabled: &plsnpe,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all fields populated", func() {
				penp := "Disabled"
				plsnpe := true
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-subnet",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzureSubnetSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-network-rg",
							},
						},
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-network-rg/providers/Microsoft.Network/virtualNetworks/prod-vnet",
							},
						},
						Name:             "prod-app-subnet",
						AddressPrefix:    "10.0.10.0/24",
						ServiceEndpoints: []string{"Microsoft.Sql", "Microsoft.Storage"},
						Delegation: &AzureSubnetDelegation{
							Name:        "web",
							ServiceName: "Microsoft.Web/serverFarms",
							Actions:     []string{"Microsoft.Network/virtualNetworks/subnets/action"},
						},
						PrivateEndpointNetworkPolicies:           &penp,
						PrivateLinkServiceNetworkPoliciesEnabled: &plsnpe,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_subnet", func() {

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
						VnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
							},
						},
						Name:          "my-subnet",
						AddressPrefix: "10.0.1.0/24",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vnet_id is missing", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "my-rg",
							},
						},
						Name:          "my-subnet",
						AddressPrefix: "10.0.1.0/24",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						AddressPrefix: "10.0.1.0/24",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds maximum length", func() {
				tooLongName := ""
				for len(tooLongName) < 81 {
					tooLongName += "a"
				}
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:          tooLongName,
						AddressPrefix: "10.0.1.0/24",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when address_prefix is missing", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name: "my-subnet",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when private_endpoint_network_policies has invalid value", func() {
				invalid := "InvalidValue"
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:                           "my-subnet",
						AddressPrefix:                  "10.0.1.0/24",
						PrivateEndpointNetworkPolicies: &invalid,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when delegation name is missing", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:          "my-subnet",
						AddressPrefix: "10.0.1.0/24",
						Delegation: &AzureSubnetDelegation{
							ServiceName: "Microsoft.DBforPostgreSQL/flexibleServers",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when delegation service_name is missing", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:          "my-subnet",
						AddressPrefix: "10.0.1.0/24",
						Delegation: &AzureSubnetDelegation{
							Name: "postgresql",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &AzureSubnet{
					ApiVersion: "wrong.version/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:          "my-subnet",
						AddressPrefix: "10.0.1.0/24",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
					Spec: &AzureSubnetSpec{
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
						Name:          "my-subnet",
						AddressPrefix: "10.0.1.0/24",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Spec: &AzureSubnetSpec{
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
						Name:          "my-subnet",
						AddressPrefix: "10.0.1.0/24",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureSubnet{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureSubnet",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-subnet",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
