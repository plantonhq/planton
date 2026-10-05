package cloudflarezerotrustmcpserverv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareZeroTrustMcpServerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustMcpServerSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func validServer(spec *CloudflareZeroTrustMcpServerSpec) *CloudflareZeroTrustMcpServer {
	return &CloudflareZeroTrustMcpServer{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustMcpServer",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-mcp-server",
		},
		Spec: spec,
	}
}

func baseSpec() *CloudflareZeroTrustMcpServerSpec {
	return &CloudflareZeroTrustMcpServerSpec{
		AccountId: testAccountID,
		ServerId:  "docs-search",
		Name:      "Docs Search",
		Hostname:  "https://mcp.example.com",
		AuthType:  "unauthenticated",
	}
}

var _ = ginkgo.Describe("CloudflareZeroTrustMcpServerSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept an unauthenticated server", func() {
			gomega.Expect(protovalidate.Validate(validServer(baseSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a bearer server with credentials", func() {
			spec := baseSpec()
			spec.AuthType = "bearer"
			spec.AuthCredentials = literal("test-token-not-real")
			gomega.Expect(protovalidate.Validate(validServer(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept an oauth server with a client secret", func() {
			spec := baseSpec()
			spec.AuthType = "oauth"
			spec.ClientSecret = literal("test-secret-not-real")
			gomega.Expect(protovalidate.Validate(validServer(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept prompt and tool overrides", func() {
			spec := baseSpec()
			spec.UpdatedPrompts = []*CloudflareZeroTrustMcpServerItemOverride{
				{Name: "search", Alias: "Search the docs"},
			}
			spec.UpdatedTools = []*CloudflareZeroTrustMcpServerItemOverride{
				{Name: "delete_page", Enabled: boolPtr(false)},
			}
			gomega.Expect(protovalidate.Validate(validServer(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject an unknown auth_type", func() {
			spec := baseSpec()
			spec.AuthType = "basic"
			gomega.Expect(protovalidate.Validate(validServer(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing server_id", func() {
			spec := baseSpec()
			spec.ServerId = ""
			gomega.Expect(protovalidate.Validate(validServer(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing hostname", func() {
			spec := baseSpec()
			spec.Hostname = ""
			gomega.Expect(protovalidate.Validate(validServer(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing name", func() {
			spec := baseSpec()
			spec.Name = ""
			gomega.Expect(protovalidate.Validate(validServer(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an override row without the upstream name", func() {
			spec := baseSpec()
			spec.UpdatedTools = []*CloudflareZeroTrustMcpServerItemOverride{
				{Alias: "renamed"},
			}
			gomega.Expect(protovalidate.Validate(validServer(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed account_id", func() {
			spec := baseSpec()
			spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(validServer(spec))).NotTo(gomega.BeNil())
		})
	})
})

func boolPtr(b bool) *bool { return &b }
