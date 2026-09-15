package awsbackupvaultv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsBackupVaultSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsBackupVaultSpec Validation Suite")
}

func svr(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

func int32Ptr(i int32) *int32 { return &i }

// minimalStandardVault is the smallest valid instance: a standard
// vault with AWS defaults everywhere.
func minimalStandardVault() *AwsBackupVaultSpec {
	return &AwsBackupVaultSpec{
		Region:    "us-west-2",
		VaultType: &AwsBackupVaultSpec_Standard{Standard: &AwsBackupVaultStandard{}},
	}
}

var _ = ginkgo.Describe("AwsBackupVaultSpec validations", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts the minimal standard vault", func() {
			gomega.Expect(protovalidate.Validate(minimalStandardVault())).To(gomega.BeNil())
		})

		ginkgo.It("accepts a KMS-encrypted standard vault with force_destroy", func() {
			spec := minimalStandardVault()
			spec.GetStandard().KmsKeyArn = svr("arn:aws:kms:us-west-2:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab")
			spec.GetStandard().ForceDestroy = true
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a governance-mode lock (no changeable window)", func() {
			spec := minimalStandardVault()
			spec.GetStandard().Lock = &AwsBackupVaultLock{
				MinRetentionDays: int32Ptr(7),
				MaxRetentionDays: int32Ptr(365),
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a compliance-mode lock at the 3-day cooling-off floor", func() {
			spec := minimalStandardVault()
			spec.GetStandard().Lock = &AwsBackupVaultLock{ChangeableForDays: int32Ptr(3)}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts notifications on a standard vault", func() {
			spec := minimalStandardVault()
			spec.GetStandard().Notifications = &AwsBackupVaultNotifications{
				SnsTopicArn: svr("arn:aws:sns:us-west-2:123456789012:backup-events"),
				Events:      []string{"BACKUP_JOB_FAILED", "RESTORE_JOB_FAILED"},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an air-gapped vault at the retention floor", func() {
			spec := &AwsBackupVaultSpec{
				Region: "us-west-2",
				VaultType: &AwsBackupVaultSpec_AirGapped{AirGapped: &AwsBackupVaultAirGapped{
					MinRetentionDays: 7,
					MaxRetentionDays: 7,
				}},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("rejects a missing region", func() {
			spec := minimalStandardVault()
			spec.Region = ""
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a vault with neither arm", func() {
			spec := &AwsBackupVaultSpec{Region: "us-west-2"}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("carries exactly one vault type by construction (setting a second arm replaces the first)", func() {
			spec := minimalStandardVault()
			spec.VaultType = &AwsBackupVaultSpec_AirGapped{AirGapped: &AwsBackupVaultAirGapped{MinRetentionDays: 7, MaxRetentionDays: 30}}
			gomega.Expect(spec.GetStandard()).To(gomega.BeNil())
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("rejects a lock block that constrains nothing", func() {
			spec := minimalStandardVault()
			spec.GetStandard().Lock = &AwsBackupVaultLock{}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("enforces nothing"))
		})

		ginkgo.It("rejects a compliance cooling-off window below 3 days", func() {
			spec := minimalStandardVault()
			spec.GetStandard().Lock = &AwsBackupVaultLock{ChangeableForDays: int32Ptr(2)}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a lock whose max retention is below its min", func() {
			spec := minimalStandardVault()
			spec.GetStandard().Lock = &AwsBackupVaultLock{
				MinRetentionDays: int32Ptr(30),
				MaxRetentionDays: int32Ptr(7),
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects notifications without events", func() {
			spec := minimalStandardVault()
			spec.GetStandard().Notifications = &AwsBackupVaultNotifications{
				SnsTopicArn: svr("arn:aws:sns:us-west-2:123456789012:backup-events"),
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown notification event", func() {
			spec := minimalStandardVault()
			spec.GetStandard().Notifications = &AwsBackupVaultNotifications{
				SnsTopicArn: svr("arn:aws:sns:us-west-2:123456789012:backup-events"),
				Events:      []string{"VAULT_DELETED"},
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an air-gapped vault below the 7-day retention floor", func() {
			spec := &AwsBackupVaultSpec{
				Region:    "us-west-2",
				VaultType: &AwsBackupVaultSpec_AirGapped{AirGapped: &AwsBackupVaultAirGapped{MinRetentionDays: 6, MaxRetentionDays: 30}},
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an air-gapped retention window with max below min", func() {
			spec := &AwsBackupVaultSpec{
				Region:    "us-west-2",
				VaultType: &AwsBackupVaultSpec_AirGapped{AirGapped: &AwsBackupVaultAirGapped{MinRetentionDays: 30, MaxRetentionDays: 7}},
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.BeNil())
		})
	})
})
