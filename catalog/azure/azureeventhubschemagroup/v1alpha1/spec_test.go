package azureeventhubschemagroupv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureEventHubSchemaGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventHubSchemaGroupSpec Validation Tests")
}

func minimalSchemaGroup() *AzureEventHubSchemaGroup {
	return &AzureEventHubSchemaGroup{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventHubSchemaGroup",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-schema-group",
		},
		Spec: &AzureEventHubSchemaGroupSpec{
			NamespaceId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "/subscriptions/s/resourceGroups/rg/providers/Microsoft.EventHub/namespaces/myapp-eh",
				},
			},
			SchemaGroupName:     "order-events",
			SchemaCompatibility: AzureEventHubSchemaCompatibility_BACKWARD,
			SchemaType:          AzureEventHubSchemaType_AVRO,
		},
	}
}

var _ = ginkgo.Describe("AzureEventHubSchemaGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_event_hub_schema_group", func() {

			ginkgo.It("should accept a minimal schema group", func() {
				gomega.Expect(protovalidate.Validate(minimalSchemaGroup())).To(gomega.BeNil())
			})

			ginkgo.It("should accept a namespace reference by valueFrom", func() {
				input := minimalSchemaGroup()
				input.Spec.NamespaceId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureEventHubNamespace,
							Name:      "shared-eventhubs",
							FieldPath: "status.outputs.namespace_id",
						},
					},
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept every declared compatibility policy", func() {
				for _, c := range []AzureEventHubSchemaCompatibility{
					AzureEventHubSchemaCompatibility_NONE,
					AzureEventHubSchemaCompatibility_BACKWARD,
					AzureEventHubSchemaCompatibility_FORWARD,
				} {
					input := minimalSchemaGroup()
					input.Spec.SchemaCompatibility = c
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept both declared schema types", func() {
				for _, schemaType := range []AzureEventHubSchemaType{
					AzureEventHubSchemaType_AVRO,
					AzureEventHubSchemaType_JSON,
				} {
					input := minimalSchemaGroup()
					input.Spec.SchemaType = schemaType
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a single-character group name", func() {
				input := minimalSchemaGroup()
				input.Spec.SchemaGroupName = "g"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a 256-character group name", func() {
				input := minimalSchemaGroup()
				input.Spec.SchemaGroupName = "a" + strings.Repeat("b", 254) + "c"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_event_hub_schema_group", func() {

			ginkgo.It("should reject a missing namespace reference", func() {
				input := minimalSchemaGroup()
				input.Spec.NamespaceId = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing group name", func() {
				input := minimalSchemaGroup()
				input.Spec.SchemaGroupName = ""
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a group name over 256 characters", func() {
				input := minimalSchemaGroup()
				input.Spec.SchemaGroupName = strings.Repeat("a", 257)
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a group name with illegal characters", func() {
				input := minimalSchemaGroup()
				input.Spec.SchemaGroupName = "bad name!"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an unspecified compatibility policy", func() {
				input := minimalSchemaGroup()
				input.Spec.SchemaCompatibility = AzureEventHubSchemaCompatibility_azure_event_hub_schema_compatibility_unspecified
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an unspecified schema type", func() {
				input := minimalSchemaGroup()
				input.Spec.SchemaType = AzureEventHubSchemaType_azure_event_hub_schema_type_unspecified
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject a missing metadata block", func() {
				input := minimalSchemaGroup()
				input.Metadata = nil
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})

			ginkgo.It("should reject an incorrect kind", func() {
				input := minimalSchemaGroup()
				input.Kind = "WrongKind"
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
			})
		})
	})
})
