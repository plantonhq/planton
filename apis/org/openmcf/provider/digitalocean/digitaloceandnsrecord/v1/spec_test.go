package digitaloceandnsrecordv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestDigitalOceanDnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanDnsRecordSpec Custom Validation Tests")
}

// Helper to create StringValueOrRef with direct value
func strVal(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

// Helper function for optional int32 fields
func int32Ptr(i int32) *int32 {
	return &i
}

var _ = ginkgo.Describe("DigitalOceanDnsRecordSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("digitalocean_dns_record", func() {

			ginkgo.It("should not return a validation error for minimal valid A record", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-a-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain: strVal("example.com"),
						Name:   "www",
						Type:   DigitalOceanDnsRecordSpec_A,
						Value:  strVal("192.0.2.1"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for AAAA record", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aaaa-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain: strVal("example.com"),
						Name:   "www",
						Type:   DigitalOceanDnsRecordSpec_AAAA,
						Value:  strVal("2001:db8::1"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CNAME record", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cname-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain: strVal("example.com"),
						Name:   "app",
						Type:   DigitalOceanDnsRecordSpec_CNAME,
						Value:  strVal("target.example.com"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for MX record with priority", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-mx-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain:   strVal("example.com"),
						Name:     "@",
						Type:     DigitalOceanDnsRecordSpec_MX,
						Value:    strVal("mail.example.com"),
						Priority: 10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for TXT record", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-txt-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain: strVal("example.com"),
						Name:   "@",
						Type:   DigitalOceanDnsRecordSpec_TXT,
						Value:  strVal("v=spf1 include:_spf.google.com ~all"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for SRV record with required fields", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-srv-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain:   strVal("example.com"),
						Name:     "_sip._tcp",
						Type:     DigitalOceanDnsRecordSpec_SRV,
						Value:    strVal("sipserver.example.com"),
						Priority: 10,
						Weight:   5,
						Port:     5060,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CAA record with tag", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-caa-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain: strVal("example.com"),
						Name:   "@",
						Type:   DigitalOceanDnsRecordSpec_CAA,
						Value:  strVal("letsencrypt.org"),
						Flags:  0,
						Tag:    "issue",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for record with custom TTL", func() {
				ttlSeconds := int32(3600)
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ttl-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain:     strVal("example.com"),
						Name:       "www",
						Type:       DigitalOceanDnsRecordSpec_A,
						Value:      strVal("192.0.2.1"),
						TtlSeconds: &ttlSeconds,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for root domain record", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-root-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain: strVal("example.com"),
						Name:   "@",
						Type:   DigitalOceanDnsRecordSpec_A,
						Value:  strVal("192.0.2.1"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for NS record", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-ns-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain: strVal("example.com"),
						Name:   "@",
						Type:   DigitalOceanDnsRecordSpec_NS,
						Value:  strVal("ns1.digitalocean.com"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("digitalocean_dns_record", func() {

			ginkgo.It("should return a validation error when domain is missing", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Name:  "www",
						Type:  DigitalOceanDnsRecordSpec_A,
						Value: strVal("192.0.2.1"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain: strVal("example.com"),
						Type:   DigitalOceanDnsRecordSpec_A,
						Value:  strVal("192.0.2.1"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when type is unspecified", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain: strVal("example.com"),
						Name:   "www",
						Type:   DigitalOceanDnsRecordSpec_record_type_unspecified,
						Value:  strVal("192.0.2.1"),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when value is missing", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain: strVal("example.com"),
						Name:   "www",
						Type:   DigitalOceanDnsRecordSpec_A,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for TTL below minimum", func() {
				ttlSeconds := int32(10)
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain:     strVal("example.com"),
						Name:       "www",
						Type:       DigitalOceanDnsRecordSpec_A,
						Value:      strVal("192.0.2.1"),
						TtlSeconds: &ttlSeconds,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for TTL exceeding max", func() {
				ttlSeconds := int32(100000)
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain:     strVal("example.com"),
						Name:       "www",
						Type:       DigitalOceanDnsRecordSpec_A,
						Value:      strVal("192.0.2.1"),
						TtlSeconds: &ttlSeconds,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for negative priority", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain:   strVal("example.com"),
						Name:     "@",
						Type:     DigitalOceanDnsRecordSpec_MX,
						Value:    strVal("mail.example.com"),
						Priority: -1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for priority exceeding max", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain:   strVal("example.com"),
						Name:     "@",
						Type:     DigitalOceanDnsRecordSpec_MX,
						Value:    strVal("mail.example.com"),
						Priority: 70000,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for SRV record without port", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain:   strVal("example.com"),
						Name:     "_sip._tcp",
						Type:     DigitalOceanDnsRecordSpec_SRV,
						Value:    strVal("sipserver.example.com"),
						Priority: 10,
						Weight:   5,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for CAA record without tag", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain: strVal("example.com"),
						Name:   "@",
						Type:   DigitalOceanDnsRecordSpec_CAA,
						Value:  strVal("letsencrypt.org"),
						Flags:  0,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for port exceeding max", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain:   strVal("example.com"),
						Name:     "_sip._tcp",
						Type:     DigitalOceanDnsRecordSpec_SRV,
						Value:    strVal("sipserver.example.com"),
						Priority: 10,
						Weight:   5,
						Port:     70000,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for flags exceeding max", func() {
				input := &DigitalOceanDnsRecord{
					ApiVersion: "digital-ocean.openmcf.org/v1",
					Kind:       "DigitalOceanDnsRecord",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-record",
					},
					Spec: &DigitalOceanDnsRecordSpec{
						Domain: strVal("example.com"),
						Name:   "@",
						Type:   DigitalOceanDnsRecordSpec_CAA,
						Value:  strVal("letsencrypt.org"),
						Flags:  256,
						Tag:    "issue",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
