package civodnsrecordv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestCivoDnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CivoDnsRecordSpec Custom Validation Tests")
}

// Helper function to create StringValueOrRef from a literal value
func strVal(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

var _ = ginkgo.Describe("CivoDnsRecordSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("civo_dns_record_spec", func() {

			ginkgo.It("should not return a validation error for minimal valid A record", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "www",
					Type:   CivoDnsRecordSpec_A,
					Value:  "192.0.2.1",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for AAAA record", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "www",
					Type:   CivoDnsRecordSpec_AAAA,
					Value:  "2001:db8::1",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CNAME record", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "app",
					Type:   CivoDnsRecordSpec_CNAME,
					Value:  "www.example.com",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for MX record with priority", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId:   strVal("zone-abc123"),
					Name:     "@",
					Type:     CivoDnsRecordSpec_MX,
					Value:    "mail.example.com",
					Priority: 10,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for TXT record", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "@",
					Type:   CivoDnsRecordSpec_TXT,
					Value:  "v=spf1 include:_spf.google.com ~all",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for SRV record", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId:   strVal("zone-abc123"),
					Name:     "_sip._tcp",
					Type:     CivoDnsRecordSpec_SRV,
					Value:    "sip.example.com",
					Priority: 10,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for NS record", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "subdomain",
					Type:   CivoDnsRecordSpec_NS,
					Value:  "ns1.example.com",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for record with valid TTL", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "www",
					Type:   CivoDnsRecordSpec_A,
					Value:  "192.0.2.1",
					Ttl:    3600,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for record with TTL at min boundary", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "www",
					Type:   CivoDnsRecordSpec_A,
					Value:  "192.0.2.1",
					Ttl:    60,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for record with TTL at max boundary", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "www",
					Type:   CivoDnsRecordSpec_A,
					Value:  "192.0.2.1",
					Ttl:    86400,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for record with default TTL (0)", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "www",
					Type:   CivoDnsRecordSpec_A,
					Value:  "192.0.2.1",
					Ttl:    0,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for root record using @", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "@",
					Type:   CivoDnsRecordSpec_A,
					Value:  "192.0.2.1",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("civo_dns_record_spec", func() {

			ginkgo.It("should return a validation error when zone_id is missing", func() {
				spec := &CivoDnsRecordSpec{
					Name:  "www",
					Type:  CivoDnsRecordSpec_A,
					Value: "192.0.2.1",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Type:   CivoDnsRecordSpec_A,
					Value:  "192.0.2.1",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when type is unspecified", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "www",
					Type:   CivoDnsRecordSpec_record_type_unspecified,
					Value:  "192.0.2.1",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when value is missing", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "www",
					Type:   CivoDnsRecordSpec_A,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid TTL below minimum", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "www",
					Type:   CivoDnsRecordSpec_A,
					Value:  "192.0.2.1",
					Ttl:    30,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for TTL exceeding max", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "www",
					Type:   CivoDnsRecordSpec_A,
					Value:  "192.0.2.1",
					Ttl:    100000,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for negative priority", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId:   strVal("zone-abc123"),
					Name:     "@",
					Type:     CivoDnsRecordSpec_MX,
					Value:    "mail.example.com",
					Priority: -1,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for priority exceeding max", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId:   strVal("zone-abc123"),
					Name:     "@",
					Type:     CivoDnsRecordSpec_MX,
					Value:    "mail.example.com",
					Priority: 70000,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for MX record without priority", func() {
				spec := &CivoDnsRecordSpec{
					ZoneId: strVal("zone-abc123"),
					Name:   "@",
					Type:   CivoDnsRecordSpec_MX,
					Value:  "mail.example.com",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
