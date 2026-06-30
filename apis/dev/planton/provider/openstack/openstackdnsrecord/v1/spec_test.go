package openstackdnsrecordv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestOpenStackDnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackDnsRecordSpec Validation Tests")
}

func int32Ptr(i int32) *int32 {
	return &i
}

func newStringValueOrRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
			Value: value,
		},
	}
}

func minimalValidDnsRecord() *OpenStackDnsRecord {
	return &OpenStackDnsRecord{
		ApiVersion: "openstack.planton.dev/v1",
		Kind:       "OpenStackDnsRecord",
		Metadata: &shared.CloudResourceMetadata{
			Name: "app-a-record",
		},
		Spec: &OpenStackDnsRecordSpec{
			ZoneId:     newStringValueOrRef("zone-uuid-1234"),
			RecordName: "app.example.com.",
			Type:       OpenStackDnsRecordSpec_A,
			Values:     []string{"192.0.2.1"},
		},
	}
}

var _ = ginkgo.Describe("OpenStackDnsRecordSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_dns_record", func() {

			ginkgo.It("should not return a validation error for minimal valid record", func() {
				input := minimalValidDnsRecord()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for A record with multiple values", func() {
				input := minimalValidDnsRecord()
				input.Spec.Values = []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for AAAA record", func() {
				input := minimalValidDnsRecord()
				input.Spec.Type = OpenStackDnsRecordSpec_AAAA
				input.Spec.Values = []string{"2001:db8::1"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CNAME record", func() {
				input := minimalValidDnsRecord()
				input.Spec.Type = OpenStackDnsRecordSpec_CNAME
				input.Spec.RecordName = "www.example.com."
				input.Spec.Values = []string{"example.com."}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for MX record", func() {
				input := minimalValidDnsRecord()
				input.Spec.Type = OpenStackDnsRecordSpec_MX
				input.Spec.RecordName = "example.com."
				input.Spec.Values = []string{"10 mail1.example.com.", "20 mail2.example.com."}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for TXT record", func() {
				input := minimalValidDnsRecord()
				input.Spec.Type = OpenStackDnsRecordSpec_TXT
				input.Spec.RecordName = "example.com."
				input.Spec.Values = []string{"v=spf1 include:_spf.google.com ~all"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for SRV record", func() {
				input := minimalValidDnsRecord()
				input.Spec.Type = OpenStackDnsRecordSpec_SRV
				input.Spec.RecordName = "_sip._tcp.example.com."
				input.Spec.Values = []string{"10 60 5060 sip.example.com."}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for NS record", func() {
				input := minimalValidDnsRecord()
				input.Spec.Type = OpenStackDnsRecordSpec_NS
				input.Spec.RecordName = "sub.example.com."
				input.Spec.Values = []string{"ns1.example.com.", "ns2.example.com."}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for PTR record", func() {
				input := minimalValidDnsRecord()
				input.Spec.Type = OpenStackDnsRecordSpec_PTR
				input.Spec.RecordName = "1.2.0.192.in-addr.arpa."
				input.Spec.Values = []string{"host.example.com."}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CAA record", func() {
				input := minimalValidDnsRecord()
				input.Spec.Type = OpenStackDnsRecordSpec_CAA
				input.Spec.RecordName = "example.com."
				input.Spec.Values = []string{"0 issue \"letsencrypt.org\""}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for record with ttl", func() {
				input := minimalValidDnsRecord()
				input.Spec.Ttl = int32Ptr(3600)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for record with description", func() {
				input := minimalValidDnsRecord()
				input.Spec.Description = "Application A record"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for record with region", func() {
				input := minimalValidDnsRecord()
				input.Spec.Region = "RegionTwo"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for wildcard record", func() {
				input := minimalValidDnsRecord()
				input.Spec.RecordName = "*.example.com."
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with value_from ref for zone_id", func() {
				input := minimalValidDnsRecord()
				input.Spec.ZoneId = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Name: "my-zone",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_dns_record", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidDnsRecord()
				input.ApiVersion = "wrong.planton.dev/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidDnsRecord()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidDnsRecord()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackDnsRecord{
					ApiVersion: "openstack.planton.dev/v1",
					Kind:       "OpenStackDnsRecord",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when zone_id is missing", func() {
				input := minimalValidDnsRecord()
				input.Spec.ZoneId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when record_name is empty", func() {
				input := minimalValidDnsRecord()
				input.Spec.RecordName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when record_name has no trailing dot", func() {
				input := minimalValidDnsRecord()
				input.Spec.RecordName = "app.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when type is unspecified", func() {
				input := minimalValidDnsRecord()
				input.Spec.Type = OpenStackDnsRecordSpec_record_type_unspecified
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when values is empty", func() {
				input := minimalValidDnsRecord()
				input.Spec.Values = []string{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
