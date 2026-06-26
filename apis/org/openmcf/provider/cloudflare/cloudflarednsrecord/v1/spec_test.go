package cloudflarednsrecordv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestCloudflareDnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareDnsRecordSpec Custom Validation Tests")
}

// zoneRef is a convenience for building a literal zone_id reference.
func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "abc123def456"}}
}

// record wraps a spec in a full resource for validation.
func record(name string, spec *CloudflareDnsRecordSpec) *CloudflareDnsRecord {
	return &CloudflareDnsRecord{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareDnsRecord",
		Metadata:   &shared.CloudResourceMetadata{Name: name},
		Spec:       spec,
	}
}

var _ = ginkgo.Describe("CloudflareDnsRecordSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("simple (content) records", func() {

			ginkgo.It("accepts a minimal A record", func() {
				err := protovalidate.Validate(record("a", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "www", Type: CloudflareDnsRecordSpec_A, Content: "192.0.2.1",
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts an AAAA record", func() {
				err := protovalidate.Validate(record("aaaa", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "www", Type: CloudflareDnsRecordSpec_AAAA, Content: "2001:db8::1",
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a CNAME record", func() {
				err := protovalidate.Validate(record("cname", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "app", Type: CloudflareDnsRecordSpec_CNAME, Content: "www.example.com",
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts an MX record with priority", func() {
				err := protovalidate.Validate(record("mx", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "@", Type: CloudflareDnsRecordSpec_MX, Content: "mail.example.com", Priority: 10,
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a TXT record", func() {
				err := protovalidate.Validate(record("txt", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "@", Type: CloudflareDnsRecordSpec_TXT, Content: "v=spf1 include:_spf.google.com ~all",
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a PTR record (new type)", func() {
				err := protovalidate.Validate(record("ptr", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "1.2.0.192.in-addr.arpa", Type: CloudflareDnsRecordSpec_PTR, Content: "host.example.com",
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a proxied A record", func() {
				err := protovalidate.Validate(record("proxied", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "www", Type: CloudflareDnsRecordSpec_A, Content: "192.0.2.1", Proxied: true,
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a record with auto TTL (1)", func() {
				err := protovalidate.Validate(record("ttl-auto", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "www", Type: CloudflareDnsRecordSpec_A, Content: "192.0.2.1", Ttl: 1,
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a record with TTL of 30 (Enterprise floor)", func() {
				err := protovalidate.Validate(record("ttl-30", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "www", Type: CloudflareDnsRecordSpec_A, Content: "192.0.2.1", Ttl: 30,
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts tags and settings", func() {
				err := protovalidate.Validate(record("extras", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "www", Type: CloudflareDnsRecordSpec_A, Content: "192.0.2.1",
					Tags:     []string{"team:web", "env:prod"},
					Settings: &CloudflareDnsRecordSettings{Ipv4Only: true},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a long comment (no artificial cap)", func() {
				err := protovalidate.Validate(record("comment", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "www", Type: CloudflareDnsRecordSpec_A, Content: "192.0.2.1",
					Comment: "This is a deliberately long comment that would have exceeded the old 100 character cap which has been removed.",
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("structured (data) records", func() {

			ginkgo.It("accepts an SRV record via data.srv", func() {
				err := protovalidate.Validate(record("srv", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "_sip._tcp", Type: CloudflareDnsRecordSpec_SRV,
					Data: &CloudflareDnsRecordSpec_Srv{Srv: &SrvData{Priority: 10, Weight: 5, Port: 5060, Target: "sip.example.com"}},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a CAA record via data.caa", func() {
				err := protovalidate.Validate(record("caa", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "@", Type: CloudflareDnsRecordSpec_CAA,
					Data: &CloudflareDnsRecordSpec_Caa{Caa: &CaaData{Flags: 0, Tag: "issue", Value: "letsencrypt.org"}},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a DS record via data.ds", func() {
				err := protovalidate.Validate(record("ds", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "sub", Type: CloudflareDnsRecordSpec_DS,
					Data: &CloudflareDnsRecordSpec_Ds{Ds: &DsData{KeyTag: 2371, Algorithm: 13, DigestType: 2, Digest: "ABCDEF0123456789"}},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts an HTTPS record via data.https", func() {
				err := protovalidate.Validate(record("https", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "@", Type: CloudflareDnsRecordSpec_HTTPS,
					Data: &CloudflareDnsRecordSpec_Https{Https: &HttpsData{Priority: 1, Target: ".", Value: "alpn=\"h2,h3\""}},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("required fields", func() {

			ginkgo.It("rejects a missing zone_id", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					Name: "www", Type: CloudflareDnsRecordSpec_A, Content: "192.0.2.1",
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a missing name", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Type: CloudflareDnsRecordSpec_A, Content: "192.0.2.1",
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects an unspecified type", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "www", Type: CloudflareDnsRecordSpec_record_type_unspecified, Content: "192.0.2.1",
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("content / data coherence", func() {

			ginkgo.It("rejects a simple type with neither content nor data", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "www", Type: CloudflareDnsRecordSpec_A,
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a structured type supplied via content", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "@", Type: CloudflareDnsRecordSpec_CAA, Content: "0 issue \"letsencrypt.org\"",
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects setting both content and a data block", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "_sip._tcp", Type: CloudflareDnsRecordSpec_SRV, Content: "10 5 5060 sip.example.com",
					Data: &CloudflareDnsRecordSpec_Srv{Srv: &SrvData{Priority: 10, Weight: 5, Port: 5060, Target: "sip.example.com"}},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a data block that does not match the type", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "@", Type: CloudflareDnsRecordSpec_SRV,
					Data: &CloudflareDnsRecordSpec_Caa{Caa: &CaaData{Tag: "issue", Value: "letsencrypt.org"}},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("field constraints", func() {

			ginkgo.It("rejects a TTL below the floor", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "www", Type: CloudflareDnsRecordSpec_A, Content: "192.0.2.1", Ttl: 10,
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a TTL exceeding the max", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "www", Type: CloudflareDnsRecordSpec_A, Content: "192.0.2.1", Ttl: 100000,
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a negative priority", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "@", Type: CloudflareDnsRecordSpec_MX, Content: "mail.example.com", Priority: -1,
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a priority exceeding the max", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "@", Type: CloudflareDnsRecordSpec_MX, Content: "mail.example.com", Priority: 70000,
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a proxied TXT record", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "@", Type: CloudflareDnsRecordSpec_TXT, Content: "test", Proxied: true,
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a proxied MX record", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "@", Type: CloudflareDnsRecordSpec_MX, Content: "mail.example.com", Priority: 10, Proxied: true,
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects an MX record without priority", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "@", Type: CloudflareDnsRecordSpec_MX, Content: "mail.example.com",
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a CAA flags value over 255", func() {
				err := protovalidate.Validate(record("r", &CloudflareDnsRecordSpec{
					ZoneId: zoneRef(), Name: "@", Type: CloudflareDnsRecordSpec_CAA,
					Data: &CloudflareDnsRecordSpec_Caa{Caa: &CaaData{Flags: 300, Tag: "issue", Value: "letsencrypt.org"}},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
