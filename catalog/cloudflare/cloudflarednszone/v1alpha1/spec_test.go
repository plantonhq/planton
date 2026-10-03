package cloudflarednszonev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareDnsZoneSpec Custom Validation Tests")
}

// literal is a convenience for building a literal record content value.
func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v}}
}

func zone(name string, spec *CloudflareDnsZoneSpec) *CloudflareDnsZone {
	return &CloudflareDnsZone{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareDnsZone",
		Metadata:   &shared.CatalogObjectMetadata{Name: name},
		Spec:       spec,
	}
}

var _ = ginkgo.Describe("CloudflareDnsZoneSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("zone core", func() {

			ginkgo.It("accepts minimal valid fields", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "test-account-123",
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a partial zone with vanity nameservers", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "test-account-123",
					Type:              CloudflareDnsZoneSpec_partial,
					VanityNameServers: []string{"ns1.example.com", "ns2.example.com"},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a paused subdomain zone", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "sub.domain.example.com", AccountId: "test-account-123", Paused: true,
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts embedded records (content)", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "test-account-123",
					Records: []*CloudflareDnsZoneRecord{
						{Name: "www", Type: CloudflareDnsZoneRecord_A, Content: literal("192.0.2.1"), Proxied: true},
						{Name: "@", Type: CloudflareDnsZoneRecord_MX, Content: literal("mail.example.com"), Priority: 10},
					},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts structured records (SRV and CAA data blocks)", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "test-account-123",
					Records: []*CloudflareDnsZoneRecord{
						{Name: "_sip._tcp", Type: CloudflareDnsZoneRecord_SRV,
							Data: &CloudflareDnsZoneRecord_Srv{Srv: &SrvData{Priority: 10, Weight: 5, Port: 5060, Target: literal("sip.example.com")}}},
						{Name: "@", Type: CloudflareDnsZoneRecord_CAA,
							Data: &CloudflareDnsZoneRecord_Caa{Caa: &CaaData{Tag: "issue", Value: "letsencrypt.org"}}},
					},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a record with tags and settings", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "test-account-123",
					Records: []*CloudflareDnsZoneRecord{
						{Name: "app", Type: CloudflareDnsZoneRecord_A, Content: literal("192.0.2.7"), Proxied: true,
							Tags:     []string{"team:web", "env:prod"},
							Settings: &CloudflareDnsZoneRecordSettings{Ipv4Only: true}},
					},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("hold and subscription", func() {

			ginkgo.It("accepts an enabled hold with subdomains and hold_after", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "test-account-123",
					Hold: &CloudflareDnsZoneHold{Enabled: true, IncludeSubdomains: true, HoldAfter: "2026-01-31T00:00:00Z"},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts a subscription with rate_plan and frequency", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "test-account-123",
					Subscription: &CloudflareDnsZoneSubscription{RatePlan: "pro", Frequency: "monthly"},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("folded dns_settings and dnssec", func() {

			ginkgo.It("accepts dns_settings with zone_mode, ns_ttl, and soa", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "test-account-123",
					DnsSettings: &CloudflareDnsZoneDnsSettings{
						FlattenAllCnames: true,
						ZoneMode:         CloudflareDnsZoneSpec_standard,
						NsTtl:            3600,
						Soa:              &CloudflareDnsZoneSoa{Refresh: 10000, Retry: 2400, Expire: 604800, MinTtl: 1800, Ttl: 3600},
						Nameservers:      &CloudflareDnsZoneNameservers{NsSet: 1, Type: "cloudflare.standard"},
					},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("accepts an enabled dnssec block", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "test-account-123",
					Dnssec: &CloudflareDnsZoneDnssec{Enabled: true, UseNsec3: true},
				}))
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("zone core", func() {

			ginkgo.It("rejects a missing zone_name", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{AccountId: "test-account-123"}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a missing account_id", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{ZoneName: "example.com"}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects an invalid zone_name (no TLD)", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{ZoneName: "invalidzone", AccountId: "a"}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects an uppercase zone_name", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{ZoneName: "EXAMPLE.COM", AccountId: "a"}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("dns_settings constraints", func() {

			ginkgo.It("rejects an ns_ttl below the floor", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					DnsSettings: &CloudflareDnsZoneDnsSettings{NsTtl: 10},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a soa.refresh out of range", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					DnsSettings: &CloudflareDnsZoneDnsSettings{Soa: &CloudflareDnsZoneSoa{Refresh: 100}},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects an ns_set out of range", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					DnsSettings: &CloudflareDnsZoneDnsSettings{Nameservers: &CloudflareDnsZoneNameservers{NsSet: 9}},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects an invalid nameserver type", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					DnsSettings: &CloudflareDnsZoneDnsSettings{Nameservers: &CloudflareDnsZoneNameservers{Type: "bogus"}},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("embedded record constraints", func() {

			ginkgo.It("rejects a proxied TXT record", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					Records: []*CloudflareDnsZoneRecord{{Name: "@", Type: CloudflareDnsZoneRecord_TXT, Content: literal("v=spf1 ~all"), Proxied: true}},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects an MX record without priority", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					Records: []*CloudflareDnsZoneRecord{{Name: "@", Type: CloudflareDnsZoneRecord_MX, Content: literal("mail.example.com")}},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a record setting both content and a data block", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					Records: []*CloudflareDnsZoneRecord{
						{Name: "_sip._tcp", Type: CloudflareDnsZoneRecord_SRV, Content: literal("bogus"),
							Data: &CloudflareDnsZoneRecord_Srv{Srv: &SrvData{Target: literal("sip.example.com")}}},
					},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a structured type without its data block", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					Records: []*CloudflareDnsZoneRecord{{Name: "_sip._tcp", Type: CloudflareDnsZoneRecord_SRV, Content: literal("10 5 5060 sip.example.com")}},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a data block that does not match the record type", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					Records: []*CloudflareDnsZoneRecord{
						{Name: "@", Type: CloudflareDnsZoneRecord_SRV,
							Data: &CloudflareDnsZoneRecord_Caa{Caa: &CaaData{Tag: "issue", Value: "letsencrypt.org"}}},
					},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects a CAA data block missing its required tag", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					Records: []*CloudflareDnsZoneRecord{
						{Name: "@", Type: CloudflareDnsZoneRecord_CAA,
							Data: &CloudflareDnsZoneRecord_Caa{Caa: &CaaData{Value: "letsencrypt.org"}}},
					},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("hold and subscription constraints", func() {

			ginkgo.It("rejects a malformed hold_after timestamp", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					Hold: &CloudflareDnsZoneHold{Enabled: true, HoldAfter: "next tuesday"},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects an unknown rate_plan", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					Subscription: &CloudflareDnsZoneSubscription{RatePlan: "platinum"},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects an invalid frequency", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					Subscription: &CloudflareDnsZoneSubscription{RatePlan: "pro", Frequency: "daily"},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects an empty subscription block", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					Subscription: &CloudflareDnsZoneSubscription{},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
