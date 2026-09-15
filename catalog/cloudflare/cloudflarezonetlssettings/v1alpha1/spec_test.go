package cloudflarezonetlssettingsv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareZoneTlsSettingsSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZoneTlsSettingsSpec Custom Validation Tests")
}

func boolPtr(b bool) *bool { return &b }

func strPtr(s string) *string { return &s }

func zoneRef() *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "023e105f4ecef8ad9ca31a8372d0c353"}}
}

func validTlsSettings(spec *CloudflareZoneTlsSettingsSpec) *CloudflareZoneTlsSettings {
	return &CloudflareZoneTlsSettings{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZoneTlsSettings",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-zone-tls-settings",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareZoneTlsSettingsSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal spec managing universal ssl", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId:              zoneRef(),
				UniversalSslEnabled: boolPtr(true),
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept total_tls with a legal certificate authority", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId: zoneRef(),
				TotalTls: &CloudflareZoneTlsSettingsTotalTls{
					Enabled:              boolPtr(true),
					CertificateAuthority: strPtr("google"),
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept total_tls switched off", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId:   zoneRef(),
				TotalTls: &CloudflareZoneTlsSettingsTotalTls{Enabled: boolPtr(false)},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a hostname row overriding only min_tls_version", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId: zoneRef(),
				HostnameSettings: []*CloudflareZoneTlsSettingsHostnameSetting{
					{Hostname: "api.example.com", MinTlsVersion: strPtr("1.3")},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a CA hostname association with an mTLS certificate reference", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId: zoneRef(),
				CaHostnameAssociations: []*CloudflareZoneTlsSettingsCaHostnameAssociation{
					{
						Hostnames: []string{"mtls.example.com"},
						MtlsCertificateId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "8d773bb3-90b2-4bd3-9d33-9c3ba33d5b9a"},
						},
					},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an open-vocabulary compliance modes list", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId:                   zoneRef(),
				OriginTlsComplianceModes: []string{"fips", "pqh"},
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a spec without zone_id", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				UniversalSslEnabled: boolPtr(true),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a spec that manages no settings at all", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId: zoneRef(),
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject total_tls that does not say on or off", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId:   zoneRef(),
				TotalTls: &CloudflareZoneTlsSettingsTotalTls{CertificateAuthority: strPtr("google")},
			})
			err := protovalidate.Validate(input)
			gomega.Expect(err).ToNot(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("total_tls.enabled"))
		})

		ginkgo.It("should reject total_tls with an unknown certificate authority", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId: zoneRef(),
				TotalTls: &CloudflareZoneTlsSettingsTotalTls{
					Enabled:              boolPtr(true),
					CertificateAuthority: strPtr("digicert"),
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a hostname row that overrides nothing", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId: zoneRef(),
				HostnameSettings: []*CloudflareZoneTlsSettingsHostnameSetting{
					{Hostname: "api.example.com"},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a hostname row without a hostname", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId: zoneRef(),
				HostnameSettings: []*CloudflareZoneTlsSettingsHostnameSetting{
					{MinTlsVersion: strPtr("1.2")},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an illegal min_tls_version on a hostname row", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId: zoneRef(),
				HostnameSettings: []*CloudflareZoneTlsSettingsHostnameSetting{
					{Hostname: "api.example.com", MinTlsVersion: strPtr("1.5")},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a CA hostname association with no hostnames", func() {
			input := validTlsSettings(&CloudflareZoneTlsSettingsSpec{
				ZoneId: zoneRef(),
				CaHostnameAssociations: []*CloudflareZoneTlsSettingsCaHostnameAssociation{
					{Hostnames: []string{}},
				},
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
