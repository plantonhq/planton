package azurestoragecontainerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureStorageContainerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureStorageContainerSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const accountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/plantonapp"

// minimal valid spec: a private container.
func minimalSpec() *AzureStorageContainer {
	return &AzureStorageContainer{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureStorageContainer",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-container",
		},
		Spec: &AzureStorageContainerSpec{
			StorageAccountId: literal(accountId),
			ContainerName:    "uploads",
		},
	}
}

var _ = ginkgo.Describe("AzureStorageContainerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal private container", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept every access type", func() {
			for _, access := range []AzureStorageContainerAccessType{
				AzureStorageContainerAccessType_PRIVATE,
				AzureStorageContainerAccessType_BLOB,
				AzureStorageContainerAccessType_CONTAINER,
			} {
				input := minimalSpec()
				input.Spec.ContainerAccessType = access
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "access type %v must be accepted", access)
			}
		})

		ginkgo.It("should accept hyphenated names within the length bounds", func() {
			for _, name := range []string{"app", "raw-events", "tenant-42-artifacts"} {
				input := minimalSpec()
				input.Spec.ContainerName = name
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "name %q must be accepted", name)
			}
		})

		ginkgo.It("should accept an encryption scope with an override posture", func() {
			override := false
			input := minimalSpec()
			input.Spec.DefaultEncryptionScope = literal("tenant42scope")
			input.Spec.EncryptionScopeOverrideEnabled = &override
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a valueFrom reference for the encryption scope", func() {
			input := minimalSpec()
			input.Spec.DefaultEncryptionScope = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureStorageEncryptionScope,
						Name:      "tenant42-scope",
						FieldPath: "status.outputs.encryption_scope_name",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept container metadata", func() {
			input := minimalSpec()
			input.Spec.Metadata = map[string]string{"purpose": "uploads", "owner": "platform"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
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

		ginkgo.It("should reject names that break the container naming rules", func() {
			for _, name := range []string{"", "ab", "UPPER", "-leading", "trailing-", "double--hyphen", "has_underscore"} {
				input := minimalSpec()
				input.Spec.ContainerName = name
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "name %q must be rejected", name)
			}
		})

		ginkgo.It("should reject a name longer than 63 characters", func() {
			input := minimalSpec()
			input.Spec.ContainerName = "a123456789b123456789c123456789d123456789e123456789f123456789g123"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject container metadata keys that are not lowercase C# identifiers", func() {
			input := minimalSpec()
			input.Spec.Metadata = map[string]string{"e2e-suite": "azure"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an override posture without an encryption scope", func() {
			override := false
			input := minimalSpec()
			input.Spec.EncryptionScopeOverrideEnabled = &override
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
