package azuredisksnapshotv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureDiskSnapshotSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDiskSnapshotSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testDiskId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/app-rg/providers/Microsoft.Compute/disks/app-disk"

// validResource returns a valid Copy-mode snapshot that individual
// cases mutate into the shape under test.
func validResource() *AzureDiskSnapshot {
	return &AzureDiskSnapshot{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDiskSnapshot",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-snapshot",
		},
		Spec: &AzureDiskSnapshotSpec{
			ResourceGroup:    literal("app-rg"),
			Name:             "app-disk-snap",
			Region:           "eastus",
			CreateOption:     "Copy",
			SourceResourceId: literal(testDiskId),
		},
	}
}

// validEncryptionSettings returns a complete ADE settings block.
func validEncryptionSettings() *AzureDiskSnapshotEncryptionSettings {
	return &AzureDiskSnapshotEncryptionSettings{
		DiskEncryptionKey: &AzureDiskSnapshotDiskEncryptionKey{
			SecretUrl:     "https://vault.vault.azure.net/secrets/dek/0000",
			SourceVaultId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/vault"),
		},
	}
}

var _ = ginkgo.Describe("AzureDiskSnapshotSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_disk_snapshot", func() {

			ginkgo.It("should not return a validation error for the minimal Copy shape", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an incremental snapshot", func() {
				input := validResource()
				input.Spec.IncrementalEnabled = true
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an Import shape with a blob source", func() {
				input := validResource()
				input.Spec.CreateOption = "Import"
				input.Spec.SourceResourceId = nil
				input.Spec.SourceUri = "https://acmevhds.blob.core.windows.net/vhds/base.vhd"
				input.Spec.StorageAccountId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/acmevhds")
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a Copy shape without a source (the provider's own looseness -- Azure validates the pairing)", func() {
				input := validResource()
				input.Spec.SourceResourceId = nil
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept every network access policy", func() {
				input := validResource()
				for _, policy := range []string{"", "AllowAll", "AllowPrivate", "DenyAll"} {
					input.Spec.NetworkAccessPolicy = policy
					gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "expected %q to be accepted", policy)
				}
			})

			ginkgo.It("should accept a private posture with a disk access resource", func() {
				input := validResource()
				input.Spec.NetworkAccessPolicy = "AllowPrivate"
				input.Spec.DiskAccessId = literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.Compute/diskAccesses/da")
				input.Spec.PublicNetworkAccessEnabled = proto.Bool(false)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept a complete encryption settings block", func() {
				input := validResource()
				input.Spec.EncryptionSettings = validEncryptionSettings()
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
				input.Spec.EncryptionSettings.KeyEncryptionKey = &AzureDiskSnapshotKeyEncryptionKey{
					KeyUrl:        "https://vault.vault.azure.net/keys/kek/0000",
					SourceVaultId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/vault"),
				}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit disk size", func() {
				input := validResource()
				input.Spec.DiskSizeGb = proto.Int32(64)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_disk_snapshot", func() {

			ginkgo.It("should reject a missing resource group, region, or create option", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.Region = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input = validResource()
				input.Spec.CreateOption = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject create options outside the provider vocabulary", func() {
				input := validResource()
				for _, option := range []string{"copy", "Restore", "FromImage"} {
					input.Spec.CreateOption = option
					gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "expected %q to be rejected", option)
				}
			})

			ginkgo.It("should reject names with dots (snapshots allow only letters, numbers, dashes, and underscores)", func() {
				input := validResource()
				input.Spec.Name = "app.disk.snap"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names over 80 characters and empty names", func() {
				input := validResource()
				input.Spec.Name = strings.Repeat("a", 81)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.Name = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown network access policy", func() {
				input := validResource()
				input.Spec.NetworkAccessPolicy = "Private"
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a zero disk size", func() {
				input := validResource()
				input.Spec.DiskSizeGb = proto.Int32(0)
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject encryption settings without the disk encryption key", func() {
				input := validResource()
				input.Spec.EncryptionSettings = &AzureDiskSnapshotEncryptionSettings{}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a disk encryption key missing its secret URL or vault", func() {
				input := validResource()
				input.Spec.EncryptionSettings = validEncryptionSettings()
				input.Spec.EncryptionSettings.DiskEncryptionKey.SecretUrl = ""
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
				input.Spec.EncryptionSettings = validEncryptionSettings()
				input.Spec.EncryptionSettings.DiskEncryptionKey.SourceVaultId = nil
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a key encryption key missing its key URL", func() {
				input := validResource()
				input.Spec.EncryptionSettings = validEncryptionSettings()
				input.Spec.EncryptionSettings.KeyEncryptionKey = &AzureDiskSnapshotKeyEncryptionKey{
					SourceVaultId: literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/vault"),
				}
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
			})
		})
	})
})
