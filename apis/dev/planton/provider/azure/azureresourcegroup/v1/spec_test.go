package azureresourcegroupv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestAzureResourceGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureResourceGroupSpec Validation Tests")
}

var _ = ginkgo.Describe("AzureResourceGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_resource_group", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AzureResourceGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureResourceGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-resource-group",
					},
					Spec: &AzureResourceGroupSpec{
						Name:   "my-resource-group",
						Region: "eastus",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with full metadata", func() {
				input := &AzureResourceGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureResourceGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "prod-rg",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &AzureResourceGroupSpec{
						Name:   "prod-platform-rg",
						Region: "westeurope",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with maximum length name", func() {
				// Azure allows up to 90 characters for resource group names
				longName := "a"
				for len(longName) < 90 {
					longName += "a"
				}
				input := &AzureResourceGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureResourceGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "long-name-rg",
					},
					Spec: &AzureResourceGroupSpec{
						Name:   longName,
						Region: "eastus",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_resource_group", func() {

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &AzureResourceGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureResourceGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-rg",
					},
					Spec: &AzureResourceGroupSpec{
						Region: "eastus",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is empty string", func() {
				input := &AzureResourceGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureResourceGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-rg",
					},
					Spec: &AzureResourceGroupSpec{
						Name:   "",
						Region: "eastus",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzureResourceGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureResourceGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-rg",
					},
					Spec: &AzureResourceGroupSpec{
						Name: "my-resource-group",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when region is empty string", func() {
				input := &AzureResourceGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureResourceGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-rg",
					},
					Spec: &AzureResourceGroupSpec{
						Name:   "my-resource-group",
						Region: "",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds maximum length", func() {
				// Azure limit is 90 characters
				tooLongName := ""
				for len(tooLongName) < 91 {
					tooLongName += "a"
				}
				input := &AzureResourceGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureResourceGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-rg",
					},
					Spec: &AzureResourceGroupSpec{
						Name:   tooLongName,
						Region: "eastus",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := &AzureResourceGroup{
					ApiVersion: "wrong.version/v1",
					Kind:       "AzureResourceGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-rg",
					},
					Spec: &AzureResourceGroupSpec{
						Name:   "my-resource-group",
						Region: "eastus",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := &AzureResourceGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "WrongKind",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-rg",
					},
					Spec: &AzureResourceGroupSpec{
						Name:   "my-resource-group",
						Region: "eastus",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := &AzureResourceGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureResourceGroup",
					Spec: &AzureResourceGroupSpec{
						Name:   "my-resource-group",
						Region: "eastus",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &AzureResourceGroup{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureResourceGroup",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-rg",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
