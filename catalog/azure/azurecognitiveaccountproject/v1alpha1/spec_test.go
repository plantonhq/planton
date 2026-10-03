package azurecognitiveaccountprojectv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureCognitiveAccountProjectSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureCognitiveAccountProjectSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testAccountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.CognitiveServices/accounts/foundry-prod"

// validResource returns a minimal valid project that individual cases
// mutate into the shape under test.
func validResource() *AzureCognitiveAccountProject {
	return &AzureCognitiveAccountProject{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureCognitiveAccountProject",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-cognitive-account-project",
		},
		Spec: &AzureCognitiveAccountProjectSpec{
			CognitiveAccountId: literal(testAccountId),
			Name:               "customer-support",
			Region:             "eastus",
			Identity: &AzureCognitiveAccountProjectIdentity{
				Type: AzureCognitiveAccountProjectIdentityType_SYSTEM_ASSIGNED,
			},
		},
	}
}

var _ = ginkgo.Describe("AzureCognitiveAccountProjectSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_cognitive_account_project", func() {

			ginkgo.It("should not return a validation error for a minimal project", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a description, display name and tags", func() {
				input := validResource()
				input.Spec.Description = "customer support agents and evaluations"
				input.Spec.DisplayName = "Customer Support"
				input.Spec.Tags = map[string]string{"team": "support"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureCognitiveAccountProjectIdentity{
					Type:        AzureCognitiveAccountProjectIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/proj")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a name with periods and underscores", func() {
				input := validResource()
				input.Spec.Name = "team_a.support-2"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_cognitive_account_project", func() {

			ginkgo.It("should reject a missing cognitive account reference", func() {
				input := validResource()
				input.Spec.CognitiveAccountId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing identity", func() {
				input := validResource()
				input.Spec.Identity = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a single-character project name", func() {
				input := validResource()
				input.Spec.Name = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a project name longer than 64 characters", func() {
				input := validResource()
				input.Spec.Name = "a1234567890123456789012345678901234567890123456789012345678901234"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a project name starting with a period", func() {
				input := validResource()
				input.Spec.Name = ".support"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureCognitiveAccountProjectIdentity{
					Type: AzureCognitiveAccountProjectIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureCognitiveAccountProjectIdentity{
					Type:        AzureCognitiveAccountProjectIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/x")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
