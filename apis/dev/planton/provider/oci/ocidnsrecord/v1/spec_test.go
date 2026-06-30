package ocidnsrecordv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOciDnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OciDnsRecordSpec Validation Tests")
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidRecord() *OciDnsRecord {
	return &OciDnsRecord{
		ApiVersion: "oci.planton.dev/v1",
		Kind:       "OciDnsRecord",
		Metadata: &shared.CloudResourceMetadata{
			Name: "app-example-com-a",
		},
		Spec: &OciDnsRecordSpec{
			ZoneNameOrId: newStringValueOrRef("ocid1.dns-zone.oc1..example"),
			Domain:       "app.example.com",
			Rtype:        "A",
			Items: []*OciDnsRecordSpec_RecordItem{
				{Rdata: "192.0.2.1", Ttl: 300},
			},
		},
	}
}

var _ = ginkgo.Describe("OciDnsRecordSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("oci_dns_rrset", func() {

			ginkgo.It("should not return a validation error for minimal A record", func() {
				input := minimalValidRecord()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with multiple A records", func() {
				input := minimalValidRecord()
				input.Spec.Items = []*OciDnsRecordSpec_RecordItem{
					{Rdata: "192.0.2.1", Ttl: 300},
					{Rdata: "192.0.2.2", Ttl: 300},
					{Rdata: "192.0.2.3", Ttl: 300},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for AAAA record", func() {
				input := minimalValidRecord()
				input.Spec.Rtype = "AAAA"
				input.Spec.Items = []*OciDnsRecordSpec_RecordItem{
					{Rdata: "2001:db8::1", Ttl: 3600},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CNAME record", func() {
				input := minimalValidRecord()
				input.Spec.Domain = "www.example.com"
				input.Spec.Rtype = "CNAME"
				input.Spec.Items = []*OciDnsRecordSpec_RecordItem{
					{Rdata: "app.example.com.", Ttl: 3600},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for MX record", func() {
				input := minimalValidRecord()
				input.Spec.Rtype = "MX"
				input.Spec.Items = []*OciDnsRecordSpec_RecordItem{
					{Rdata: "10 mail1.example.com.", Ttl: 3600},
					{Rdata: "20 mail2.example.com.", Ttl: 3600},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for TXT record", func() {
				input := minimalValidRecord()
				input.Spec.Rtype = "TXT"
				input.Spec.Items = []*OciDnsRecordSpec_RecordItem{
					{Rdata: "\"v=spf1 include:example.com ~all\"", Ttl: 3600},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with view_id for private zone", func() {
				input := minimalValidRecord()
				input.Spec.ViewId = newStringValueOrRef("ocid1.dnsview.oc1..example")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with zone_name_or_id via valueFrom ref", func() {
				input := minimalValidRecord()
				input.Spec.ZoneNameOrId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-dns-zone",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with view_id via valueFrom ref", func() {
				input := minimalValidRecord()
				input.Spec.ViewId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-dns-view",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with TTL of 1 second", func() {
				input := minimalValidRecord()
				input.Spec.Items = []*OciDnsRecordSpec_RecordItem{
					{Rdata: "192.0.2.1", Ttl: 1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with high TTL", func() {
				input := minimalValidRecord()
				input.Spec.Items = []*OciDnsRecordSpec_RecordItem{
					{Rdata: "192.0.2.1", Ttl: 86400},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with zone name instead of OCID", func() {
				input := minimalValidRecord()
				input.Spec.ZoneNameOrId = newStringValueOrRef("example.com")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("oci_dns_rrset", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidRecord()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidRecord()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidRecord()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OciDnsRecord{
					ApiVersion: "oci.planton.dev/v1",
					Kind:       "OciDnsRecord",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when zone_name_or_id is missing", func() {
				input := minimalValidRecord()
				input.Spec.ZoneNameOrId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when domain is empty", func() {
				input := minimalValidRecord()
				input.Spec.Domain = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when rtype is empty", func() {
				input := minimalValidRecord()
				input.Spec.Rtype = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when items is empty", func() {
				input := minimalValidRecord()
				input.Spec.Items = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when item rdata is empty", func() {
				input := minimalValidRecord()
				input.Spec.Items = []*OciDnsRecordSpec_RecordItem{
					{Rdata: "", Ttl: 300},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when item ttl is zero", func() {
				input := minimalValidRecord()
				input.Spec.Items = []*OciDnsRecordSpec_RecordItem{
					{Rdata: "192.0.2.1", Ttl: 0},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when item ttl is negative", func() {
				input := minimalValidRecord()
				input.Spec.Items = []*OciDnsRecordSpec_RecordItem{
					{Rdata: "192.0.2.1", Ttl: -1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

		})
	})
})
