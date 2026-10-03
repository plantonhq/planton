package cloudflarezerotrustgatewaysettingsv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareZeroTrustGatewaySettingsSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustGatewaySettingsSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func validSettings(spec *CloudflareZeroTrustGatewaySettingsSpec) *CloudflareZeroTrustGatewaySettings {
	return &CloudflareZeroTrustGatewaySettings{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustGatewaySettings",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-gateway-settings",
		},
		Spec: spec,
	}
}

func baseSpec() *CloudflareZeroTrustGatewaySettingsSpec {
	return &CloudflareZeroTrustGatewaySettingsSpec{
		AccountId: testAccountID,
	}
}

var _ = ginkgo.Describe("CloudflareZeroTrustGatewaySettingsSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept an account-only spec (nothing managed yet)", func() {
			gomega.Expect(protovalidate.Validate(validSettings(baseSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a branded block page", func() {
			spec := baseSpec()
			spec.Settings = &CloudflareZeroTrustGatewayConfig{
				BlockPage: &CloudflareZeroTrustGatewayBlockPage{
					Enabled:         proto.Bool(true),
					Mode:            "customized_block_page",
					Name:            "Blocked by Acme IT",
					HeaderText:      "This site is blocked",
					BackgroundColor: "#1e293b",
				},
			}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a redirect block page with a target", func() {
			spec := baseSpec()
			spec.Settings = &CloudflareZeroTrustGatewayConfig{
				BlockPage: &CloudflareZeroTrustGatewayBlockPage{
					Mode:      "redirect_uri",
					TargetUri: "https://it.example.com/blocked",
				},
			}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept the full logging tree", func() {
			spec := baseSpec()
			spec.Logging = &CloudflareZeroTrustGatewayLogging{
				RedactPii: proto.Bool(true),
				SettingsByRuleType: &CloudflareZeroTrustGatewayLoggingByRuleType{
					Dns:  &CloudflareZeroTrustGatewayLoggingRule{LogAll: proto.Bool(true)},
					Http: &CloudflareZeroTrustGatewayLoggingRule{LogBlocks: proto.Bool(true)},
					L4:   &CloudflareZeroTrustGatewayLoggingRule{},
				},
			}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a PAC file row", func() {
			spec := baseSpec()
			spec.PacFiles = []*CloudflareZeroTrustGatewayPacFile{
				{
					Name:     "default-proxy",
					Contents: "function FindProxyForURL(url, host) { return \"DIRECT\"; }",
				},
			}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a certificate selection and TTL cap", func() {
			spec := baseSpec()
			spec.Settings = &CloudflareZeroTrustGatewayConfig{
				Certificate: &CloudflareZeroTrustGatewayCertificate{
					Id: literal("f70ff985a4ef4643bbbc4a0ed4fc8415"),
				},
				MaxTtlSecs: proto.Int64(300),
			}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject an unknown block page mode", func() {
			spec := baseSpec()
			spec.Settings = &CloudflareZeroTrustGatewayConfig{
				BlockPage: &CloudflareZeroTrustGatewayBlockPage{Mode: "inline"},
			}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject redirect mode without a target", func() {
			spec := baseSpec()
			spec.Settings = &CloudflareZeroTrustGatewayConfig{
				BlockPage: &CloudflareZeroTrustGatewayBlockPage{Mode: "redirect_uri"},
			}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown body-scanning mode", func() {
			spec := baseSpec()
			spec.Settings = &CloudflareZeroTrustGatewayConfig{
				BodyScanning: &CloudflareZeroTrustGatewayBodyScanning{InspectionMode: "full"},
			}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown inspection mode", func() {
			spec := baseSpec()
			spec.Settings = &CloudflareZeroTrustGatewayConfig{
				Inspection: &CloudflareZeroTrustGatewayInspection{Mode: "auto"},
			}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown sandbox fallback action", func() {
			spec := baseSpec()
			spec.Settings = &CloudflareZeroTrustGatewayConfig{
				Sandbox: &CloudflareZeroTrustGatewaySandbox{FallbackAction: "quarantine"},
			}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a TTL cap below 60 seconds", func() {
			spec := baseSpec()
			spec.Settings = &CloudflareZeroTrustGatewayConfig{MaxTtlSecs: proto.Int64(59)}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a TTL cap above 36000 seconds", func() {
			spec := baseSpec()
			spec.Settings = &CloudflareZeroTrustGatewayConfig{MaxTtlSecs: proto.Int64(36001)}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a certificate block without an id", func() {
			spec := baseSpec()
			spec.Settings = &CloudflareZeroTrustGatewayConfig{
				Certificate: &CloudflareZeroTrustGatewayCertificate{},
			}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a PAC file without contents", func() {
			spec := baseSpec()
			spec.PacFiles = []*CloudflareZeroTrustGatewayPacFile{
				{Name: "default-proxy"},
			}
			gomega.Expect(protovalidate.Validate(validSettings(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed account_id", func() {
			spec := baseSpec()
			spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(validSettings(spec))).NotTo(gomega.BeNil())
		})
	})
})
