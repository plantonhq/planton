package azureuserassignedidentityv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAzureUserAssignedIdentitySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureUserAssignedIdentitySpec Validation Tests")
}

var _ = ginkgo.Describe("AzureUserAssignedIdentitySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_user_assigned_identity", func() {

			ginkgo.It("should not return a validation error for minimal valid fields (no role assignments)", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-identity",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-resource-group",
							},
						},
						Name: "test-identity",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for identity with role assignments", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-identity",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "westeurope",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-identity-rg",
							},
						},
						Name: "prod-platform-identity",
						RoleAssignments: []*RoleAssignment{
							{
								Scope: &foreignkeyv1.StringValueOrRef{
									LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
										Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.KeyVault/vaults/prod-kv",
									},
								},
								RoleDefinitionName: "Key Vault Secrets User",
							},
							{
								Scope: &foreignkeyv1.StringValueOrRef{
									LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
										Value: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/prod-rg/providers/Microsoft.Storage/storageAccounts/prodstorage",
									},
								},
								RoleDefinitionName: "Storage Blob Data Contributor",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with name at minimum length (3 chars)", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "min-name",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "abc",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with role assignment using valueFrom reference", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "ref-test",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "ref-test-identity",
						RoleAssignments: []*RoleAssignment{
							{
								Scope: &foreignkeyv1.StringValueOrRef{
									LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
										ValueFrom: &foreignkeyv1.ValueFromRef{
											Name: "platform-kv",
										},
									},
								},
								RoleDefinitionName: "Key Vault Secrets User",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with subscription-scoped role assignment", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "sub-scope-test",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "sub-scoped-identity",
						RoleAssignments: []*RoleAssignment{
							{
								Scope: &foreignkeyv1.StringValueOrRef{
									LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
										Value: "/subscriptions/00000000-0000-0000-0000-000000000000",
									},
								},
								RoleDefinitionName: "Reader",
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
		ginkgo.Context("azure_user_assigned_identity", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-identity",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-identity",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-identity",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "eastus",
						Name:   "test-identity",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-identity",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is too short (less than 3 chars)", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-identity",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "ab",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when role_assignment scope is missing", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-identity",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-identity",
						RoleAssignments: []*RoleAssignment{
							{
								RoleDefinitionName: "Reader",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when role_assignment role_definition_name is missing", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-identity",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-identity",
						RoleAssignments: []*RoleAssignment{
							{
								Scope: &foreignkeyv1.StringValueOrRef{
									LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
										Value: "/subscriptions/00000000-0000-0000-0000-000000000000",
									},
								},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "wrong.version/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-identity",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-identity",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-identity",
					},
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-identity",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Spec: &AzureUserAssignedIdentitySpec{
						Region: "eastus",
						ResourceGroup: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-rg",
							},
						},
						Name: "test-identity",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureUserAssignedIdentity{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureUserAssignedIdentity",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-identity",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
