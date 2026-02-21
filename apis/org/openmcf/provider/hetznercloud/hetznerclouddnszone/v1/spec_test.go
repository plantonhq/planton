package hetznerclouddnszonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestHetznerCloudDnsZoneSpec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HetznerCloudDnsZoneSpec Validation Suite")
}

func strRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s},
	}
}

var _ = Describe("HetznerCloudDnsZoneSpec validations", func() {

	Context("with valid specs", func() {
		It("should accept a minimal primary zone with no records", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_primary,
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a primary zone with a single A record set", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_primary,
				RecordSets: []*HetznerCloudDnsZoneSpec_RecordSet{
					{
						Name: "@",
						Type: "A",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("93.184.216.34")},
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a primary zone with multiple record types", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_primary,
				Ttl:        proto.Int32(300),
				RecordSets: []*HetznerCloudDnsZoneSpec_RecordSet{
					{
						Name: "@",
						Type: "A",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("93.184.216.34")},
							{Value: strRef("93.184.216.35")},
						},
					},
					{
						Name: "www",
						Type: "CNAME",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("example.com.")},
						},
					},
					{
						Name: "@",
						Type: "MX",
						Ttl:  proto.Int32(3600),
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("10 mail.example.com.")},
							{Value: strRef("20 mail2.example.com.")},
						},
					},
					{
						Name: "@",
						Type: "TXT",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("\"v=spf1 include:_spf.google.com ~all\"")},
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a primary zone with record comments", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_primary,
				RecordSets: []*HetznerCloudDnsZoneSpec_RecordSet{
					{
						Name: "app",
						Type: "A",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("10.0.0.1"), Comment: "primary server"},
							{Value: strRef("10.0.0.2"), Comment: "standby server"},
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a primary zone with wildcard record", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_primary,
				RecordSets: []*HetznerCloudDnsZoneSpec_RecordSet{
					{
						Name: "*",
						Type: "A",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("93.184.216.34")},
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a primary zone with per-rrset TTL override", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_primary,
				Ttl:        proto.Int32(3600),
				RecordSets: []*HetznerCloudDnsZoneSpec_RecordSet{
					{
						Name: "@",
						Type: "A",
						Ttl:  proto.Int32(60),
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("93.184.216.34")},
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a primary zone with delete protection", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName:       "example.com",
				Mode:             HetznerCloudDnsZoneSpec_primary,
				DeleteProtection: true,
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a secondary zone with primary nameservers", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_secondary,
				PrimaryNameservers: []*HetznerCloudDnsZoneSpec_PrimaryNameserver{
					{Address: "203.0.113.53"},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a secondary zone with TSIG authentication", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_secondary,
				PrimaryNameservers: []*HetznerCloudDnsZoneSpec_PrimaryNameserver{
					{
						Address:       "203.0.113.53",
						Port:          proto.Int32(5353),
						TsigAlgorithm: "hmac-sha256",
						TsigKey:       "supersecretkey==",
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a secondary zone with multiple primary nameservers", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_secondary,
				PrimaryNameservers: []*HetznerCloudDnsZoneSpec_PrimaryNameserver{
					{Address: "203.0.113.53"},
					{Address: "203.0.113.54", Port: proto.Int32(5353)},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})

		It("should accept a fully populated primary zone spec", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName:       "example.com",
				Mode:             HetznerCloudDnsZoneSpec_primary,
				Ttl:              proto.Int32(300),
				DeleteProtection: true,
				RecordSets: []*HetznerCloudDnsZoneSpec_RecordSet{
					{
						Name: "@",
						Type: "A",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("93.184.216.34"), Comment: "web server"},
						},
					},
					{
						Name: "@",
						Type: "AAAA",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("2606:2800:220:1:248:1893:25c8:1946")},
						},
					},
					{
						Name: "www",
						Type: "CNAME",
						Ttl:  proto.Int32(3600),
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("example.com.")},
						},
					},
					{
						Name: "@",
						Type: "MX",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("10 mail.example.com.")},
							{Value: strRef("20 mail2.example.com.")},
						},
					},
					{
						Name: "@",
						Type: "TXT",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("\"v=spf1 include:_spf.google.com ~all\"")},
						},
					},
					{
						Name: "_dmarc",
						Type: "TXT",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("\"v=DMARC1; p=reject; rua=mailto:dmarc@example.com\"")},
						},
					},
					{
						Name: "@",
						Type: "CAA",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("0 issue \"letsencrypt.org\"")},
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).To(BeNil())
		})
	})

	Context("with invalid specs", func() {
		It("should reject an empty domain_name", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "",
				Mode:       HetznerCloudDnsZoneSpec_primary,
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject unspecified mode", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_mode_unspecified,
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a primary zone with primary_nameservers", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_primary,
				PrimaryNameservers: []*HetznerCloudDnsZoneSpec_PrimaryNameserver{
					{Address: "203.0.113.53"},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a secondary zone without primary_nameservers", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_secondary,
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a secondary zone with record_sets", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_secondary,
				PrimaryNameservers: []*HetznerCloudDnsZoneSpec_PrimaryNameserver{
					{Address: "203.0.113.53"},
				},
				RecordSets: []*HetznerCloudDnsZoneSpec_RecordSet{
					{
						Name: "@",
						Type: "A",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("1.2.3.4")},
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a record set with empty name", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_primary,
				RecordSets: []*HetznerCloudDnsZoneSpec_RecordSet{
					{
						Name: "",
						Type: "A",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("1.2.3.4")},
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a record set with empty type", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_primary,
				RecordSets: []*HetznerCloudDnsZoneSpec_RecordSet{
					{
						Name: "@",
						Type: "",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Value: strRef("1.2.3.4")},
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a record set with no records", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_primary,
				RecordSets: []*HetznerCloudDnsZoneSpec_RecordSet{
					{
						Name:    "@",
						Type:    "A",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a record value without value", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_primary,
				RecordSets: []*HetznerCloudDnsZoneSpec_RecordSet{
					{
						Name: "@",
						Type: "A",
						Records: []*HetznerCloudDnsZoneSpec_RecordValue{
							{Comment: "missing value"},
						},
					},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})

		It("should reject a primary nameserver with empty address", func() {
			spec := &HetznerCloudDnsZoneSpec{
				DomainName: "example.com",
				Mode:       HetznerCloudDnsZoneSpec_secondary,
				PrimaryNameservers: []*HetznerCloudDnsZoneSpec_PrimaryNameserver{
					{Address: ""},
				},
			}
			Expect(protovalidate.Validate(spec)).ToNot(BeNil())
		})
	})
})
