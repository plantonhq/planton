package cloudflarezerotrustdevicecustomprofilev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareZeroTrustDeviceCustomProfileSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustDeviceCustomProfileSpec Custom Validation Tests")
}

const (
	testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"
	testVnetID    = "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415"
)

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func validProfile(spec *CloudflareZeroTrustDeviceCustomProfileSpec) *CloudflareZeroTrustDeviceCustomProfile {
	return &CloudflareZeroTrustDeviceCustomProfile{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustDeviceCustomProfile",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-custom-profile",
		},
		Spec: spec,
	}
}

func baseSpec() *CloudflareZeroTrustDeviceCustomProfileSpec {
	return &CloudflareZeroTrustDeviceCustomProfileSpec{
		AccountId:  testAccountID,
		Name:       "contractors",
		Match:      `identity.email == "dev@example.com"`,
		Precedence: 100,
	}
}

var _ = ginkgo.Describe("CloudflareZeroTrustDeviceCustomProfileSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal profile", func() {
			gomega.Expect(protovalidate.Validate(validProfile(baseSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a disabled profile with a description", func() {
			spec := baseSpec()
			spec.Enabled = proto.Bool(false)
			spec.Description = "settings for contractor laptops"
			gomega.Expect(protovalidate.Validate(validProfile(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept the shared settings body", func() {
			spec := baseSpec()
			spec.AutoConnect = proto.Int64(0)
			spec.SwitchLocked = proto.Bool(true)
			spec.LanAllowMinutes = proto.Int64(15)
			spec.Exclude = []*CloudflareZeroTrustDeviceCustomProfileSplitTunnelEntry{
				{Address: "192.0.2.0/24"},
			}
			spec.VirtualNetworks = &CloudflareZeroTrustDeviceCustomProfileVirtualNetworks{
				Allowed:                 []*foreignkeyv1.StringValueOrRef{literal(testVnetID)},
				DefaultVirtualNetworkId: literal(testVnetID),
			}
			spec.FallbackDomains = []*CloudflareZeroTrustDeviceCustomProfileFallbackDomain{
				{Suffix: "corp.internal", DnsServer: []string{"10.0.0.53"}},
			}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing name", func() {
			spec := baseSpec()
			spec.Name = ""
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing match expression", func() {
			spec := baseSpec()
			spec.Match = ""
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing precedence -- ordering must be explicit", func() {
			spec := baseSpec()
			spec.Precedence = 0
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject exclude and include set together", func() {
			spec := baseSpec()
			spec.Exclude = []*CloudflareZeroTrustDeviceCustomProfileSplitTunnelEntry{{Address: "192.0.2.0/24"}}
			spec.Include = []*CloudflareZeroTrustDeviceCustomProfileSplitTunnelEntry{{Address: "10.0.0.0/8"}}
			gomega.Expect(protovalidate.Validate(validProfile(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a split tunnel entry with both address and host", func() {
			spec := baseSpec()
			spec.Include = []*CloudflareZeroTrustDeviceCustomProfileSplitTunnelEntry{
				{Address: "10.0.0.0/8", Host: "internal.example.com"},
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
