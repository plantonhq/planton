package azureeventgridnamespacetopicv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureEventgridNamespaceTopicSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventgridNamespaceTopicSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testNamespaceId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.EventGrid/namespaces/acme-events"

// validResource returns a valid namespace topic that individual cases
// mutate into the shape under test.
func validResource() *AzureEventgridNamespaceTopic {
	return &AzureEventgridNamespaceTopic{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventgridNamespaceTopic",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-egnst",
		},
		Spec: &AzureEventgridNamespaceTopicSpec{
			NamespaceId: literal(testNamespaceId),
			Name:        "orders",
		},
	}
}

var _ = ginkgo.Describe("AzureEventgridNamespaceTopicSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_eventgrid_namespace_topic", func() {

			ginkgo.It("should not return a validation error for the minimal shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept retention at both bounds", func() {
				input := validResource()
				input.Spec.EventRetentionInDays = proto.Int32(1)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.EventRetentionInDays = proto.Int32(7)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a 3-character and a 50-character name", func() {
				input := validResource()
				input.Spec.Name = "abc"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Name = strings.Repeat("a", 50)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_eventgrid_namespace_topic", func() {

			ginkgo.It("should reject a missing namespace id", func() {
				input := validResource()
				input.Spec.NamespaceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a 2-character and a 51-character name", func() {
				input := validResource()
				input.Spec.Name = "ab"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = strings.Repeat("a", 51)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with invalid characters", func() {
				input := validResource()
				input.Spec.Name = "orders.stream"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject retention outside 1-7", func() {
				input := validResource()
				input.Spec.EventRetentionInDays = proto.Int32(0)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.EventRetentionInDays = proto.Int32(8)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
