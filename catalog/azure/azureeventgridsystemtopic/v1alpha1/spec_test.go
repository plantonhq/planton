package azureeventgridsystemtopicv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureEventgridSystemTopicSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventgridSystemTopicSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testSourceId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Storage/storageAccounts/appdata"

const testIdentityId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-uai"

// validResource returns a valid system topic that individual cases
// mutate into the shape under test.
func validResource() *AzureEventgridSystemTopic {
	return &AzureEventgridSystemTopic{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventgridSystemTopic",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-egst",
		},
		Spec: &AzureEventgridSystemTopicSpec{
			ResourceGroup:    literal("app-rg"),
			Name:             "appdata-events",
			Region:           "eastus",
			SourceResourceId: literal(testSourceId),
			TopicType:        "Microsoft.Storage.StorageAccounts",
		},
	}
}

var _ = ginkgo.Describe("AzureEventgridSystemTopicSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_eventgrid_system_topic", func() {

			ginkgo.It("should not return a validation error for the minimal shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the Global region for global sources", func() {
				input := validResource()
				input.Spec.Region = "Global"
				input.Spec.SourceResourceId = literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg")
				input.Spec.TopicType = "Microsoft.Resources.ResourceGroups"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a 128-character name", func() {
				input := validResource()
				input.Spec.Name = strings.Repeat("a", 128)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridSystemTopicIdentity{
					Type: AzureEventgridSystemTopicIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity carrying an identity id", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridSystemTopicIdentity{
					Type:        AzureEventgridSystemTopicIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the combined identity mode carrying an identity id", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridSystemTopicIdentity{
					Type:        AzureEventgridSystemTopicIdentityType_SYSTEM_AND_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept tags", func() {
				input := validResource()
				input.Spec.Tags = map[string]string{"team": "platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_eventgrid_system_topic", func() {

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a 2-character name", func() {
				input := validResource()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a 129-character name", func() {
				input := validResource()
				input.Spec.Name = strings.Repeat("a", 129)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with invalid characters", func() {
				input := validResource()
				input.Spec.Name = "app_data.events"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing source resource id", func() {
				input := validResource()
				input.Spec.SourceResourceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing topic type", func() {
				input := validResource()
				input.Spec.TopicType = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an identity block without a flavor", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridSystemTopicIdentity{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridSystemTopicIdentity{
					Type: AzureEventgridSystemTopicIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject the combined identity mode without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridSystemTopicIdentity{
					Type: AzureEventgridSystemTopicIdentityType_SYSTEM_AND_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridSystemTopicIdentity{
					Type:        AzureEventgridSystemTopicIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
