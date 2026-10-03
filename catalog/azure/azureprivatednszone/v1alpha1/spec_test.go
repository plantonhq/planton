package azureprivatednszonev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzurePrivateDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePrivateDnsZoneSpec Validation Tests")
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

// validResource returns a minimal valid AzurePrivateDnsZone that individual
// cases then mutate into the shape under test.
func validResource() *AzurePrivateDnsZone {
	return &AzurePrivateDnsZone{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePrivateDnsZone",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "pg-private-dns",
		},
		Spec: &AzurePrivateDnsZoneSpec{
			ResourceGroup: literal("my-rg"),
			Name:          "privatelink.postgres.database.azure.com",
		},
	}
}

var _ = ginkgo.Describe("AzurePrivateDnsZoneSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_private_dns_zone", func() {

			ginkgo.It("should not return a validation error for minimal valid fields (privatelink zone)", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the resource group as a reference", func() {
				input := validResource()
				input.Spec.ResourceGroup = ref("platform-rg")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a custom internal zone name", func() {
				input := validResource()
				input.Spec.Name = "corp.internal"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a single-label zone name", func() {
				input := validResource()
				input.Spec.Name = "internal"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an SOA record customization", func() {
				input := validResource()
				minimumTtl := int64(30)
				input.Spec.SoaRecord = &AzurePrivateDnsZoneSoaRecord{
					Email:      "dnsadmin.contoso.com",
					MinimumTtl: &minimumTtl,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept user tags", func() {
				input := validResource()
				input.Spec.Tags = map[string]string{
					"cost-center": "platform",
					"owner":       "network-team",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_private_dns_zone", func() {

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

			ginkgo.It("should return a validation error for an invalid zone name", func() {
				input := validResource()
				input.Spec.Name = "Not A Zone Name!"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a zone name with an uppercase label", func() {
				input := validResource()
				input.Spec.Name = "Corp.Internal"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the SOA record omits its email", func() {
				input := validResource()
				input.Spec.SoaRecord = &AzurePrivateDnsZoneSoaRecord{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a negative SOA timer", func() {
				input := validResource()
				expire := int64(-1)
				input.Spec.SoaRecord = &AzurePrivateDnsZoneSoaRecord{
					Email:      "dnsadmin.contoso.com",
					ExpireTime: &expire,
				}
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
