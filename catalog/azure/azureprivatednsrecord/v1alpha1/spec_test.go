package azureprivatednsrecordv1alpha1

import (
	"strconv"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzurePrivateDnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePrivateDnsRecordSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

// hostAddresses returns n distinct syntactically valid IPv4 addresses.
func hostAddresses(n int) []*foreignkeyv1.StringValueOrRef {
	addresses := make([]*foreignkeyv1.StringValueOrRef, 0, n)
	for i := 1; i <= n; i++ {
		addresses = append(addresses, literal("10.0.0."+strconv.Itoa(i)))
	}
	return addresses
}

const testZoneId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/privateDnsZones/internal.contoso.com"

// validResource returns a minimal valid A record that individual cases
// mutate into the shape under test.
func validResource() *AzurePrivateDnsRecord {
	return &AzurePrivateDnsRecord{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePrivateDnsRecord",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-record",
		},
		Spec: &AzurePrivateDnsRecordSpec{
			PrivateDnsZoneId: literal(testZoneId),
			Name:             "db",
			A:                []*foreignkeyv1.StringValueOrRef{literal("10.0.0.5")},
		},
	}
}

var _ = ginkgo.Describe("AzurePrivateDnsRecordSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_private_dns_record", func() {

			ginkgo.It("should not return a validation error for the minimal A record", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an A record at Azure's 20-address cap with explicit ttl and tags", func() {
				input := validResource()
				input.Spec.A = hostAddresses(20)
				input.Spec.TtlSeconds = int32Ptr(3600)
				input.Spec.Tags = map[string]string{"owner": "platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an AAAA record with compressed IPv6 addresses", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Aaaa = []*foreignkeyv1.StringValueOrRef{literal("2001:db8::1"), literal("2001:db8::2")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a CNAME record with a literal target", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Cname = literal("db.internal.contoso.com")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an apex MX record with multiple entries", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "@"
				input.Spec.Mx = []*AzurePrivateDnsMxEntry{
					{Preference: int32Ptr(10), Exchange: literal("mail1.internal.contoso.com")},
					{Preference: int32Ptr(20), Exchange: literal("mail2.internal.contoso.com")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a PTR record in a reverse zone", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "5"
				input.Spec.Ptr = []*foreignkeyv1.StringValueOrRef{literal("db.internal.contoso.com")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an SRV record with the underscore-led service name", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Name = "_sip._tcp"
				input.Spec.Srv = []*AzurePrivateDnsSrvEntry{
					{Priority: int32Ptr(0), Weight: int32Ptr(5), Port: int32Ptr(5060), Target: literal("sip.internal.contoso.com")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a TXT record mixing literal and reference values", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Txt = []*foreignkeyv1.StringValueOrRef{
					literal("v=spf1 -all"),
					literal("service-discovery-token"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept wildcard and dotted record names", func() {
				for _, name := range []string{"*", "*.app", "api.v1", "_dmarc"} {
					input := validResource()
					input.Spec.Name = name
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil(), "name %q should be valid", name)
				}
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_private_dns_record", func() {

			ginkgo.It("should return a validation error when no payload is set", func() {
				input := validResource()
				input.Spec.A = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "exactly one record payload")).To(gomega.BeTrue())
			})

			ginkgo.It("should return a validation error when two payloads are set", func() {
				input := validResource()
				input.Spec.Txt = []*foreignkeyv1.StringValueOrRef{literal("v=spf1 -all")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "exactly one record payload")).To(gomega.BeTrue())
			})

			ginkgo.It("should return a validation error for an A record beyond the 20-address cap", func() {
				input := validResource()
				input.Spec.A = hostAddresses(21)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed IPv4 address", func() {
				input := validResource()
				input.Spec.A = []*foreignkeyv1.StringValueOrRef{literal("10.0.0.999")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a malformed IPv6 address", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Aaaa = []*foreignkeyv1.StringValueOrRef{literal("not-an-ip")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an uppercase record name", func() {
				input := validResource()
				input.Spec.Name = "DB"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
				gomega.Expect(strings.Contains(err.Error(), "Record name")).To(gomega.BeTrue())
			})

			ginkgo.It("should return a validation error for a hyphen-led record name", func() {
				input := validResource()
				input.Spec.Name = "-bad"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an MX entry missing its exchange", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Mx = []*AzurePrivateDnsMxEntry{{Preference: int32Ptr(10)}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an MX preference beyond the 16-bit range", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Mx = []*AzurePrivateDnsMxEntry{{Preference: int32Ptr(65536), Exchange: literal("mail.internal.contoso.com")}}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an SRV entry with port 0", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Srv = []*AzurePrivateDnsSrvEntry{
					{Priority: int32Ptr(0), Weight: int32Ptr(0), Port: int32Ptr(0), Target: literal("sip.internal.contoso.com")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an SRV entry missing its target", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Srv = []*AzurePrivateDnsSrvEntry{
					{Priority: int32Ptr(0), Weight: int32Ptr(0), Port: int32Ptr(5060)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an empty PTR hostname", func() {
				input := validResource()
				input.Spec.A = nil
				input.Spec.Ptr = []*foreignkeyv1.StringValueOrRef{literal("")}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the zone reference is missing", func() {
				input := validResource()
				input.Spec.PrivateDnsZoneId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a negative TTL", func() {
				input := validResource()
				input.Spec.TtlSeconds = int32Ptr(-1)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
