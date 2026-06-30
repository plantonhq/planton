package civodnszonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	dnsrecordtype "github.com/plantonhq/planton/apis/dev/planton/shared/networking/enums/dnsrecordtype"
)

func TestCivoDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CivoDnsZoneSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CivoDnsZoneSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("minimal valid zone", func() {
			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-dns-zone",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a subdomain as domain name", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "subdomain-zone",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "sub.example.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a multi-level subdomain", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "deep-subdomain-zone",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "api.staging.example.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("zone with basic DNS records", func() {
			ginkgo.It("should not return a validation error for A records", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "zone-with-a-records",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "@",
								Type: dnsrecordtype.DnsRecordType_A,
								Values: []*foreignkeyv1.StringValueOrRef{
									{
										LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
											Value: "198.51.100.42",
										},
									},
								},
								TtlSeconds: 3600,
							},
							{
								Name: "www",
								Type: dnsrecordtype.DnsRecordType_A,
								Values: []*foreignkeyv1.StringValueOrRef{
									{
										LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
											Value: "198.51.100.42",
										},
									},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should support multiple A record values for round-robin", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "zone-with-multi-a",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "api",
								Type: dnsrecordtype.DnsRecordType_A,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "203.0.113.10"}},
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "203.0.113.11"}},
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "203.0.113.12"}},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for CNAME records", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "zone-with-cname",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "www",
								Type: dnsrecordtype.DnsRecordType_CNAME,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example.com"}},
								},
								TtlSeconds: 3600,
							},
							{
								Name: "cdn",
								Type: dnsrecordtype.DnsRecordType_CNAME,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "cdn-provider.example.net"}},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for MX records", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "zone-with-mx",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "@",
								Type: dnsrecordtype.DnsRecordType_MX,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "10 mail.example.com"}},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should support multiple MX records with different priorities", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "zone-with-multi-mx",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "@",
								Type: dnsrecordtype.DnsRecordType_MX,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "10 mail1.example.com"}},
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "20 mail2.example.com"}},
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "30 mail3.example.com"}},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for TXT records", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "zone-with-txt",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "@",
								Type: dnsrecordtype.DnsRecordType_TXT,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "v=spf1 mx ~all"}},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should support AAAA records for IPv6", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "zone-with-aaaa",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "@",
								Type: dnsrecordtype.DnsRecordType_AAAA,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "2001:db8::1"}},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should support SRV records", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "zone-with-srv",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "_service._tcp",
								Type: dnsrecordtype.DnsRecordType_SRV,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "10 60 5060 bigbox.example.com"}},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("real-world DNS configuration scenarios", func() {
			ginkgo.It("should support basic web hosting setup (A + CNAME)", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "web-hosting",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "@",
								Type: dnsrecordtype.DnsRecordType_A,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "198.51.100.42"}},
								},
								TtlSeconds: 3600,
							},
							{
								Name: "www",
								Type: dnsrecordtype.DnsRecordType_CNAME,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "example.com"}},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should support email setup with SPF and DKIM", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "email-setup",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "@",
								Type: dnsrecordtype.DnsRecordType_MX,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "10 mail.example.com"}},
								},
								TtlSeconds: 3600,
							},
							{
								Name: "mail",
								Type: dnsrecordtype.DnsRecordType_A,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "198.51.100.50"}},
								},
								TtlSeconds: 3600,
							},
							{
								Name: "@",
								Type: dnsrecordtype.DnsRecordType_TXT,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "v=spf1 mx ~all"}},
								},
								TtlSeconds: 3600,
							},
							{
								Name: "google._domainkey",
								Type: dnsrecordtype.DnsRecordType_TXT,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "v=DKIM1; k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC..."}},
								},
								TtlSeconds: 3600,
							},
							{
								Name: "_dmarc",
								Type: dnsrecordtype.DnsRecordType_TXT,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "v=DMARC1; p=quarantine; rua=mailto:postmaster@example.com"}},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should support API with multiple A records for load distribution", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "api-load-balancing",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "api",
								Type: dnsrecordtype.DnsRecordType_A,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "203.0.113.10"}},
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "203.0.113.11"}},
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "203.0.113.12"}},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should support ACME DNS-01 challenge for cert-manager", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "cert-validation",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "_acme-challenge",
								Type: dnsrecordtype.DnsRecordType_TXT,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "ACME_token_for_validation_abc123"}},
								},
								TtlSeconds: 600, // Lower TTL for faster propagation
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("TTL variations", func() {
			ginkgo.It("should accept custom TTL values", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "custom-ttl",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "short-ttl",
								Type: dnsrecordtype.DnsRecordType_A,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "198.51.100.42"}},
								},
								TtlSeconds: 600, // 10 minutes
							},
							{
								Name: "long-ttl",
								Type: dnsrecordtype.DnsRecordType_A,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "198.51.100.43"}},
								},
								TtlSeconds: 86400, // 24 hours
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept zero TTL (will default to 3600)", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "default-ttl",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "default",
								Type: dnsrecordtype.DnsRecordType_A,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "198.51.100.42"}},
								},
								TtlSeconds: 0, // Will default to 3600 in implementation
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("invalid domain names", func() {
			ginkgo.It("should return a validation error for missing domain name", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-domain",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "", // Missing required field
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for invalid domain pattern (no TLD)", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "invalid-domain",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example", // No TLD
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for domain with invalid characters", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "invalid-chars",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "exam!ple@test.com", // Invalid characters
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for domain with spaces", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "spaces-domain",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example .com", // Space in domain
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid record configurations", func() {
			ginkgo.It("should return a validation error for missing record name", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "missing-record-name",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "", // Missing required field
								Type: dnsrecordtype.DnsRecordType_A,
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "198.51.100.42"}},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing record type", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "missing-record-type",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name: "www",
								// Type missing (required)
								Values: []*foreignkeyv1.StringValueOrRef{
									{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "198.51.100.42"}},
								},
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing record values", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "missing-values",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name:       "www",
								Type:       dnsrecordtype.DnsRecordType_A,
								Values:     []*foreignkeyv1.StringValueOrRef{}, // Empty, but min_items=1
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for nil values array", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "nil-values",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
						Records: []*CivoDnsZoneRecord{
							{
								Name:       "www",
								Type:       dnsrecordtype.DnsRecordType_A,
								Values:     nil, // Nil values
								TtlSeconds: 3600,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid metadata", func() {
			ginkgo.It("should return a validation error for missing metadata", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata:   nil, // Missing required metadata
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing spec", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-spec",
					},
					Spec: nil, // Missing required spec
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for wrong api_version", func() {
				input := &CivoDnsZone{
					ApiVersion: "wrong.api.version/v1", // Wrong value
					Kind:       "CivoDnsZone",
					Metadata: &shared.CloudResourceMetadata{
						Name: "wrong-api",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for wrong kind", func() {
				input := &CivoDnsZone{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "WrongKind", // Wrong kind
					Metadata: &shared.CloudResourceMetadata{
						Name: "wrong-kind",
					},
					Spec: &CivoDnsZoneSpec{
						DomainName: "example.com",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
