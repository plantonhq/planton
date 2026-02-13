package azurecontainerregistryv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}

func TestAzureContainerRegistrySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureContainerRegistrySpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AzureContainerRegistrySpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_container_registry with minimal configuration", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AzureContainerRegistry{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureContainerRegistry",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-container-registry",
					},
					Spec: &AzureContainerRegistrySpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						RegistryName:  "testregistry123",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for Basic SKU", func() {
				input := &AzureContainerRegistry{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureContainerRegistry",
					Metadata: &shared.CloudResourceMetadata{
						Name: "basic-acr",
					},
					Spec: &AzureContainerRegistrySpec{
						Region:        "westus",
						ResourceGroup: stringRef("test-rg"),
						RegistryName:  "basicacr123",
						Sku:           AzureContainerRegistrySku_BASIC,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for Standard SKU", func() {
				input := &AzureContainerRegistry{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureContainerRegistry",
					Metadata: &shared.CloudResourceMetadata{
						Name: "standard-acr",
					},
					Spec: &AzureContainerRegistrySpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						RegistryName:  "standardacr123",
						Sku:           AzureContainerRegistrySku_STANDARD,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for Premium SKU with geo-replication", func() {
				input := &AzureContainerRegistry{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureContainerRegistry",
					Metadata: &shared.CloudResourceMetadata{
						Name: "premium-acr",
					},
					Spec: &AzureContainerRegistrySpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						RegistryName:  "premiumacr123",
						Sku:           AzureContainerRegistrySku_PREMIUM,
						GeoReplicationRegions: []string{
							"westeurope",
							"southeastasia",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with admin user enabled", func() {
				input := &AzureContainerRegistry{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureContainerRegistry",
					Metadata: &shared.CloudResourceMetadata{
						Name: "admin-acr",
					},
					Spec: &AzureContainerRegistrySpec{
						Region:           "eastus",
						ResourceGroup:    stringRef("test-rg"),
						RegistryName:     "adminacr123",
						Sku:              AzureContainerRegistrySku_STANDARD,
						AdminUserEnabled: true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_container_registry", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzureContainerRegistry{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureContainerRegistry",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-acr",
					},
					Spec: &AzureContainerRegistrySpec{
						ResourceGroup: stringRef("test-rg"),
						RegistryName:  "testacr123",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when registry_name is missing", func() {
				input := &AzureContainerRegistry{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureContainerRegistry",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-acr",
					},
					Spec: &AzureContainerRegistrySpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when registry_name is too short", func() {
				input := &AzureContainerRegistry{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureContainerRegistry",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-acr",
					},
					Spec: &AzureContainerRegistrySpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						RegistryName:  "acr", // Only 3 chars, minimum is 5
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when registry_name is too long", func() {
				input := &AzureContainerRegistry{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureContainerRegistry",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-acr",
					},
					Spec: &AzureContainerRegistrySpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						RegistryName:  "thisregistrynameiswaytoolongandexceedsfiftycharacterslimit", // > 50 chars
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when registry_name contains uppercase letters", func() {
				input := &AzureContainerRegistry{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureContainerRegistry",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-acr",
					},
					Spec: &AzureContainerRegistrySpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						RegistryName:  "MyACR123", // Contains uppercase - not allowed
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when registry_name contains special characters", func() {
				input := &AzureContainerRegistry{
					ApiVersion: "azure.openmcf.org/v1",
					Kind:       "AzureContainerRegistry",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-acr",
					},
					Spec: &AzureContainerRegistrySpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						RegistryName:  "my-acr-123", // Contains dashes - not allowed
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
