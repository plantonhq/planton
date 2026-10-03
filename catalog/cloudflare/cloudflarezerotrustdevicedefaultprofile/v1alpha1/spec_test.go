package cloudflarezerotrustdevicedefaultprofilev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareZeroTrustDeviceDefaultProfileSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustDeviceDefaultProfileSpec Custom Validation Tests")
}

const (
	testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"
	testVnetID    = "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415"
	testZoneID    = "023e105f4ecef8ad9ca31a8372d0c353"
)

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func validProfile(spec *CloudflareZeroTrustDeviceDefaultProfileSpec) *CloudflareZeroTrustDeviceDefaultProfile {
	return &CloudflareZeroTrustDeviceDefaultProfile{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustDeviceDefaultProfile",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-default-profile",
		},
		Spec: spec,
	}
}

func baseSpec() *CloudflareZeroTrustDeviceDefaultProfileSpec {
	return &CloudflareZeroTrustDeviceDefaultProfileSpec{
		AccountId: testAccountID,
	}
}

var _ = ginkgo.Describe("CloudflareZeroTrustDeviceDefaultProfileSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal profile (account only)", func() {
			gomega.Expect(protovalidate.Validate(validProfile(baseSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept the full toggle body", func() {
			spec := baseSpec()
			spec.AllowModeSwitch = proto.Bool(true)
			spec.AllowedToLeave = proto.Bool(false)
			spec.AutoConnect = proto.Int64(600)
			spec.CaptivePortal = proto.Int64(180)
			spec.SwitchLocked = proto.Bool(true)
			spec.TunnelProtocol = "masque"
			spec.LanAllowMinutes = proto.Int64(30)
			spec.LanAllowSubnetSize = proto.Int64(24)
			gomega.Expect(protovalidate.Validate(validProfile(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept an exclude-mode split tunnel", func() {
			spec := baseSpec()
			spec.Exclude = []*CloudflareZeroTrustDeviceDefaultProfileSplitTunnelEntry{
				{Address: "192.0.2.0/24", Description: "lab"},
				{Host: "internal.example.com"},
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept an include-mode split tunnel", func() {
			spec := baseSpec()
			spec.Include = []*CloudflareZeroTrustDeviceDefaultProfileSplitTunnelEntry{
				{Address: "10.0.0.0/8"},
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept virtual networks with a default", func() {
			spec := baseSpec()
			spec.VirtualNetworks = &CloudflareZeroTrustDeviceDefaultProfileVirtualNetworks{
				Allowed:                 []*foreignkeyv1.StringValueOrRef{literal(testVnetID)},
				DefaultVirtualNetworkId: literal(testVnetID),
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept the zone certificates fold with enabled false", func() {
			spec := baseSpec()
			spec.ZoneCertificates = &CloudflareZeroTrustDeviceDefaultProfileZoneCertificates{
				ZoneId:  literal(testZoneID),
				Enabled: proto.Bool(false),
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept fallback domains with resolvers", func() {
			spec := baseSpec()
			spec.FallbackDomains = []*CloudflareZeroTrustDeviceDefaultProfileFallbackDomain{
				{Suffix: "corp.internal", DnsServer: []string{"10.0.0.53"}},
				{Suffix: "localdomain"},
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a proxy service mode with a port", func() {
			spec := baseSpec()
			spec.ServiceModeV2 = &CloudflareZeroTrustDeviceDefaultProfileServiceMode{
				Mode: "proxy",
				Port: proto.Int64(40080),
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject exclude and include set together", func() {
			spec := baseSpec()
			spec.Exclude = []*CloudflareZeroTrustDeviceDefaultProfileSplitTunnelEntry{{Address: "192.0.2.0/24"}}
			spec.Include = []*CloudflareZeroTrustDeviceDefaultProfileSplitTunnelEntry{{Address: "10.0.0.0/8"}}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a split tunnel entry with both address and host", func() {
			spec := baseSpec()
			spec.Exclude = []*CloudflareZeroTrustDeviceDefaultProfileSplitTunnelEntry{
				{Address: "192.0.2.0/24", Host: "internal.example.com"},
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a split tunnel entry with neither address nor host", func() {
			spec := baseSpec()
			spec.Include = []*CloudflareZeroTrustDeviceDefaultProfileSplitTunnelEntry{{Description: "empty"}}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject virtual networks without a default", func() {
			spec := baseSpec()
			spec.VirtualNetworks = &CloudflareZeroTrustDeviceDefaultProfileVirtualNetworks{
				Allowed: []*foreignkeyv1.StringValueOrRef{literal(testVnetID)},
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject virtual networks with an empty allowed list", func() {
			spec := baseSpec()
			spec.VirtualNetworks = &CloudflareZeroTrustDeviceDefaultProfileVirtualNetworks{
				DefaultVirtualNetworkId: literal(testVnetID),
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject zone certificates without the explicit enabled flag", func() {
			spec := baseSpec()
			spec.ZoneCertificates = &CloudflareZeroTrustDeviceDefaultProfileZoneCertificates{
				ZoneId: literal(testZoneID),
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject zone certificates without a zone", func() {
			spec := baseSpec()
			spec.ZoneCertificates = &CloudflareZeroTrustDeviceDefaultProfileZoneCertificates{
				Enabled: proto.Bool(true),
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a fallback domain without a suffix", func() {
			spec := baseSpec()
			spec.FallbackDomains = []*CloudflareZeroTrustDeviceDefaultProfileFallbackDomain{
				{DnsServer: []string{"10.0.0.53"}},
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a service mode port above 65535", func() {
			spec := baseSpec()
			spec.ServiceModeV2 = &CloudflareZeroTrustDeviceDefaultProfileServiceMode{
				Mode: "proxy",
				Port: proto.Int64(70000),
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed account_id", func() {
			spec := baseSpec()
			spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})
	})
})
