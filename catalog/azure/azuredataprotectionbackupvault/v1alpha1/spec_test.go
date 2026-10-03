package azuredataprotectionbackupvaultv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureDataProtectionBackupVaultSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDataProtectionBackupVaultSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func strPtr(v string) *string { return &v }
func int32Ptr(v int32) *int32 { return &v }

// validResource returns a minimal valid locally-redundant vault that
// individual cases mutate into the shape under test.
func validResource() *AzureDataProtectionBackupVault {
	return &AzureDataProtectionBackupVault{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDataProtectionBackupVault",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-backup-vault",
		},
		Spec: &AzureDataProtectionBackupVaultSpec{
			Region:        "eastus",
			ResourceGroup: literal("backup-rg"),
			Name:          "app-backup-vault",
			DatastoreType: "VaultStore",
			Redundancy:    "LocallyRedundant",
		},
	}
}

var _ = ginkgo.Describe("AzureDataProtectionBackupVaultSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_data_protection_backup_vault", func() {

			ginkgo.It("should not return a validation error for a minimal vault", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every datastore tier", func() {
				for _, tier := range []string{"VaultStore", "OperationalStore", "ArchiveStore"} {
					input := validResource()
					input.Spec.DatastoreType = tier
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept cross-region restore on a geo-redundant vault", func() {
				input := validResource()
				input.Spec.Redundancy = "GeoRedundant"
				input.Spec.CrossRegionRestoreEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the soft-delete retention bounds", func() {
				for _, days := range []int32{14, 180} {
					input := validResource()
					input.Spec.RetentionDurationInDays = int32Ptr(days)
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept every soft-delete posture", func() {
				for _, state := range []string{"On", "Off", "AlwaysOn"} {
					input := validResource()
					input.Spec.SoftDelete = strPtr(state)
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept every immutability posture", func() {
				for _, state := range []string{"Disabled", "Unlocked", "Locked"} {
					input := validResource()
					input.Spec.Immutability = strPtr(state)
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataProtectionBackupVaultIdentity{
					Type: AzureDataProtectionBackupVaultIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataProtectionBackupVaultIdentity{
					Type:        AzureDataProtectionBackupVaultIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept encryption with a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataProtectionBackupVaultIdentity{
					Type: AzureDataProtectionBackupVaultIdentityType_SYSTEM_ASSIGNED,
				}
				input.Spec.Encryption = &AzureDataProtectionBackupVaultEncryption{
					KeyId: literal("https://kv.vault.azure.net/keys/backup-key"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept encryption with a combined identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataProtectionBackupVaultIdentity{
					Type:        AzureDataProtectionBackupVaultIdentityType_SYSTEM_AND_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai")},
				}
				input.Spec.Encryption = &AzureDataProtectionBackupVaultEncryption{
					KeyId: literal("https://kv.vault.azure.net/keys/backup-key"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept user tags", func() {
				input := validResource()
				input.Spec.Tags = map[string]string{"cost-center": "platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_data_protection_backup_vault", func() {

			ginkgo.It("should reject a missing name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a one-character name", func() {
				input := validResource()
				input.Spec.Name = "a"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name over 50 characters", func() {
				input := validResource()
				input.Spec.Name = "a123456789a123456789a123456789a123456789a1234567891"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with invalid characters", func() {
				input := validResource()
				input.Spec.Name = "vault_with_underscores"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing datastore type", func() {
				input := validResource()
				input.Spec.DatastoreType = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown datastore type", func() {
				input := validResource()
				input.Spec.DatastoreType = "SnapshotStore"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown redundancy", func() {
				input := validResource()
				input.Spec.Redundancy = "ReadAccessGeoRedundant"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject cross-region restore without geo redundancy", func() {
				input := validResource()
				input.Spec.Redundancy = "LocallyRedundant"
				input.Spec.CrossRegionRestoreEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject soft-delete retention below 14 days", func() {
				input := validResource()
				input.Spec.RetentionDurationInDays = int32Ptr(13)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject soft-delete retention above 180 days", func() {
				input := validResource()
				input.Spec.RetentionDurationInDays = int32Ptr(181)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown soft-delete posture", func() {
				input := validResource()
				input.Spec.SoftDelete = strPtr("Enabled")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown immutability posture", func() {
				input := validResource()
				input.Spec.Immutability = strPtr("Enabled")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataProtectionBackupVaultIdentity{
					Type: AzureDataProtectionBackupVaultIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataProtectionBackupVaultIdentity{
					Type:        AzureDataProtectionBackupVaultIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai")},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject encryption without an identity", func() {
				input := validResource()
				input.Spec.Encryption = &AzureDataProtectionBackupVaultEncryption{
					KeyId: literal("https://kv.vault.azure.net/keys/backup-key"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject encryption with a user-assigned-only identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataProtectionBackupVaultIdentity{
					Type:        AzureDataProtectionBackupVaultIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal("/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/uai")},
				}
				input.Spec.Encryption = &AzureDataProtectionBackupVaultEncryption{
					KeyId: literal("https://kv.vault.azure.net/keys/backup-key"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject encryption without a key", func() {
				input := validResource()
				input.Spec.Identity = &AzureDataProtectionBackupVaultIdentity{
					Type: AzureDataProtectionBackupVaultIdentityType_SYSTEM_ASSIGNED,
				}
				input.Spec.Encryption = &AzureDataProtectionBackupVaultEncryption{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
