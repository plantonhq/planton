package azureavailabilitysetv1alpha1

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

func TestAzureAvailabilitySetSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureAvailabilitySetSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// validResource returns a valid availability set that individual cases
// mutate into the shape under test.
func validResource() *AzureAvailabilitySet {
	return &AzureAvailabilitySet{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureAvailabilitySet",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-avset",
		},
		Spec: &AzureAvailabilitySetSpec{
			ResourceGroup: literal("app-rg"),
			Name:          "app-avset",
			Region:        "eastus",
		},
	}
}

var _ = ginkgo.Describe("AzureAvailabilitySetSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_availability_set", func() {

			ginkgo.It("should not return a validation error for the minimal shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept domain counts at both bounds", func() {
				input := validResource()
				input.Spec.PlatformUpdateDomainCount = proto.Int32(1)
				input.Spec.PlatformFaultDomainCount = proto.Int32(1)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.PlatformUpdateDomainCount = proto.Int32(20)
				input.Spec.PlatformFaultDomainCount = proto.Int32(3)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept managed set explicitly true or false", func() {
				input := validResource()
				input.Spec.Managed = proto.Bool(true)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Managed = proto.Bool(false)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a proximity placement group and tags", func() {
				input := validResource()
				input.Spec.ProximityPlacementGroupId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/proximityPlacementGroups/ppg")
				input.Spec.Tags = map[string]string{"tier": "web"}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept names at the 80-character cap and single-character names", func() {
				input := validResource()
				input.Spec.Name = "a" + strings.Repeat("b", 78) + "_"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.Name = "a"
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_availability_set", func() {

			ginkgo.It("should reject a missing resource group, name, or region", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Name = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Region = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names that start or end outside the provider's rule", func() {
				input := validResource()
				input.Spec.Name = "-avset"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = "avset."
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = "avset-"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names over 80 characters", func() {
				input := validResource()
				input.Spec.Name = "a" + strings.Repeat("b", 80)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject domain counts outside the provider bounds", func() {
				input := validResource()
				input.Spec.PlatformUpdateDomainCount = proto.Int32(0)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.PlatformUpdateDomainCount = proto.Int32(21)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.PlatformFaultDomainCount = proto.Int32(0)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.PlatformFaultDomainCount = proto.Int32(4)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
