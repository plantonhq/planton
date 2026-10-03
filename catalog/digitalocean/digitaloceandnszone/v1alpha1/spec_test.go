package digitaloceandnszonev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/plantonhq/planton/shared/networking/enums/dnsrecordtype"
)

func TestDigitalOceanDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanDnsZoneSpec Custom Validation Tests")
}

// strVal builds a StringValueOrRef carrying a literal value.
func strVal(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

// uint32Ptr returns a pointer for optional uint32 fields.
func uint32Ptr(u uint32) *uint32 {
	return &u
}

// zone returns a minimal valid zone the tests mutate per case.
func zone() *DigitalOceanDnsZone {
	return &DigitalOceanDnsZone{
		ApiVersion: "digital-ocean.planton.dev/v1alpha1",
		Kind:       "DigitalOceanDnsZone",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-dns-zone",
		},
		Spec: &DigitalOceanDnsZoneSpec{
			DomainName: "example.com",
		},
	}
}

// aRecord returns a valid apex A record entry the tests mutate per case.
func aRecord() *DigitalOceanDnsZoneRecord {
	return &DigitalOceanDnsZoneRecord{
		Name:   "@",
		Type:   dnsrecordtype.DnsRecordType_A,
		Values: []*foreignkeyv1.StringValueOrRef{strVal("192.0.2.1")},
	}
}

var _ = ginkgo.Describe("DigitalOceanDnsZoneSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal zone with no records", func() {
			gomega.Expect(protovalidate.Validate(zone())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a zone with an apex A record", func() {
			input := zone()
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{aRecord()}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts multiple values on one record", func() {
			input := zone()
			rec := aRecord()
			rec.Values = []*foreignkeyv1.StringValueOrRef{strVal("192.0.2.1"), strVal("192.0.2.2")}
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{rec}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an MX record with a priority", func() {
			input := zone()
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{{
				Name:     "@",
				Type:     dnsrecordtype.DnsRecordType_MX,
				Values:   []*foreignkeyv1.StringValueOrRef{strVal("mail.example.com")},
				Priority: uint32Ptr(10),
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an SRV record with priority, weight, and port", func() {
			input := zone()
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{{
				Name:     "_sip._tcp",
				Type:     dnsrecordtype.DnsRecordType_SRV,
				Values:   []*foreignkeyv1.StringValueOrRef{strVal("sip.example.com")},
				Priority: uint32Ptr(10),
				Weight:   uint32Ptr(60),
				Port:     uint32Ptr(5060),
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a CAA record with flags and tag", func() {
			input := zone()
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{{
				Name:   "@",
				Type:   dnsrecordtype.DnsRecordType_CAA,
				Values: []*foreignkeyv1.StringValueOrRef{strVal("letsencrypt.org")},
				Flags:  uint32Ptr(0),
				Tag:    "issue",
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a TXT record with a custom TTL", func() {
			input := zone()
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{{
				Name:       "@",
				Type:       dnsrecordtype.DnsRecordType_TXT,
				Values:     []*foreignkeyv1.StringValueOrRef{strVal("v=spf1 -all")},
				TtlSeconds: 300,
			}}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an ip_address seeding an initial apex A record", func() {
			input := zone()
			input.Spec.IpAddress = strVal("192.0.2.10")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a record value referencing another resource", func() {
			input := zone()
			rec := aRecord()
			rec.Values = []*foreignkeyv1.StringValueOrRef{{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{Name: "my-droplet"},
				},
			}}
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{rec}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("rejects a missing domain_name", func() {
			input := zone()
			input.Spec.DomainName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a domain_name that is not a FQDN", func() {
			input := zone()
			input.Spec.DomainName = "not-a-domain"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a record without a name", func() {
			input := zone()
			rec := aRecord()
			rec.Name = ""
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{rec}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a record with no values", func() {
			input := zone()
			rec := aRecord()
			rec.Values = nil
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{rec}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a record with an unspecified type", func() {
			input := zone()
			rec := aRecord()
			rec.Type = dnsrecordtype.DnsRecordType_unspecified
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{rec}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an ALIAS record (unsupported by DigitalOcean)", func() {
			input := zone()
			rec := aRecord()
			rec.Type = dnsrecordtype.DnsRecordType_ALIAS
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{rec}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a PTR record (unsupported by DigitalOcean)", func() {
			input := zone()
			rec := aRecord()
			rec.Type = dnsrecordtype.DnsRecordType_PTR
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{rec}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an MX record without priority", func() {
			input := zone()
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{{
				Name:   "@",
				Type:   dnsrecordtype.DnsRecordType_MX,
				Values: []*foreignkeyv1.StringValueOrRef{strVal("mail.example.com")},
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an SRV record missing port", func() {
			input := zone()
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{{
				Name:     "_sip._tcp",
				Type:     dnsrecordtype.DnsRecordType_SRV,
				Values:   []*foreignkeyv1.StringValueOrRef{strVal("sip.example.com")},
				Priority: uint32Ptr(10),
				Weight:   uint32Ptr(60),
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a CAA record without a tag", func() {
			input := zone()
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{{
				Name:   "@",
				Type:   dnsrecordtype.DnsRecordType_CAA,
				Values: []*foreignkeyv1.StringValueOrRef{strVal("letsencrypt.org")},
				Flags:  uint32Ptr(0),
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid CAA tag value", func() {
			input := zone()
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{{
				Name:   "@",
				Type:   dnsrecordtype.DnsRecordType_CAA,
				Values: []*foreignkeyv1.StringValueOrRef{strVal("letsencrypt.org")},
				Flags:  uint32Ptr(0),
				Tag:    "invalid",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a priority above 65535", func() {
			input := zone()
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{{
				Name:     "@",
				Type:     dnsrecordtype.DnsRecordType_MX,
				Values:   []*foreignkeyv1.StringValueOrRef{strVal("mail.example.com")},
				Priority: uint32Ptr(65536),
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects flags above 255", func() {
			input := zone()
			input.Spec.Records = []*DigitalOceanDnsZoneRecord{{
				Name:   "@",
				Type:   dnsrecordtype.DnsRecordType_CAA,
				Values: []*foreignkeyv1.StringValueOrRef{strVal("letsencrypt.org")},
				Flags:  uint32Ptr(256),
				Tag:    "issue",
			}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
