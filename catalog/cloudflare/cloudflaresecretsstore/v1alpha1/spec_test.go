package cloudflaresecretsstorev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestCloudflareSecretsStoreSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareSecretsStoreSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func validStore(spec *CloudflareSecretsStoreSpec) *CloudflareSecretsStore {
	return &CloudflareSecretsStore{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareSecretsStore",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-secrets-store",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("CloudflareSecretsStoreSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept an account id and a name", func() {
			input := validStore(&CloudflareSecretsStoreSpec{
				AccountId: testAccountID,
				Name:      "account-secrets",
			})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a malformed account id", func() {
			input := validStore(&CloudflareSecretsStoreSpec{
				AccountId: "not-an-account",
				Name:      "account-secrets",
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing name", func() {
			input := validStore(&CloudflareSecretsStoreSpec{
				AccountId: testAccountID,
			})
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
