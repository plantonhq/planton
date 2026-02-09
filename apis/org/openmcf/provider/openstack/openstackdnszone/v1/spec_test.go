package openstackdnszonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestOpenStackDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "OpenStackDnsZoneSpec Validation Tests")
}

func int32Ptr(i int32) *int32 {
	return &i
}

func minimalValidDnsZone() *OpenStackDnsZone {
	return &OpenStackDnsZone{
		ApiVersion: "openstack.openmcf.org/v1",
		Kind:       "OpenStackDnsZone",
		Metadata: &shared.CloudResourceMetadata{
			Name: "my-zone",
		},
		Spec: &OpenStackDnsZoneSpec{
			DomainName: "example.com",
		},
	}
}

var _ = ginkgo.Describe("OpenStackDnsZoneSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("openstack_dns_zone", func() {

			ginkgo.It("should not return a validation error for minimal valid zone", func() {
				input := minimalValidDnsZone()
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for zone with email", func() {
				input := minimalValidDnsZone()
				input.Spec.Email = "admin@example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for zone with description", func() {
				input := minimalValidDnsZone()
				input.Spec.Description = "Production DNS zone"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for zone with ttl", func() {
				input := minimalValidDnsZone()
				input.Spec.Ttl = int32Ptr(3600)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for PRIMARY zone", func() {
				input := minimalValidDnsZone()
				input.Spec.Type = "PRIMARY"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for SECONDARY zone with masters", func() {
				input := minimalValidDnsZone()
				input.Spec.Type = "SECONDARY"
				input.Spec.Masters = []string{"ns1.upstream.com", "ns2.upstream.com"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for zone with region", func() {
				input := minimalValidDnsZone()
				input.Spec.Region = "RegionOne"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for zone with inline records", func() {
				ttl := int32(300)
				input := minimalValidDnsZone()
				input.Spec.Records = []*OpenStackDnsRecord{
					{
						RecordType: OpenStackDnsRecord_A,
						RecordName: "www.example.com.",
						Values:     []string{"192.0.2.1"},
						Ttl:        &ttl,
					},
					{
						RecordType: OpenStackDnsRecord_CNAME,
						RecordName: "api.example.com.",
						Values:     []string{"www.example.com."},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for zone with multiple A records (round-robin)", func() {
				input := minimalValidDnsZone()
				input.Spec.Records = []*OpenStackDnsRecord{
					{
						RecordType: OpenStackDnsRecord_A,
						RecordName: "lb.example.com.",
						Values:     []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for zone with MX record", func() {
				input := minimalValidDnsZone()
				input.Spec.Records = []*OpenStackDnsRecord{
					{
						RecordType: OpenStackDnsRecord_MX,
						RecordName: "example.com.",
						Values:     []string{"10 mail1.example.com.", "20 mail2.example.com."},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for zone with TXT record", func() {
				input := minimalValidDnsZone()
				input.Spec.Records = []*OpenStackDnsRecord{
					{
						RecordType: OpenStackDnsRecord_TXT,
						RecordName: "example.com.",
						Values:     []string{"v=spf1 include:_spf.google.com ~all"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for subdomain domain_name", func() {
				input := minimalValidDnsZone()
				input.Spec.DomainName = "staging.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("openstack_dns_zone", func() {

			ginkgo.It("should return a validation error when api_version is wrong", func() {
				input := minimalValidDnsZone()
				input.ApiVersion = "wrong.openmcf.org/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is wrong", func() {
				input := minimalValidDnsZone()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := minimalValidDnsZone()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := &OpenStackDnsZone{
					ApiVersion: "openstack.openmcf.org/v1",
					Kind:       "OpenStackDnsZone",
					Metadata:   &shared.CloudResourceMetadata{Name: "test"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when domain_name is missing", func() {
				input := minimalValidDnsZone()
				input.Spec.DomainName = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when domain_name is invalid (no TLD)", func() {
				input := minimalValidDnsZone()
				input.Spec.DomainName = "example"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when domain_name has uppercase", func() {
				input := minimalValidDnsZone()
				input.Spec.DomainName = "Example.COM"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when type is invalid", func() {
				input := minimalValidDnsZone()
				input.Spec.Type = "TERTIARY"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when SECONDARY zone has no masters", func() {
				input := minimalValidDnsZone()
				input.Spec.Type = "SECONDARY"
				input.Spec.Masters = []string{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when inline record has unspecified type", func() {
				input := minimalValidDnsZone()
				input.Spec.Records = []*OpenStackDnsRecord{
					{
						RecordType: OpenStackDnsRecord_record_type_unspecified,
						RecordName: "www.example.com.",
						Values:     []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when inline record has empty record_name", func() {
				input := minimalValidDnsZone()
				input.Spec.Records = []*OpenStackDnsRecord{
					{
						RecordType: OpenStackDnsRecord_A,
						RecordName: "",
						Values:     []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when inline record name has no trailing dot", func() {
				input := minimalValidDnsZone()
				input.Spec.Records = []*OpenStackDnsRecord{
					{
						RecordType: OpenStackDnsRecord_A,
						RecordName: "www.example.com",
						Values:     []string{"192.0.2.1"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when inline record has empty values", func() {
				input := minimalValidDnsZone()
				input.Spec.Records = []*OpenStackDnsRecord{
					{
						RecordType: OpenStackDnsRecord_A,
						RecordName: "www.example.com.",
						Values:     []string{},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
