package azurepublicipprefixv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzurePublicIpPrefixSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePublicIpPrefixSpec Validation Tests")
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

// validResource returns a minimal valid AzurePublicIpPrefix that individual
// cases then mutate into the shape under test.
func validResource() *AzurePublicIpPrefix {
	return &AzurePublicIpPrefix{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePublicIpPrefix",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-prefix",
		},
		Spec: &AzurePublicIpPrefixSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "test-egress-prefix",
		},
	}
}

var _ = ginkgo.Describe("AzurePublicIpPrefixSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_public_ip_prefix", func() {

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

			ginkgo.It("should accept an explicit prefix length with zone redundancy", func() {
				input := validResource()
				length := int32(30)
				input.Spec.PrefixLength = &length
				input.Spec.Zones = []string{"1", "2", "3"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an IPv6 prefix", func() {
				input := validResource()
				length := int32(124)
				input.Spec.PrefixLength = &length
				input.Spec.IpVersion = AzurePublicIpPrefixIpVersion_IPV6
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a global-tier prefix on the standard SKU", func() {
				input := validResource()
				input.Spec.Sku = AzurePublicIpPrefixSku_STANDARD
				input.Spec.SkuTier = AzurePublicIpPrefixSkuTier_GLOBAL
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a StandardV2 regional prefix", func() {
				input := validResource()
				input.Spec.Sku = AzurePublicIpPrefixSku_STANDARD_V2
				input.Spec.SkuTier = AzurePublicIpPrefixSkuTier_REGIONAL
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a custom IP prefix parent", func() {
				input := validResource()
				input.Spec.CustomIpPrefixId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/customIpPrefixes/byoip"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_public_ip_prefix", func() {

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

			ginkgo.It("should return a validation error when prefix_length is out of range", func() {
				input := validResource()
				length := int32(128)
				input.Spec.PrefixLength = &length
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid zone value", func() {
				input := validResource()
				input.Spec.Zones = []string{"4"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a StandardV2 global-tier prefix", func() {
				input := validResource()
				input.Spec.Sku = AzurePublicIpPrefixSku_STANDARD_V2
				input.Spec.SkuTier = AzurePublicIpPrefixSkuTier_GLOBAL
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
