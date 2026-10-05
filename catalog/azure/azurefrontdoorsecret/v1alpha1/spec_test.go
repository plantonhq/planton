package azurefrontdoorsecretv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureFrontDoorSecretSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFrontDoorSecretSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const profileId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd"

const kvCertificateId = "https://planton-vault.vault.azure.net/certificates/wildcard-example-com"

// minimal valid spec: a secret wrapping a versionless Key Vault
// certificate reference (rotation follows the latest version).
func minimalSpec() *AzureFrontDoorSecret {
	return &AzureFrontDoorSecret{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFrontDoorSecret",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-front-door-secret",
		},
		Spec: &AzureFrontDoorSecretSpec{
			ProfileId:             literal(profileId),
			SecretName:            "wildcard-example-com",
			KeyVaultCertificateId: literal(kvCertificateId),
		},
	}
}

var _ = ginkgo.Describe("AzureFrontDoorSecretSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal secret", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a versioned Key Vault certificate reference", func() {
			input := minimalSpec()
			input.Spec.KeyVaultCertificateId = literal(kvCertificateId + "/0123456789abcdef0123456789abcdef")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept secret name boundaries (2 and 260 characters)", func() {
			input := minimalSpec()
			input.Spec.SecretName = "ab"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.SecretName = "a" + strings.Repeat("b", 258) + "c"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing profile reference", func() {
			input := minimalSpec()
			input.Spec.ProfileId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing secret name", func() {
			input := minimalSpec()
			input.Spec.SecretName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a one-character secret name", func() {
			input := minimalSpec()
			input.Spec.SecretName = "a"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a secret name over 260 characters", func() {
			input := minimalSpec()
			input.Spec.SecretName = "a" + strings.Repeat("b", 259) + "c"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a secret name with a trailing hyphen", func() {
			input := minimalSpec()
			input.Spec.SecretName = "wildcard-"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a secret name with dots", func() {
			input := minimalSpec()
			input.Spec.SecretName = "wildcard.example.com"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing Key Vault certificate reference", func() {
			input := minimalSpec()
			input.Spec.KeyVaultCertificateId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a wrong kind", func() {
			input := minimalSpec()
			input.Kind = "AzureFrontDoorSecrets"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject missing metadata", func() {
			input := minimalSpec()
			input.Metadata = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
