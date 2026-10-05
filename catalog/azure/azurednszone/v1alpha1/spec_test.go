package azurednszonev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDnsZoneSpec Validation Tests")
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

// validResource returns a minimal valid AzureDnsZone that individual
// cases then mutate into the shape under test.
func validResource() *AzureDnsZone {
	return &AzureDnsZone{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDnsZone",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "public-dns",
		},
		Spec: &AzureDnsZoneSpec{
			ZoneName:      "example.com",
			ResourceGroup: literal("my-rg"),
		},
	}
}

var _ = ginkgo.Describe("AzureDnsZoneSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_dns_zone", func() {

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

			ginkgo.It("should accept a delegated subdomain zone name", func() {
				input := validResource()
				input.Spec.ZoneName = "team.platform.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an SOA record customization", func() {
				input := validResource()
				minimumTtl := int64(60)
				input.Spec.SoaRecord = &AzureDnsZoneSoaRecord{
					Email:      "dnsadmin.example.com",
					MinimumTtl: &minimumTtl,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an SOA record with every timer set", func() {
				input := validResource()
				expire := int64(2419200)
				minimum := int64(300)
				refresh := int64(3600)
				retry := int64(300)
				serial := int64(1)
				ttl := int64(3600)
				input.Spec.SoaRecord = &AzureDnsZoneSoaRecord{
					Email:        "hostmaster.example.com",
					ExpireTime:   &expire,
					MinimumTtl:   &minimum,
					RefreshTime:  &refresh,
					RetryTime:    &retry,
					SerialNumber: &serial,
					Ttl:          &ttl,
					Tags:         map[string]string{"team": "network"},
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
		ginkgo.Context("azure_dns_zone", func() {

			ginkgo.It("should return a validation error when zone_name is missing", func() {
				input := validResource()
				input.Spec.ZoneName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a single-label zone name", func() {
				input := validResource()
				input.Spec.ZoneName = "example"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a zone name with a trailing dot", func() {
				input := validResource()
				input.Spec.ZoneName = "example.com."
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a zone name with uppercase labels", func() {
				input := validResource()
				input.Spec.ZoneName = "Example.Com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the SOA record omits its email", func() {
				input := validResource()
				input.Spec.SoaRecord = &AzureDnsZoneSoaRecord{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an SOA email containing an @", func() {
				input := validResource()
				input.Spec.SoaRecord = &AzureDnsZoneSoaRecord{
					Email: "dnsadmin@example.com",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an SOA email with consecutive dots", func() {
				input := validResource()
				input.Spec.SoaRecord = &AzureDnsZoneSoaRecord{
					Email: "dnsadmin..example.com",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an SOA email with a single segment", func() {
				input := validResource()
				input.Spec.SoaRecord = &AzureDnsZoneSoaRecord{
					Email: "dnsadmin",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when zone name plus SOA email exceed 253 characters", func() {
				input := validResource()
				input.Spec.ZoneName = strings.Repeat("a", 60) + "." + strings.Repeat("b", 60) + "." + strings.Repeat("c", 60) + ".com"
				input.Spec.SoaRecord = &AzureDnsZoneSoaRecord{
					Email: strings.Repeat("d", 60) + "." + strings.Repeat("e", 30) + ".example.com",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a negative SOA timer", func() {
				input := validResource()
				expire := int64(-1)
				input.Spec.SoaRecord = &AzureDnsZoneSoaRecord{
					Email:      "dnsadmin.example.com",
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
