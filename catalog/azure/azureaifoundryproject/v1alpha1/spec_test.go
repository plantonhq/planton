package azureaifoundryprojectv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureAiFoundryProjectSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureAiFoundryProjectSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testHubId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/team-hub"
	testUaiId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/proj-uai"
)

// validResource returns a minimal valid project that individual cases
// mutate into the shape under test.
func validResource() *AzureAiFoundryProject {
	return &AzureAiFoundryProject{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureAiFoundryProject",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ai-foundry-project",
		},
		Spec: &AzureAiFoundryProjectSpec{
			Region:          "eastus",
			Name:            "team-project",
			AiServicesHubId: literal(testHubId),
		},
	}
}

var _ = ginkgo.Describe("AzureAiFoundryProjectSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_ai_foundry_project", func() {

			ginkgo.It("should not return a validation error for a minimal project (identity is optional)", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureAiFoundryProjectIdentity{
					Type: AzureAiFoundryProjectIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with a primary identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureAiFoundryProjectIdentity{
					Type:        AzureAiFoundryProjectIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testUaiId)},
				}
				input.Spec.PrimaryUserAssignedIdentity = literal(testUaiId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a name with underscores (the provider's code regex allows them)", func() {
				input := validResource()
				input.Spec.Name = "team_project_01"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the descriptive surface", func() {
				input := validResource()
				input.Spec.Description = "The fraud-detection team's Foundry project"
				input.Spec.FriendlyName = "Fraud Detection"
				input.Spec.HighBusinessImpactEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_ai_foundry_project", func() {

			ginkgo.It("should reject a project without a hub reference", func() {
				input := validResource()
				input.Spec.AiServicesHubId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name shorter than 3 characters", func() {
				input := validResource()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name longer than 33 characters", func() {
				input := validResource()
				input.Spec.Name = "a123456789012345678901234567890123"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a primary identity without the identity block", func() {
				input := validResource()
				input.Spec.PrimaryUserAssignedIdentity = literal(testUaiId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity_ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureAiFoundryProjectIdentity{
					Type: AzureAiFoundryProjectIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity_ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureAiFoundryProjectIdentity{
					Type:        AzureAiFoundryProjectIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testUaiId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a project without a region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
