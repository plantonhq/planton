package azurebackupprotectedvmv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureBackupProtectedVmSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureBackupProtectedVmSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testVmId     = "/subscriptions/s/resourceGroups/vm-rg/providers/Microsoft.Compute/virtualMachines/app-vm"
	testPolicyId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.RecoveryServices/vaults/backup-vault/backupPolicies/daily-policy"
)

// validResource returns a minimal valid protection binding that
// individual cases mutate into the shape under test.
func validResource() *AzureBackupProtectedVm {
	return &AzureBackupProtectedVm{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureBackupProtectedVm",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-backup-protected-vm",
		},
		Spec: &AzureBackupProtectedVmSpec{
			ResourceGroup:     literal("backup-rg"),
			RecoveryVaultName: literal("backup-vault"),
			SourceVmId:        literal(testVmId),
			BackupPolicyId:    literal(testPolicyId),
		},
	}
}

var _ = ginkgo.Describe("AzureBackupProtectedVmSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_backup_protected_vm", func() {

			ginkgo.It("should not return a validation error for a minimal protection binding", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept excluded disk LUNs", func() {
				input := validResource()
				input.Spec.ExcludeDiskLuns = []int32{1, 2}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept included disk LUNs", func() {
				input := validResource()
				input.Spec.IncludeDiskLuns = []int32{0}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every protection posture", func() {
				for _, state := range []string{"Protected", "BackupsSuspended", "ProtectionStopped"} {
					input := validResource()
					input.Spec.ProtectionState = state
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_backup_protected_vm", func() {

			ginkgo.It("should reject a missing source VM", func() {
				input := validResource()
				input.Spec.SourceVmId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing backup policy", func() {
				input := validResource()
				input.Spec.BackupPolicyId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing vault name", func() {
				input := validResource()
				input.Spec.RecoveryVaultName = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject mixing include and exclude disk LUNs", func() {
				input := validResource()
				input.Spec.ExcludeDiskLuns = []int32{1}
				input.Spec.IncludeDiskLuns = []int32{0}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a negative disk LUN", func() {
				input := validResource()
				input.Spec.ExcludeDiskLuns = []int32{-1}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown protection posture", func() {
				input := validResource()
				input.Spec.ProtectionState = "Paused"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
