package azurenatgatewayv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureNatGatewaySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureNatGatewaySpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// ref builds a StringValueOrRef carrying a value_from reference.
func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

// validResource returns a minimal valid AzureNatGateway that individual
// cases then mutate into the shape under test.
func validResource() *AzureNatGateway {
	return &AzureNatGateway{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureNatGateway",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-nat",
		},
		Spec: &AzureNatGatewaySpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "prod-egress",
		},
	}
}

var _ = ginkgo.Describe("AzureNatGatewaySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_nat_gateway", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the resource group as a reference", func() {
				input := validResource()
				input.Spec.ResourceGroup = ref("platform-rg")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a zonal standard gateway with public IP references", func() {
				input := validResource()
				input.Spec.SkuName = AzureNatGatewaySkuName_STANDARD
				input.Spec.Zones = []string{"1"}
				input.Spec.PublicIpIds = []*foreignkeyv1.StringValueOrRef{ref("prod-nat-ip")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a zone-redundant StandardV2 gateway without zones", func() {
				input := validResource()
				input.Spec.SkuName = AzureNatGatewaySkuName_STANDARD_V2
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept public IP prefix references and an idle timeout", func() {
				input := validResource()
				timeout := int32(30)
				input.Spec.IdleTimeoutInMinutes = &timeout
				input.Spec.PublicIpPrefixIds = []*foreignkeyv1.StringValueOrRef{
					literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/publicIPPrefixes/egress"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept user tags", func() {
				input := validResource()
				input.Spec.Tags = map[string]string{"cost-center": "networking"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_nat_gateway", func() {

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

			ginkgo.It("should return a validation error when name starts with a hyphen", func() {
				input := validResource()
				input.Spec.Name = "-bad"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when idle timeout is below 4", func() {
				input := validResource()
				timeout := int32(3)
				input.Spec.IdleTimeoutInMinutes = &timeout
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when idle timeout is above 120", func() {
				input := validResource()
				timeout := int32(121)
				input.Spec.IdleTimeoutInMinutes = &timeout
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid zone value", func() {
				input := validResource()
				input.Spec.Zones = []string{"0"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a StandardV2 gateway pins zones", func() {
				input := validResource()
				input.Spec.SkuName = AzureNatGatewaySkuName_STANDARD_V2
				input.Spec.Zones = []string{"1"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := validResource()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := validResource()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := validResource()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := validResource()
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
