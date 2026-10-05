package azurepublicipv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzurePublicIpSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePublicIpSpec Validation Tests")
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

// validResource returns a minimal valid AzurePublicIp that individual cases
// then mutate into the shape under test.
func validResource() *AzurePublicIp {
	return &AzurePublicIp{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePublicIp",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-public-ip",
		},
		Spec: &AzurePublicIpSpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "prod-frontend",
		},
	}
}

var _ = ginkgo.Describe("AzurePublicIpSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_public_ip", func() {

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

			ginkgo.It("should accept a zone-redundant standard address", func() {
				input := validResource()
				input.Spec.Sku = AzurePublicIpSku_STANDARD
				input.Spec.Zones = []string{"1", "2", "3"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an IPv6 address", func() {
				input := validResource()
				input.Spec.IpVersion = AzurePublicIpIpVersion_IPV6
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a global-tier frontend on the standard SKU", func() {
				input := validResource()
				input.Spec.Sku = AzurePublicIpSku_STANDARD
				input.Spec.SkuTier = AzurePublicIpSkuTier_GLOBAL
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept allocation from a public IP prefix", func() {
				input := validResource()
				input.Spec.PublicIpPrefixId = ref("prod-egress-prefix")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a DNS label with a reuse scope", func() {
				input := validResource()
				input.Spec.DomainNameLabel = "prod-gateway"
				input.Spec.DomainNameLabelScope = AzurePublicIpDomainNameLabelScope_TENANT_REUSE
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept dedicated DDoS protection with a plan", func() {
				input := validResource()
				input.Spec.DdosProtectionMode = AzurePublicIpDdosProtectionMode_ENABLED
				input.Spec.DdosProtectionPlanId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/ddosProtectionPlans/shield"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept ip_tags, reverse_fqdn, and user tags", func() {
				input := validResource()
				input.Spec.IpTags = map[string]string{"RoutingPreference": "Internet"}
				input.Spec.ReverseFqdn = "mail.example.com"
				input.Spec.Tags = map[string]string{"cost-center": "networking"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_public_ip", func() {

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

			ginkgo.It("should return a validation error when name ends with a period", func() {
				input := validResource()
				input.Spec.Name = "bad."
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a StandardV2 global-tier address", func() {
				input := validResource()
				input.Spec.Sku = AzurePublicIpSku_STANDARD_V2
				input.Spec.SkuTier = AzurePublicIpSkuTier_GLOBAL
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid zone value", func() {
				input := validResource()
				input.Spec.Zones = []string{"4"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed domain name label", func() {
				input := validResource()
				input.Spec.DomainNameLabel = "Bad-Label"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a label scope is set without a label", func() {
				input := validResource()
				input.Spec.DomainNameLabelScope = AzurePublicIpDomainNameLabelScope_NO_REUSE
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when a DDoS plan is set without the enabled mode", func() {
				input := validResource()
				input.Spec.DdosProtectionPlanId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Network/ddosProtectionPlans/shield"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when idle timeout is out of range", func() {
				input := validResource()
				timeout := int32(31)
				input.Spec.IdleTimeoutInMinutes = &timeout
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
