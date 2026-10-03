package azurekeyvaultsecretv1alpha1

import (
	"fmt"
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureKeyVaultSecretSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureKeyVaultSecretSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testKeyVaultId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.KeyVault/vaults/platform-kv"

// validResource returns a minimal valid secret that individual cases
// mutate into the shape under test.
func validResource() *AzureKeyVaultSecret {
	return &AzureKeyVaultSecret{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureKeyVaultSecret",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-secret",
		},
		Spec: &AzureKeyVaultSecretSpec{
			Name:       "db-password",
			KeyVaultId: literal(testKeyVaultId),
			Value:      literal("s3cr3t-value"),
		},
	}
}

var _ = ginkgo.Describe("AzureKeyVaultSecretSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_key_vault_secret", func() {

			ginkgo.It("should not return a validation error for a minimal secret", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a content type", func() {
				input := validResource()
				input.Spec.ContentType = "text/plain"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept RFC 3339 activation and expiry instants", func() {
				input := validResource()
				notBefore := "2027-01-01T00:00:00Z"
				expiration := "2028-01-01T00:00:00Z"
				input.Spec.NotBeforeDate = &notBefore
				input.Spec.ExpirationDate = &expiration
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a fractional-seconds RFC 3339 instant", func() {
				input := validResource()
				expiration := "2028-01-01T00:00:00.500Z"
				input.Spec.ExpirationDate = &expiration
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a 127-character name", func() {
				input := validResource()
				input.Spec.Name = strings.Repeat("a", 127)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept 15 user tags (Key Vault's own cap)", func() {
				input := validResource()
				input.Spec.Tags = map[string]string{}
				for i := 0; i < 15; i++ {
					input.Spec.Tags[fmt.Sprintf("tag-%02d", i)] = "value"
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_key_vault_secret", func() {

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with an underscore", func() {
				input := validResource()
				input.Spec.Name = "db_password"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name over 127 characters", func() {
				input := validResource()
				input.Spec.Name = strings.Repeat("a", 128)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing vault reference", func() {
				input := validResource()
				input.Spec.KeyVaultId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing value", func() {
				input := validResource()
				input.Spec.Value = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a date-only activation instant", func() {
				input := validResource()
				notBefore := "2027-01-01"
				input.Spec.NotBeforeDate = &notBefore
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an expiry instant without the Z suffix", func() {
				input := validResource()
				expiration := "2028-01-01T00:00:00+05:30"
				input.Spec.ExpirationDate = &expiration
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject 16 user tags", func() {
				input := validResource()
				input.Spec.Tags = map[string]string{}
				for i := 0; i < 16; i++ {
					input.Spec.Tags[fmt.Sprintf("tag-%02d", i)] = "value"
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
