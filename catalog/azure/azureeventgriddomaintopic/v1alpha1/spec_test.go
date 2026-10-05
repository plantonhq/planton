package azureeventgriddomaintopicv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureEventgridDomainTopicSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureEventgridDomainTopicSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testDomainId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.EventGrid/domains/tenant-events"

// validResource returns a valid domain topic that individual cases
// mutate into the shape under test.
func validResource() *AzureEventgridDomainTopic {
	return &AzureEventgridDomainTopic{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureEventgridDomainTopic",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-egdt",
		},
		Spec: &AzureEventgridDomainTopicSpec{
			DomainId: literal(testDomainId),
			Name:     "customer-fabrikam",
		},
	}
}

var _ = ginkgo.Describe("AzureEventgridDomainTopicSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_eventgrid_domain_topic", func() {

			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a 3-character name", func() {
				input := validResource()
				input.Spec.Name = "abc"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a 128-character name", func() {
				input := validResource()
				name := make([]byte, 128)
				for i := range name {
					name[i] = 'a'
				}
				input.Spec.Name = string(name)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_eventgrid_domain_topic", func() {

			ginkgo.It("should reject a missing domain id", func() {
				input := validResource()
				input.Spec.DomainId = nil
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
				name := make([]byte, 129)
				for i := range name {
					name[i] = 'a'
				}
				input.Spec.Name = string(name)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name carrying an underscore", func() {
				input := validResource()
				input.Spec.Name = "customer_fabrikam"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("letters, numbers, and hyphens"))
			})
		})
	})
})
