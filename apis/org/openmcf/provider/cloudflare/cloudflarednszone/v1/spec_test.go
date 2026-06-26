package cloudflarednszonev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestCloudflareDnsZoneSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareDnsZoneSpec Custom Validation Tests")
}

func zone(name string, spec *CloudflareDnsZoneSpec) *CloudflareDnsZone {
	return &CloudflareDnsZone{
		ApiVersion: "cloudflare.openmcf.org/v1",
		Kind:       "CloudflareDnsZone",
		Metadata:   &shared.CloudResourceMetadata{Name: name},
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
						{Name: "www", Type: CloudflareDnsZoneRecord_A, Content: "192.0.2.1", Proxied: true},
						{Name: "@", Type: CloudflareDnsZoneRecord_MX, Content: "mail.example.com", Priority: 10},
					},
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
					Records: []*CloudflareDnsZoneRecord{{Name: "@", Type: CloudflareDnsZoneRecord_TXT, Content: "v=spf1 ~all", Proxied: true}},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("rejects an MX record without priority", func() {
				err := protovalidate.Validate(zone("z", &CloudflareDnsZoneSpec{
					ZoneName: "example.com", AccountId: "a",
					Records: []*CloudflareDnsZoneRecord{{Name: "@", Type: CloudflareDnsZoneRecord_MX, Content: "mail.example.com"}},
				}))
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
