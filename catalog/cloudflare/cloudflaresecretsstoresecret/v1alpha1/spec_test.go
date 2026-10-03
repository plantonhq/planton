package cloudflaresecretsstoresecretv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestCloudflareSecretsStoreSecretSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareSecretsStoreSecretSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func literal(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func validSecret(spec *CloudflareSecretsStoreSecretSpec) *CloudflareSecretsStoreSecret {
	return &CloudflareSecretsStoreSecret{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareSecretsStoreSecret",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-secret",
		},
		Spec: spec,
	}
}

func baseSpec() *CloudflareSecretsStoreSecretSpec {
	return &CloudflareSecretsStoreSecretSpec{
		AccountId: testAccountID,
		StoreId:   literal("7b0a3d5c1e9f42c68d1a2b3c4d5e6f70"),
		Name:      "openai-api-key",
		Value:     literal("sk-test-not-a-real-key"),
		Scopes:    []string{"workers"},
	}
}

var _ = ginkgo.Describe("CloudflareSecretsStoreSecretSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a single scope", func() {
			gomega.Expect(protovalidate.Validate(validSecret(baseSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("should accept multiple scopes in alphabetical order", func() {
			spec := baseSpec()
			spec.Scopes = []string{"access", "ai_gateway", "workers"}
			gomega.Expect(protovalidate.Validate(validSecret(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept every scope in canonical order", func() {
			spec := baseSpec()
			spec.Scopes = []string{"access", "ai_gateway", "dex", "workers"}
			gomega.Expect(protovalidate.Validate(validSecret(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject scopes out of alphabetical order -- Cloudflare returns them sorted and any other order drifts forever", func() {
			spec := baseSpec()
			spec.Scopes = []string{"workers", "access"}
			gomega.Expect(protovalidate.Validate(validSecret(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown scope", func() {
			spec := baseSpec()
			spec.Scopes = []string{"pages"}
			gomega.Expect(protovalidate.Validate(validSecret(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject duplicate scopes", func() {
			spec := baseSpec()
			spec.Scopes = []string{"workers", "workers"}
			gomega.Expect(protovalidate.Validate(validSecret(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty scopes list", func() {
			spec := baseSpec()
			spec.Scopes = nil
			gomega.Expect(protovalidate.Validate(validSecret(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing value", func() {
			spec := baseSpec()
			spec.Value = nil
			gomega.Expect(protovalidate.Validate(validSecret(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing store reference", func() {
			spec := baseSpec()
			spec.StoreId = nil
			gomega.Expect(protovalidate.Validate(validSecret(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing name", func() {
			spec := baseSpec()
			spec.Name = ""
			gomega.Expect(protovalidate.Validate(validSecret(spec))).NotTo(gomega.BeNil())
		})
	})
})
