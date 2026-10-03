package azureeventgridtopicv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureEventgridTopicSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventgridTopicSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func strPtr(s string) *string { return &s }

func boolPtr(b bool) *bool { return &b }

const testIdentityId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/app-uai"

// validResource returns a valid topic that individual cases mutate
// into the shape under test.
func validResource() *AzureEventgridTopic {
	return &AzureEventgridTopic{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventgridTopic",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-egt",
		},
		Spec: &AzureEventgridTopicSpec{
			ResourceGroup: literal("app-rg"),
			Name:          "orders-events",
			Region:        "eastus",
		},
	}
}

var _ = ginkgo.Describe("AzureEventgridTopicSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_eventgrid_topic", func() {

			ginkgo.It("should not return a validation error for the minimal shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept each input schema token", func() {
				for _, schema := range []string{"EventGridSchema", "CloudEventSchemaV1_0", "CustomEventSchema"} {
					input := validResource()
					input.Spec.InputSchema = strPtr(schema)
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a custom schema with input mappings", func() {
				input := validResource()
				input.Spec.InputSchema = strPtr("CustomEventSchema")
				input.Spec.InputMappingFields = &AzureEventgridTopicInputMappingFields{
					Id:        "payloadId",
					EventType: "action",
					Subject:   "entity",
					EventTime: "occurredAt",
				}
				input.Spec.InputMappingDefaultValues = &AzureEventgridTopicInputMappingDefaultValues{
					EventType:   "app.event",
					DataVersion: "1.0",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the locked-down shape", func() {
				input := validResource()
				input.Spec.PublicNetworkAccessEnabled = boolPtr(false)
				input.Spec.LocalAuthEnabled = boolPtr(false)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept inbound IP rules as CIDR ranges and single addresses", func() {
				input := validResource()
				input.Spec.InboundIpRules = []string{"203.0.113.0/24", "198.51.100.7"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridTopicIdentity{
					Type: AzureEventgridTopicIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity carrying an identity id", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridTopicIdentity{
					Type:        AzureEventgridTopicIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept tags", func() {
				input := validResource()
				input.Spec.Tags = map[string]string{"team": "orders"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept 3-character and 50-character names", func() {
				input := validResource()
				input.Spec.Name = "abc"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Name = "a123456789012345678901234567890123456789012345678b"[:50]
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_eventgrid_topic", func() {

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

			ginkgo.It("should reject a 51-character name", func() {
				input := validResource()
				input.Spec.Name = "a12345678901234567890123456789012345678901234567890"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name carrying an underscore", func() {
				input := validResource()
				input.Spec.Name = "orders_events"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("letters, numbers, and hyphens"))
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown input schema token", func() {
				input := validResource()
				input.Spec.InputSchema = strPtr("CloudEventSchemaV2_0")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty inbound rule entry", func() {
				input := validResource()
				input.Spec.InboundIpRules = []string{""}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an identity block without a flavor", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridTopicIdentity{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridTopicIdentity{
					Type: AzureEventgridTopicIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("identity_ids"))
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureEventgridTopicIdentity{
					Type:        AzureEventgridTopicIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("identity_ids"))
			})
		})
	})
})
