package azurerecoveryservicesvaultv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureRecoveryServicesVaultSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureRecoveryServicesVaultSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func boolPtr(v bool) *bool { return &v }

const (
	testKeyVersionlessId = "https://test-kv.vault.azure.net/keys/vault-cmk"
	testIdentityId       = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/rsv-uai"
	testResourceGuardId  = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.DataProtection/resourceGuards/guard"
)

// validResource returns a minimal valid vault that individual cases
// mutate into the shape under test.
func validResource() *AzureRecoveryServicesVault {
	return &AzureRecoveryServicesVault{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureRecoveryServicesVault",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-recovery-services-vault",
		},
		Spec: &AzureRecoveryServicesVaultSpec{
			Region:        "eastus",
			ResourceGroup: literal("backup-rg"),
			Name:          "backup-vault",
		},
	}
}

var _ = ginkgo.Describe("AzureRecoveryServicesVaultSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_recovery_services_vault", func() {

			ginkgo.It("should not return a validation error for a minimal vault", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the legacy RS0 sku", func() {
				input := validResource()
				sku := "RS0"
				input.Spec.Sku = &sku
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept locally redundant storage without cross-region restore", func() {
				input := validResource()
				mode := "LocallyRedundant"
				input.Spec.StorageModeType = &mode
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept cross-region restore on the geo-redundant default", func() {
				input := validResource()
				input.Spec.CrossRegionRestoreEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept cross-region restore with explicit GeoRedundant storage", func() {
				input := validResource()
				mode := "GeoRedundant"
				input.Spec.StorageModeType = &mode
				input.Spec.CrossRegionRestoreEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every immutability posture", func() {
				for _, state := range []string{"Locked", "Unlocked", "Disabled"} {
					input := validResource()
					input.Spec.Immutability = state
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureRecoveryServicesVaultIdentity{
					Type: AzureRecoveryServicesVaultIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept CMK encryption with the system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureRecoveryServicesVaultIdentity{
					Type: AzureRecoveryServicesVaultIdentityType_SYSTEM_ASSIGNED,
				}
				input.Spec.Encryption = &AzureRecoveryServicesVaultEncryption{
					KeyId: literal(testKeyVersionlessId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept CMK encryption unwrapped by a user-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureRecoveryServicesVaultIdentity{
					Type:        AzureRecoveryServicesVaultIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				input.Spec.Encryption = &AzureRecoveryServicesVaultEncryption{
					KeyId:                     literal(testKeyVersionlessId),
					UseSystemAssignedIdentity: boolPtr(false),
					UserAssignedIdentityId:    literal(testIdentityId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a monitoring block turning a v5-only switch off", func() {
				input := validResource()
				input.Spec.Monitoring = &AzureRecoveryServicesVaultMonitoring{
					AlertsForAllJobFailuresEnabled:           boolPtr(true),
					EmailNotificationsForSiteRecoveryEnabled: boolPtr(false),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a resource guard association", func() {
				input := validResource()
				input.Spec.ResourceGuardId = literal(testResourceGuardId)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_recovery_services_vault", func() {

			ginkgo.It("should reject a vault name that starts with a digit", func() {
				input := validResource()
				input.Spec.Name = "1vault"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a one-character vault name", func() {
				input := validResource()
				input.Spec.Name = "v"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a vault name longer than 50 characters", func() {
				input := validResource()
				input.Spec.Name = "v-3456789012345678901234567890123456789012345678901"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown sku", func() {
				input := validResource()
				sku := "Premium"
				input.Spec.Sku = &sku
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown storage mode", func() {
				input := validResource()
				mode := "ReadAccessGeoRedundant"
				input.Spec.StorageModeType = &mode
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject cross-region restore on locally redundant storage", func() {
				input := validResource()
				mode := "LocallyRedundant"
				input.Spec.StorageModeType = &mode
				input.Spec.CrossRegionRestoreEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown immutability state", func() {
				input := validResource()
				input.Spec.Immutability = "Frozen"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject encryption without the identity block", func() {
				input := validResource()
				input.Spec.Encryption = &AzureRecoveryServicesVaultEncryption{
					KeyId: literal(testKeyVersionlessId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject encryption without a key", func() {
				input := validResource()
				input.Spec.Identity = &AzureRecoveryServicesVaultIdentity{
					Type: AzureRecoveryServicesVaultIdentityType_SYSTEM_ASSIGNED,
				}
				input.Spec.Encryption = &AzureRecoveryServicesVaultEncryption{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned encryption identity while the system identity stays on", func() {
				input := validResource()
				input.Spec.Identity = &AzureRecoveryServicesVaultIdentity{
					Type:        AzureRecoveryServicesVaultIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				input.Spec.Encryption = &AzureRecoveryServicesVaultEncryption{
					KeyId:                  literal(testKeyVersionlessId),
					UserAssignedIdentityId: literal(testIdentityId),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject turning the system encryption identity off without a user-assigned one", func() {
				input := validResource()
				input.Spec.Identity = &AzureRecoveryServicesVaultIdentity{
					Type: AzureRecoveryServicesVaultIdentityType_SYSTEM_ASSIGNED,
				}
				input.Spec.Encryption = &AzureRecoveryServicesVaultEncryption{
					KeyId:                     literal(testKeyVersionlessId),
					UseSystemAssignedIdentity: boolPtr(false),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity flavor without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureRecoveryServicesVaultIdentity{
					Type: AzureRecoveryServicesVaultIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject identity ids on a system-assigned-only identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureRecoveryServicesVaultIdentity{
					Type:        AzureRecoveryServicesVaultIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing resource group", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
