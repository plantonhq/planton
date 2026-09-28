package azurednsrecordv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureDnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDnsRecordSpec Validation Tests")
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

func int32Ptr(v int32) *int32 { return &v }

// validResource returns a minimal valid AzureDnsRecord (an A record) that
// individual cases then mutate into the shape under test.
func validResource() *AzureDnsRecord {
	return &AzureDnsRecord{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDnsRecord",
		Metadata: &shared.CloudResourceMetadata{
			Name: "www-a-record",
		},
		Spec: &AzureDnsRecordSpec{
			ResourceGroup: literal("my-rg"),
			ZoneName:      ref("my-zone"),
			Name:          "www",
			A: &AzureDnsARecord{
				Addresses: []*foreignkeyv1.StringValueOrRef{literal("203.0.113.10")},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureDnsRecordSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_dns_record", func() {

			ginkgo.It("should not return a validation error for a minimal A record", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an A record whose address is another resource's output", func() {
				input := validResource()
				input.Spec.A.Addresses = []*foreignkeyv1.StringValueOrRef{ref("web-public-ip")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an NS delegation whose name servers are another zone's outputs", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Ns = []*foreignkeyv1.StringValueOrRef{ref("team-zone"), literal("ns2-01.azure-dns.net")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a multi-address A record with a TTL and tags", func() {
				input := validResource()
				input.Spec.A.Addresses = []*foreignkeyv1.StringValueOrRef{literal("203.0.113.10"), literal("203.0.113.11")}
				ttl := int32(60)
				input.Spec.TtlSeconds = &ttl
				input.Spec.Tags = map[string]string{"team": "web"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an alias A record at the zone apex", func() {
				input := validResource()
				input.Spec.Name = "@"
				input.Spec.A = &AzureDnsARecord{
					TargetResourceId: ref("frontend-public-ip"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an AAAA record", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Aaaa = &AzureDnsAaaaRecord{
					Addresses: []*foreignkeyv1.StringValueOrRef{literal("2001:db8::1")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a CNAME record with a literal target", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Cname = &AzureDnsCnameRecord{Value: literal("myapp.azurefd.net")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a CNAME record whose target is a reference", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Cname = &AzureDnsCnameRecord{Value: ref("frontdoor-endpoint")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an alias CNAME record", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Cname = &AzureDnsCnameRecord{
					TargetResourceId: literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg/providers/Microsoft.Cdn/profiles/p/afdEndpoints/e"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an MX record set with multiple preferences", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "@"
				input.Spec.Mx = []*AzureDnsMxEntry{
					{Preference: int32Ptr(10), Exchange: literal("mail1.example.com")},
					{Preference: int32Ptr(20), Exchange: literal("mail2.example.com")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a null-MX entry with preference 0", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "@"
				input.Spec.Mx = []*AzureDnsMxEntry{
					{Preference: int32Ptr(0), Exchange: literal(".")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an SRV record on an underscore service name", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "_sip._tcp"
				input.Spec.Srv = []*AzureDnsSrvEntry{
					{Priority: int32Ptr(0), Weight: int32Ptr(5), Port: int32Ptr(5060), Target: literal("sip.example.com")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a CAA record set", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "@"
				input.Spec.Caa = []*AzureDnsCaaEntry{
					{Flags: int32Ptr(0), Tag: AzureDnsCaaTag_ISSUE, Value: "letsencrypt.org"},
					{Flags: int32Ptr(0), Tag: AzureDnsCaaTag_IODEF, Value: "mailto:security@example.com"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a TXT record with a long value", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "_dmarc"
				longValue := make([]byte, 1000)
				for i := range longValue {
					longValue[i] = 'a'
				}
				input.Spec.Txt = []*foreignkeyv1.StringValueOrRef{
					literal("v=DMARC1; p=reject;"),
					literal(string(longValue)),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a TXT record mixing literal and referenced values", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "_dnsauth"
				input.Spec.Txt = []*foreignkeyv1.StringValueOrRef{
					ref("frontdoor-custom-domain"),
					literal("v=spf1 -all"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an NS delegation record", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "team"
				input.Spec.Ns = []*foreignkeyv1.StringValueOrRef{literal("ns1-01.azure-dns.com."), literal("ns2-01.azure-dns.net.")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a PTR record", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "10"
				input.Spec.Ptr = []*foreignkeyv1.StringValueOrRef{literal("host.example.com")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a wildcard record name", func() {
				input := validResource()
				input.Spec.Name = "*.app"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_dns_record", func() {

			ginkgo.It("should return a validation error for a literal A address that is not IPv4", func() {
				input := validResource()
				input.Spec.A.Addresses = []*foreignkeyv1.StringValueOrRef{literal("2001:db8::1")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("each literal address is an IPv4 address")))
			})

			ginkgo.It("should return a validation error for a literal AAAA address that is not IPv6", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Aaaa = &AzureDnsAaaaRecord{Addresses: []*foreignkeyv1.StringValueOrRef{literal("203.0.113.10")}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring("each literal address is an IPv6 address")))
			})

			ginkgo.It("should return a validation error when no payload is set", func() {
				input := validResource()
				input.Spec.A = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when two payloads are set", func() {
				input := validResource()
				input.Spec.Cname = &AzureDnsCnameRecord{Value: literal("target.example.com")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an A record with neither addresses nor alias", func() {
				input := validResource()
				input.Spec.A = &AzureDnsARecord{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an A record with both addresses and alias", func() {
				input := validResource()
				input.Spec.A = &AzureDnsARecord{
					Addresses:        []*foreignkeyv1.StringValueOrRef{literal("203.0.113.10")},
					TargetResourceId: ref("frontend-public-ip"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an A record with an invalid IPv4 address", func() {
				input := validResource()
				input.Spec.A.Addresses = []*foreignkeyv1.StringValueOrRef{literal("999.0.113.10")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an AAAA record with an IPv4 address", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Aaaa = &AzureDnsAaaaRecord{Addresses: []*foreignkeyv1.StringValueOrRef{literal("203.0.113.10")}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a CNAME with both value and alias", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Cname = &AzureDnsCnameRecord{
					Value:            literal("target.example.com"),
					TargetResourceId: ref("cdn-endpoint"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an empty CNAME payload", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Cname = &AzureDnsCnameRecord{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an MX entry without a preference", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Mx = []*AzureDnsMxEntry{{Exchange: literal("mail.example.com")}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an MX entry without an exchange", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Mx = []*AzureDnsMxEntry{{Preference: int32Ptr(10)}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an out-of-range MX preference", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Mx = []*AzureDnsMxEntry{{Preference: int32Ptr(70000), Exchange: literal("mail.example.com")}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an SRV entry missing its port", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "_sip._tcp"
				input.Spec.Srv = []*AzureDnsSrvEntry{
					{Priority: int32Ptr(0), Weight: int32Ptr(5), Target: literal("sip.example.com")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an SRV entry with an out-of-range port", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "_sip._tcp"
				input.Spec.Srv = []*AzureDnsSrvEntry{
					{Priority: int32Ptr(0), Weight: int32Ptr(5), Port: int32Ptr(70000), Target: literal("sip.example.com")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a CAA entry with an unspecified tag", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Caa = []*AzureDnsCaaEntry{
					{Flags: int32Ptr(0), Value: "letsencrypt.org"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a CAA entry with out-of-range flags", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Caa = []*AzureDnsCaaEntry{
					{Flags: int32Ptr(300), Tag: AzureDnsCaaTag_ISSUE, Value: "letsencrypt.org"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid record name", func() {
				input := validResource()
				input.Spec.Name = "Not A Record!"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a negative TTL", func() {
				input := validResource()
				ttl := int32(-1)
				input.Spec.TtlSeconds = &ttl
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when zone_name is missing", func() {
				input := validResource()
				input.Spec.ZoneName = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
