package azurebackupcontainerstorageaccountv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureBackupContainerStorageAccountSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureBackupContainerStorageAccountSpec Validation Tests")
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

// validResource returns a minimal valid registration that individual
// cases mutate into the shape under test.
func validResource() *AzureBackupContainerStorageAccount {
	return &AzureBackupContainerStorageAccount{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureBackupContainerStorageAccount",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-backup-container-storage-account",
		},
		Spec: &AzureBackupContainerStorageAccountSpec{
			ResourceGroup:     literal("backup-rg"),
			RecoveryVaultName: literal("backup-vault"),
			StorageAccountId:  literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/data-rg/providers/Microsoft.Storage/storageAccounts/appfiles"),
		},
	}
}

var _ = ginkgo.Describe("AzureBackupContainerStorageAccountSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_backup_container_storage_account", func() {

			ginkgo.It("should not return a validation error for literal values", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept references end to end", func() {
				input := validResource()
				input.Spec.ResourceGroup = ref("backup-rg")
				input.Spec.RecoveryVaultName = ref("backup-vault")
				input.Spec.StorageAccountId = ref("appfiles-account")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_backup_container_storage_account", func() {

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

			ginkgo.It("should reject a missing storage account", func() {
				input := validResource()
				input.Spec.StorageAccountId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an empty storage account literal", func() {
				input := validResource()
				input.Spec.StorageAccountId = literal("")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a wrong kind constant", func() {
				input := validResource()
				input.Kind = "AzureBackupContainer"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
