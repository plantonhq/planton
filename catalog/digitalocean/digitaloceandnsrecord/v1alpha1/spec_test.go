package digitaloceandnsrecordv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestDigitalOceanDnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanDnsRecordSpec Custom Validation Tests")
}

// strVal builds a StringValueOrRef carrying a literal value.
func strVal(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

// int32Ptr returns a pointer for optional int32 fields.
func int32Ptr(i int32) *int32 {
	return &i
}

// record returns a valid A record the tests mutate per case.
func record() *DigitalOceanDnsRecord {
	return &DigitalOceanDnsRecord{
		ApiVersion: "digital-ocean.planton.dev/v1alpha1",
		Kind:       "DigitalOceanDnsRecord",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-record",
		},
		Spec: &DigitalOceanDnsRecordSpec{
			Domain: strVal("example.com"),
			Name:   "www",
			Type:   DigitalOceanDnsRecordSpec_A,
			Value:  strVal("192.0.2.1"),
		},
	}
}

var _ = ginkgo.Describe("DigitalOceanDnsRecordSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal A record", func() {
			gomega.Expect(protovalidate.Validate(record())).To(gomega.BeNil())
		})

		ginkgo.It("accepts an AAAA record", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_AAAA
			input.Spec.Value = strVal("2001:db8::1")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a CNAME record", func() {
			input := record()
			input.Spec.Name = "app"
			input.Spec.Type = DigitalOceanDnsRecordSpec_CNAME
			input.Spec.Value = strVal("target.example.com")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an MX record with a priority", func() {
			input := record()
			input.Spec.Name = "@"
			input.Spec.Type = DigitalOceanDnsRecordSpec_MX
			input.Spec.Value = strVal("mail.example.com")
			input.Spec.Priority = int32Ptr(10)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an MX record with an explicit zero priority", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_MX
			input.Spec.Value = strVal("mail.example.com")
			input.Spec.Priority = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a TXT record", func() {
			input := record()
			input.Spec.Name = "@"
			input.Spec.Type = DigitalOceanDnsRecordSpec_TXT
			input.Spec.Value = strVal("v=spf1 include:_spf.google.com ~all")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an SRV record with priority, weight, and port", func() {
			input := record()
			input.Spec.Name = "_sip._tcp"
			input.Spec.Type = DigitalOceanDnsRecordSpec_SRV
			input.Spec.Value = strVal("sip.example.com")
			input.Spec.Priority = int32Ptr(10)
			input.Spec.Weight = int32Ptr(60)
			input.Spec.Port = int32Ptr(5060)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a CAA record with flags and tag", func() {
			input := record()
			input.Spec.Name = "@"
			input.Spec.Type = DigitalOceanDnsRecordSpec_CAA
			input.Spec.Value = strVal("letsencrypt.org")
			input.Spec.Flags = int32Ptr(0)
			input.Spec.Tag = "issue"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an NS record", func() {
			input := record()
			input.Spec.Name = "sub"
			input.Spec.Type = DigitalOceanDnsRecordSpec_NS
			input.Spec.Value = strVal("ns1.example.com")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an SOA record", func() {
			input := record()
			input.Spec.Name = "@"
			input.Spec.Type = DigitalOceanDnsRecordSpec_SOA
			input.Spec.Value = strVal("ns1.digitalocean.com")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts the provider's TTL floor of 1 second", func() {
			input := record()
			input.Spec.TtlSeconds = int32Ptr(1)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a domain reference instead of a literal", func() {
			input := record()
			input.Spec.Domain = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-zone"},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("rejects a missing domain", func() {
			input := record()
			input.Spec.Domain = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing name", func() {
			input := record()
			input.Spec.Name = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a missing value", func() {
			input := record()
			input.Spec.Value = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an unspecified type", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_record_type_unspecified
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a TTL of zero", func() {
			input := record()
			input.Spec.TtlSeconds = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an MX record without priority", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_MX
			input.Spec.Value = strVal("mail.example.com")
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an SRV record without port", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_SRV
			input.Spec.Priority = int32Ptr(10)
			input.Spec.Weight = int32Ptr(60)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an SRV record without weight", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_SRV
			input.Spec.Priority = int32Ptr(10)
			input.Spec.Port = int32Ptr(5060)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an SRV record without priority", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_SRV
			input.Spec.Weight = int32Ptr(60)
			input.Spec.Port = int32Ptr(5060)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a CAA record without tag", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_CAA
			input.Spec.Flags = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a CAA record without flags", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_CAA
			input.Spec.Tag = "issue"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid CAA tag value", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_CAA
			input.Spec.Flags = int32Ptr(0)
			input.Spec.Tag = "invalid"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a negative priority", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_MX
			input.Spec.Priority = int32Ptr(-1)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a priority above 65535", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_MX
			input.Spec.Priority = int32Ptr(65536)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a port above 65535", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_SRV
			input.Spec.Priority = int32Ptr(10)
			input.Spec.Weight = int32Ptr(60)
			input.Spec.Port = int32Ptr(65536)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects flags above 255", func() {
			input := record()
			input.Spec.Type = DigitalOceanDnsRecordSpec_CAA
			input.Spec.Flags = int32Ptr(256)
			input.Spec.Tag = "issue"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
