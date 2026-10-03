package azurestorageencryptionscopev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureStorageEncryptionScopeSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureStorageEncryptionScopeSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const accountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/plantonapp"

// minimal valid spec: a platform-managed-key scope.
func minimalSpec() *AzureStorageEncryptionScope {
	return &AzureStorageEncryptionScope{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureStorageEncryptionScope",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-scope",
		},
		Spec: &AzureStorageEncryptionScopeSpec{
			StorageAccountId: literal(accountId),
			ScopeName:        "tenant42scope",
			Source:           AzureStorageEncryptionScopeSource_MICROSOFT_STORAGE,
		},
	}
}

var _ = ginkgo.Describe("AzureStorageEncryptionScopeSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal platform-managed scope", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a Key Vault scope with a key reference", func() {
			input := minimalSpec()
			input.Spec.Source = AzureStorageEncryptionScopeSource_MICROSOFT_KEY_VAULT
			input.Spec.KeyVaultKeyId = literal("https://myvault.vault.azure.net/keys/tenant42")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a Key Vault scope with a valueFrom key reference", func() {
			input := minimalSpec()
			input.Spec.Source = AzureStorageEncryptionScopeSource_MICROSOFT_KEY_VAULT
			input.Spec.KeyVaultKeyId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureKeyVaultKey,
						Name:      "tenant42-key",
						FieldPath: "status.outputs.versionless_id",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept infrastructure encryption", func() {
			input := minimalSpec()
			input.Spec.InfrastructureEncryptionRequired = true
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept mixed-case alphanumeric names within the length bounds", func() {
			for _, name := range []string{"tenA", "Tenant42Scope", "SCOPE2026"} {
				input := minimalSpec()
				input.Spec.ScopeName = name
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "name %q must be accepted", name)
			}
		})

		ginkgo.It("should accept a valueFrom reference for the account", func() {
			input := minimalSpec()
			input.Spec.StorageAccountId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureStorageAccount,
						Name:      "app-storage",
						FieldPath: "status.outputs.storage_account_id",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing storage account reference", func() {
			input := minimalSpec()
			input.Spec.StorageAccountId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject names that break the scope naming rules", func() {
			for _, name := range []string{"", "abc", "has-hyphen", "has_underscore", "has space"} {
				input := minimalSpec()
				input.Spec.ScopeName = name
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "name %q must be rejected", name)
			}
		})

		ginkgo.It("should reject a name longer than 63 characters", func() {
			input := minimalSpec()
			input.Spec.ScopeName = "a123456789b123456789c123456789d123456789e123456789f123456789g123"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unspecified source", func() {
			input := minimalSpec()
			input.Spec.Source = AzureStorageEncryptionScopeSource_azure_storage_encryption_scope_source_unspecified
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an undefined source enum value", func() {
			input := minimalSpec()
			input.Spec.Source = AzureStorageEncryptionScopeSource(99)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a Key Vault source without a key reference", func() {
			input := minimalSpec()
			input.Spec.Source = AzureStorageEncryptionScopeSource_MICROSOFT_KEY_VAULT
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
