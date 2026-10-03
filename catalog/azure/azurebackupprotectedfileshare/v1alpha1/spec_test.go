package azurebackupprotectedfilesharev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureBackupProtectedFileShareSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureBackupProtectedFileShareSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// ref builds a StringValueOrRef carrying a named reference.
func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

// validResource returns a minimal valid protected file share that
// individual cases mutate into the shape under test.
func validResource() *AzureBackupProtectedFileShare {
	return &AzureBackupProtectedFileShare{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureBackupProtectedFileShare",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-backup-protected-file-share",
		},
		Spec: &AzureBackupProtectedFileShareSpec{
			ResourceGroup:          literal("backup-rg"),
			RecoveryVaultName:      literal("backup-vault"),
			SourceStorageAccountId: literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/data-rg/providers/Microsoft.Storage/storageAccounts/appfiles"),
			SourceFileShareName:    literal("team-share"),
			BackupPolicyId:         literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/backup-rg/providers/Microsoft.RecoveryServices/vaults/backup-vault/backupPolicies/daily-share-policy"),
		},
	}
}

var _ = ginkgo.Describe("AzureBackupProtectedFileShareSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_backup_protected_file_share", func() {

			ginkgo.It("should not return a validation error for literal values", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept references end to end", func() {
				input := validResource()
				input.Spec.ResourceGroup = ref("backup-rg")
				input.Spec.RecoveryVaultName = ref("backup-vault")
				input.Spec.SourceStorageAccountId = ref("appfiles-registration")
				input.Spec.SourceFileShareName = ref("team-share")
				input.Spec.BackupPolicyId = ref("daily-share-policy")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_backup_protected_file_share", func() {

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing vault name", func() {
				input := validResource()
				input.Spec.RecoveryVaultName = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing storage account reference", func() {
				input := validResource()
				input.Spec.SourceStorageAccountId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty storage account literal", func() {
				input := validResource()
				input.Spec.SourceStorageAccountId = literal("")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing file share name", func() {
				input := validResource()
				input.Spec.SourceFileShareName = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing backup policy", func() {
				input := validResource()
				input.Spec.BackupPolicyId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a wrong kind constant", func() {
				input := validResource()
				input.Kind = "AzureBackupProtectedVm"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
