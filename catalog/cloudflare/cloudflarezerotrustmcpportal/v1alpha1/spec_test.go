package cloudflarezerotrustmcpportalv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareZeroTrustMcpPortalSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustMcpPortalSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func validPortal(spec *CloudflareZeroTrustMcpPortalSpec) *CloudflareZeroTrustMcpPortal {
	return &CloudflareZeroTrustMcpPortal{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustMcpPortal",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-portal",
		},
		Spec: spec,
	}
}

func baseSpec() *CloudflareZeroTrustMcpPortalSpec {
	return &CloudflareZeroTrustMcpPortalSpec{
		AccountId: testAccountID,
		PortalId:  "eng-tools",
		Hostname:  "mcp.example.com",
		Name:      "Engineering Tools",
	}
}

var _ = ginkgo.Describe("CloudflareZeroTrustMcpPortalSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal portal", func() {
			gomega.Expect(protovalidate.Validate(validPortal(baseSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a portal with server rows and overrides", func() {
			spec := baseSpec()
			spec.Servers = []*CloudflareZeroTrustMcpPortalServer{
				{
					ServerId: literal("docs-search"),
					UpdatedTools: []*CloudflareZeroTrustMcpPortalItemOverride{
						{Name: "delete_page", Enabled: boolPtr(false)},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(validPortal(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing portal_id", func() {
			spec := baseSpec()
			spec.PortalId = ""
			gomega.Expect(protovalidate.Validate(validPortal(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing hostname", func() {
			spec := baseSpec()
			spec.Hostname = ""
			gomega.Expect(protovalidate.Validate(validPortal(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing name", func() {
			spec := baseSpec()
			spec.Name = ""
			gomega.Expect(protovalidate.Validate(validPortal(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a server row without a server reference", func() {
			spec := baseSpec()
			spec.Servers = []*CloudflareZeroTrustMcpPortalServer{{}}
			gomega.Expect(protovalidate.Validate(validPortal(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an override row without the upstream name", func() {
			spec := baseSpec()
			spec.Servers = []*CloudflareZeroTrustMcpPortalServer{
				{
					ServerId: literal("docs-search"),
					UpdatedPrompts: []*CloudflareZeroTrustMcpPortalItemOverride{
						{Alias: "renamed"},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(validPortal(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed account_id", func() {
			spec := baseSpec()
			spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(validPortal(spec))).NotTo(gomega.BeNil())
		})
	})
})

func boolPtr(b bool) *bool { return &b }
