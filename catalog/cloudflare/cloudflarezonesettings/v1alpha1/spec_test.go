package cloudflarezonesettingsv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareZoneSettingsSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZoneSettingsSpec Custom Validation Tests")
}

func boolPtr(b bool) *bool { return &b }

func strPtr(s string) *string { return &s }

func int64Ptr(i int64) *int64 { return &i }

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "023e105f4ecef8ad9ca31a8372d0c353"}}
}

func validZoneSettings(spec *CloudflareZoneSettingsSpec) *CloudflareZoneSettings {
	return &CloudflareZoneSettings{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZoneSettings",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-zone-settings",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareZoneSettingsSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal spec managing one toggle", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId:         zoneRef(),
				AlwaysUseHttps: boolPtr(true),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept every enum-walled setting at a legal value", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId:               zoneRef(),
				CacheLevel:           strPtr("aggressive"),
				CnameFlattening:      strPtr("flatten_at_root"),
				H2Prioritization:     strPtr("custom"),
				ImageResizing:        strPtr("open"),
				MinTlsVersion:        strPtr("1.2"),
				OriginMaxHttpVersion: strPtr("2"),
				Polish:               strPtr("lossless"),
				PseudoIpv4:           strPtr("add_header"),
				SecurityLevel:        strPtr("under_attack"),
				Ssl:                  strPtr("strict"),
				Tls_1_3:              strPtr("zrt"),
				Transformations:      strPtr("on"),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a spec managing only the ciphers list", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId:  zoneRef(),
				Ciphers: []string{"ECDHE-ECDSA-AES128-GCM-SHA256"},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a full security_header block", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId: zoneRef(),
				SecurityHeader: &CloudflareZoneSettingsSecurityHeader{
					Enabled:           boolPtr(true),
					IncludeSubdomains: true,
					MaxAge:            31536000,
					Nosniff:           true,
					Preload:           false,
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a complete automatic_platform_optimization block", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId: zoneRef(),
				AutomaticPlatformOptimization: &CloudflareZoneSettingsAutomaticPlatformOptimization{
					Enabled:           boolPtr(true),
					CacheByDeviceType: true,
					Cf:                true,
					Hostnames:         []string{"example.com", "www.example.com"},
					Wordpress:         true,
					WpPlugin:          true,
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept managed switches stated off (nel, security_header, automatic_platform_optimization)", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId: zoneRef(),
				Nel:    &CloudflareZoneSettingsNel{Enabled: boolPtr(false)},
				SecurityHeader: &CloudflareZoneSettingsSecurityHeader{
					Enabled: boolPtr(false),
				},
				AutomaticPlatformOptimization: &CloudflareZoneSettingsAutomaticPlatformOptimization{
					Enabled:   boolPtr(false),
					Hostnames: []string{"example.com"},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept managed headers and url normalization together", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId: zoneRef(),
				ManagedRequestHeaders: []*CloudflareZoneSettingsManagedHeader{
					{Id: "add_true_client_ip_headers", Enabled: true},
				},
				ManagedResponseHeaders: []*CloudflareZoneSettingsManagedHeader{
					{Id: "remove_x-powered-by_header", Enabled: true},
				},
				UrlNormalization: &CloudflareZoneSettingsUrlNormalization{
					Scope: "incoming",
					Type:  "cloudflare",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an origin cloud region row with a legal vendor", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId: zoneRef(),
				OriginCloudRegions: []*CloudflareZoneSettingsOriginCloudRegion{
					{OriginIp: "203.0.113.10", Region: "us-east-1", Vendor: "aws"},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a spec without zone_id", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				AlwaysUseHttps: boolPtr(true),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a spec that manages no settings at all", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId: zoneRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an illegal ssl mode", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId: zoneRef(),
				Ssl:    strPtr("very-strict"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an illegal min_tls_version", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId:        zoneRef(),
				MinTlsVersion: strPtr("1.4"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an illegal security_level", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId:        zoneRef(),
				SecurityLevel: strPtr("paranoid"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a negative browser_cache_ttl", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId:          zoneRef(),
				BrowserCacheTtl: int64Ptr(-1),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a managed-switch block that does not say on or off", func() {
			for _, spec := range []*CloudflareZoneSettingsSpec{
				{ZoneId: zoneRef(), Nel: &CloudflareZoneSettingsNel{}},
				{ZoneId: zoneRef(), SecurityHeader: &CloudflareZoneSettingsSecurityHeader{MaxAge: 31536000}},
				{ZoneId: zoneRef(), AutomaticPlatformOptimization: &CloudflareZoneSettingsAutomaticPlatformOptimization{Hostnames: []string{"example.com"}}},
			} {
				err := protovalidate.Validate(validZoneSettings(spec))
				gomega.Expect(err).ToNot(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("enabled"))
			}
		})

		ginkgo.It("should reject a security_header with negative max_age", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId: zoneRef(),
				SecurityHeader: &CloudflareZoneSettingsSecurityHeader{
					Enabled: boolPtr(true),
					MaxAge:  -5,
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject url_normalization with an illegal scope", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId: zoneRef(),
				UrlNormalization: &CloudflareZoneSettingsUrlNormalization{
					Scope: "outgoing",
					Type:  "cloudflare",
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an origin cloud region row with an unknown vendor", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId: zoneRef(),
				OriginCloudRegions: []*CloudflareZoneSettingsOriginCloudRegion{
					{OriginIp: "203.0.113.10", Region: "us-east1", Vendor: "digitalocean"},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a managed header row without an id", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId: zoneRef(),
				ManagedRequestHeaders: []*CloudflareZoneSettingsManagedHeader{
					{Enabled: true},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject automatic_platform_optimization without hostnames", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId: zoneRef(),
				AutomaticPlatformOptimization: &CloudflareZoneSettingsAutomaticPlatformOptimization{
					Enabled:   boolPtr(true),
					Wordpress: true,
					WpPlugin:  true,
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an illegal tls_1_3 value", func() {
			input := validZoneSettings(&CloudflareZoneSettingsSpec{
				ZoneId:  zoneRef(),
				Tls_1_3: strPtr("enabled"),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
