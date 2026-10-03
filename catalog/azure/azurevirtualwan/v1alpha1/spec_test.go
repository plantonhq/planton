package azurevirtualwanv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureVirtualWanSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureVirtualWanSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// breakout returns a pointer to the given breakout-category enum value
// (the field is optional, so the generated Go type is a pointer).
func breakout(value AzureVirtualWanOffice365BreakoutCategory) *AzureVirtualWanOffice365BreakoutCategory {
	return &value
}

// strPtr returns a pointer to the given string (for optional fields).
func strPtr(value string) *string {
	return &value
}

// boolPtr returns a pointer to the given bool (for optional fields).
func boolPtr(value bool) *bool {
	return &value
}

// validResource returns a minimal valid Standard WAN that individual
// cases mutate into the shape under test.
func validResource() *AzureVirtualWan {
	return &AzureVirtualWan{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureVirtualWan",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-vwan",
		},
		Spec: &AzureVirtualWanSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "global-wan",
		},
	}
}

var _ = ginkgo.Describe("AzureVirtualWanSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_virtual_wan", func() {

			ginkgo.It("should not return a validation error for a minimal Standard WAN", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a fully-tuned WAN", func() {
				input := validResource()
				input.Spec.DisableVpnEncryption = true
				input.Spec.AllowBranchToBranchTraffic = boolPtr(false)
				input.Spec.Office365LocalBreakoutCategory = breakout(AzureVirtualWanOffice365BreakoutCategory_OPTIMIZE)
				input.Spec.Type = strPtr("Basic")
				input.Spec.Tags = map[string]string{"cost-center": "networking"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every breakout category", func() {
				for _, category := range []AzureVirtualWanOffice365BreakoutCategory{
					AzureVirtualWanOffice365BreakoutCategory_NONE,
					AzureVirtualWanOffice365BreakoutCategory_ALL,
					AzureVirtualWanOffice365BreakoutCategory_OPTIMIZE,
					AzureVirtualWanOffice365BreakoutCategory_OPTIMIZE_AND_ALLOW,
				} {
					input := validResource()
					input.Spec.Office365LocalBreakoutCategory = breakout(category)
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_virtual_wan", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an undefined breakout category", func() {
				input := validResource()
				input.Spec.Office365LocalBreakoutCategory = breakout(AzureVirtualWanOffice365BreakoutCategory(99))
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
