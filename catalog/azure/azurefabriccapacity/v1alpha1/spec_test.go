package azurefabriccapacityv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureFabricCapacitySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFabricCapacitySpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// validResource returns a valid capacity that individual cases mutate
// into the shape under test.
func validResource() *AzureFabricCapacity {
	return &AzureFabricCapacity{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFabricCapacity",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-fabric",
		},
		Spec: &AzureFabricCapacitySpec{
			ResourceGroup:         literal("app-rg"),
			Name:                  "acmefabric",
			Region:                "eastus",
			SkuName:               "F2",
			AdministrationMembers: []string{"admin@acme.example"},
		},
	}
}

var _ = ginkgo.Describe("AzureFabricCapacitySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_fabric_capacity", func() {

			ginkgo.It("should not return a validation error for the minimal shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept 3-character and 63-character names", func() {
				input := validResource()
				input.Spec.Name = "ab1"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Name = "a" + strings.Repeat("b", 62)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept every F-SKU in the vocabulary", func() {
				input := validResource()
				for _, sku := range []string{"F2", "F4", "F8", "F16", "F32", "F64", "F128", "F256", "F512", "F1024", "F2048"} {
					input.Spec.SkuName = sku
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "expected %q to be accepted", sku)
				}
			})

			ginkgo.It("should accept multiple distinct administrators", func() {
				input := validResource()
				input.Spec.AdministrationMembers = []string{
					"admin@acme.example",
					"11111111-2222-3333-4444-555555555555",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_fabric_capacity", func() {

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names outside the format rules", func() {
				input := validResource()
				input.Spec.Name = "ab"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = "a" + strings.Repeat("b", 63)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = "1fabric"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = "AcmeFabric"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = "acme-fabric"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing or unknown SKU", func() {
				input := validResource()
				input.Spec.SkuName = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.SkuName = "F3"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty administrators list", func() {
				input := validResource()
				input.Spec.AdministrationMembers = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject duplicate or empty administrator entries", func() {
				input := validResource()
				input.Spec.AdministrationMembers = []string{"admin@acme.example", "admin@acme.example"}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.AdministrationMembers = []string{""}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
