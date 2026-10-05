package azuredatafactoryv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureDataFactorySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDataFactorySpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testIdentityId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-uai"

const testKeyVaultKeyId = "https://app-vault.vault.azure.net/keys/df-cmk"

const testStorageAccountId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Storage/storageAccounts/appdata"

const testPrivateLinkServiceId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Network/privateLinkServices/app-pls"

const testUuid = "11111111-2222-3333-4444-555555555555"

// validResource returns a valid factory that individual cases mutate
// into the shape under test.
func validResource() *AzureDataFactory {
	return &AzureDataFactory{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDataFactory",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-adf",
		},
		Spec: &AzureDataFactorySpec{
			ResourceGroup: literal("app-rg"),
			Name:          "acme-data-factory",
			Region:        "eastus",
		},
	}
}

var _ = ginkgo.Describe("AzureDataFactorySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_data_factory", func() {

			ginkgo.It("should not return a validation error for the minimal shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept 3-character, 63-character, and hyphenated names", func() {
				input := validResource()
				input.Spec.Name = "ab1"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Name = strings.Repeat("a", 63)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Name = "acme-df-01"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataFactoryIdentity{
					Type: AzureDataFactoryIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the combined identity mode carrying an identity id", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataFactoryIdentity{
					Type:        AzureDataFactoryIdentityType_SYSTEM_AND_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a full GitHub repository binding", func() {
				input := validResource()
				input.Spec.GithubConfiguration = &AzureDataFactoryGithubConfiguration{
					AccountName:    "acme-corp",
					BranchName:     "main",
					GitUrl:         "https://github.mycompany.com",
					RepositoryName: "data-pipelines",
					RootFolder:     "/",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a full Azure DevOps repository binding", func() {
				input := validResource()
				input.Spec.VstsConfiguration = &AzureDataFactoryVstsConfiguration{
					AccountName:    "acme-corp",
					BranchName:     "main",
					ProjectName:    "data-platform",
					RepositoryName: "data-pipelines",
					RootFolder:     "/",
					TenantId:       testUuid,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept global parameters across the type vocabulary", func() {
				input := validResource()
				input.Spec.GlobalParameters = []*AzureDataFactoryGlobalParameter{
					{Name: "env", Type: "String", Value: "prod"},
					{Name: "retries", Type: "Int", Value: "3"},
					{Name: "rate", Type: "Float", Value: "0.5"},
					{Name: "enabled", Type: "Bool", Value: "true"},
					{Name: "regions", Type: "Array", Value: "[\"eastus\",\"westus\"]"},
					{Name: "limits", Type: "Object", Value: "{\"rows\":100}"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a managed private endpoint on a regular ARM target inside the managed vnet", func() {
				input := validResource()
				managedVnetEnabled := true
				input.Spec.ManagedVirtualNetworkEnabled = &managedVnetEnabled
				input.Spec.ManagedPrivateEndpoints = []*AzureDataFactoryManagedPrivateEndpoint{
					{
						Name:             "blob-endpoint",
						TargetResourceId: literal(testStorageAccountId),
						SubresourceName:  "blob",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a managed private endpoint on a Private Link Service target with fqdns", func() {
				input := validResource()
				managedVnetEnabled := true
				input.Spec.ManagedVirtualNetworkEnabled = &managedVnetEnabled
				input.Spec.ManagedPrivateEndpoints = []*AzureDataFactoryManagedPrivateEndpoint{
					{
						Name:             "pls-endpoint",
						TargetResourceId: literal(testPrivateLinkServiceId),
						Fqdns:            []string{"internal.acme.example"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept customer-managed-key encryption with a user-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataFactoryIdentity{
					Type:        AzureDataFactoryIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				input.Spec.CustomerManagedKey = &AzureDataFactoryCustomerManagedKey{
					KeyVaultKeyId:          literal(testKeyVaultKeyId),
					UserAssignedIdentityId: literal(testIdentityId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept both credential flavors with distinct names", func() {
				input := validResource()
				input.Spec.UserManagedIdentityCredentials = []*AzureDataFactoryUserManagedIdentityCredential{
					{
						Name:        "etl-identity",
						IdentityId:  literal(testIdentityId),
						Description: "identity the ETL linked services run as",
						Annotations: []string{"team:data"},
					},
				}
				input.Spec.ServicePrincipalCredentials = []*AzureDataFactoryServicePrincipalCredential{
					{
						Name:               "legacy-sp",
						TenantId:           testUuid,
						ServicePrincipalId: testUuid,
						ServicePrincipalKey: &AzureDataFactoryServicePrincipalKey{
							LinkedServiceName: "keyvault-ls",
							SecretName:        "legacy-sp-key",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_data_factory", func() {

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names outside the format rules", func() {
				input := validResource()
				input.Spec.Name = "ab"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = strings.Repeat("a", 64)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = "-acme-df"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = "acme-df-"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = "acme_df"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject both repository bindings set at once", func() {
				input := validResource()
				input.Spec.GithubConfiguration = &AzureDataFactoryGithubConfiguration{
					AccountName:    "acme-corp",
					BranchName:     "main",
					RepositoryName: "data-pipelines",
					RootFolder:     "/",
				}
				input.Spec.VstsConfiguration = &AzureDataFactoryVstsConfiguration{
					AccountName:    "acme-corp",
					BranchName:     "main",
					ProjectName:    "data-platform",
					RepositoryName: "data-pipelines",
					RootFolder:     "/",
					TenantId:       testUuid,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an identity block without a flavor", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataFactoryIdentity{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataFactoryIdentity{
					Type: AzureDataFactoryIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataFactoryIdentity{
					Type:        AzureDataFactoryIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a VSTS binding whose tenant id is not a UUID", func() {
				input := validResource()
				input.Spec.VstsConfiguration = &AzureDataFactoryVstsConfiguration{
					AccountName:    "acme-corp",
					BranchName:     "main",
					ProjectName:    "data-platform",
					RepositoryName: "data-pipelines",
					RootFolder:     "/",
					TenantId:       "not-a-uuid",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a global parameter with an unknown type or duplicate name", func() {
				input := validResource()
				input.Spec.GlobalParameters = []*AzureDataFactoryGlobalParameter{
					{Name: "env", Type: "Text", Value: "prod"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.GlobalParameters = []*AzureDataFactoryGlobalParameter{
					{Name: "env", Type: "String", Value: "prod"},
					{Name: "env", Type: "String", Value: "dev"},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject customer-managed-key encryption without a user-assigned-capable identity", func() {
				input := validResource()
				input.Spec.CustomerManagedKey = &AzureDataFactoryCustomerManagedKey{
					KeyVaultKeyId:          literal(testKeyVaultKeyId),
					UserAssignedIdentityId: literal(testIdentityId),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Identity = &AzureDataFactoryIdentity{
					Type: AzureDataFactoryIdentityType_SYSTEM_ASSIGNED,
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a customer-managed-key block missing its identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataFactoryIdentity{
					Type:        AzureDataFactoryIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				input.Spec.CustomerManagedKey = &AzureDataFactoryCustomerManagedKey{
					KeyVaultKeyId: literal(testKeyVaultKeyId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate credential names across the two flavors", func() {
				input := validResource()
				input.Spec.UserManagedIdentityCredentials = []*AzureDataFactoryUserManagedIdentityCredential{
					{Name: "shared-name", IdentityId: literal(testIdentityId)},
				}
				input.Spec.ServicePrincipalCredentials = []*AzureDataFactoryServicePrincipalCredential{
					{Name: "shared-name", TenantId: testUuid, ServicePrincipalId: testUuid},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject managed private endpoints without the managed virtual network", func() {
				input := validResource()
				input.Spec.ManagedPrivateEndpoints = []*AzureDataFactoryManagedPrivateEndpoint{
					{
						Name:             "blob-endpoint",
						TargetResourceId: literal(testStorageAccountId),
						SubresourceName:  "blob",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a managed private endpoint setting both arms or neither arm", func() {
				input := validResource()
				managedVnetEnabled := true
				input.Spec.ManagedVirtualNetworkEnabled = &managedVnetEnabled
				input.Spec.ManagedPrivateEndpoints = []*AzureDataFactoryManagedPrivateEndpoint{
					{
						Name:             "both-arms",
						TargetResourceId: literal(testStorageAccountId),
						SubresourceName:  "blob",
						Fqdns:            []string{"internal.acme.example"},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.ManagedPrivateEndpoints = []*AzureDataFactoryManagedPrivateEndpoint{
					{
						Name:             "no-arms",
						TargetResourceId: literal(testStorageAccountId),
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a managed private endpoint with a malformed name or short subresource", func() {
				input := validResource()
				managedVnetEnabled := true
				input.Spec.ManagedVirtualNetworkEnabled = &managedVnetEnabled
				input.Spec.ManagedPrivateEndpoints = []*AzureDataFactoryManagedPrivateEndpoint{
					{
						Name:             "x",
						TargetResourceId: literal(testStorageAccountId),
						SubresourceName:  "blob",
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.ManagedPrivateEndpoints = []*AzureDataFactoryManagedPrivateEndpoint{
					{
						Name:             "blob-endpoint",
						TargetResourceId: literal(testStorageAccountId),
						SubresourceName:  "ab",
					},
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate managed private endpoint names", func() {
				input := validResource()
				managedVnetEnabled := true
				input.Spec.ManagedVirtualNetworkEnabled = &managedVnetEnabled
				input.Spec.ManagedPrivateEndpoints = []*AzureDataFactoryManagedPrivateEndpoint{
					{
						Name:             "blob-endpoint",
						TargetResourceId: literal(testStorageAccountId),
						SubresourceName:  "blob",
					},
					{
						Name:             "blob-endpoint",
						TargetResourceId: literal(testStorageAccountId),
						SubresourceName:  "table",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
