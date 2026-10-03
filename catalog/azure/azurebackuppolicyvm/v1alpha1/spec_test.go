package azurebackuppolicyvmv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureBackupPolicyVmSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureBackupPolicyVmSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func strPtr(v string) *string { return &v }
func int32Ptr(v int32) *int32 { return &v }

// validResource returns a minimal valid daily policy that individual
// cases mutate into the shape under test.
func validResource() *AzureBackupPolicyVm {
	return &AzureBackupPolicyVm{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureBackupPolicyVm",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-backup-policy-vm",
		},
		Spec: &AzureBackupPolicyVmSpec{
			ResourceGroup:     literal("backup-rg"),
			RecoveryVaultName: literal("backup-vault"),
			Name:              "daily-policy",
			Backup: &AzureBackupPolicyVmSchedule{
				Frequency: "Daily",
				Time:      "23:00",
			},
			RetentionDaily: &AzureBackupPolicyVmRetentionDaily{Count: 7},
		},
	}
}

var _ = ginkgo.Describe("AzureBackupPolicyVmSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_backup_policy_vm", func() {

			ginkgo.It("should not return a validation error for a minimal daily policy", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a single-day daily retention (count 1)", func() {
				input := validResource()
				input.Spec.RetentionDaily = &AzureBackupPolicyVmRetentionDaily{Count: 1}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a weekly policy with weekly retention", func() {
				input := validResource()
				input.Spec.Backup = &AzureBackupPolicyVmSchedule{
					Frequency: "Weekly",
					Time:      "02:30",
					Weekdays:  []string{"Sunday", "Wednesday"},
				}
				input.Spec.RetentionDaily = nil
				input.Spec.RetentionWeekly = &AzureBackupPolicyVmRetentionWeekly{
					Count:    12,
					Weekdays: []string{"Sunday"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an hourly V2 policy", func() {
				input := validResource()
				input.Spec.PolicyType = strPtr("V2")
				// V2 defaults Instant RP to 7; daily count 7 is not strictly greater.
				input.Spec.InstantRestoreRetentionDays = int32Ptr(2)
				input.Spec.Backup = &AzureBackupPolicyVmSchedule{
					Frequency:    "Hourly",
					Time:         "08:00",
					HourInterval: int32Ptr(4),
					HourDuration: int32Ptr(12),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the full grandfather-father-son retention ladder", func() {
				input := validResource()
				input.Spec.RetentionWeekly = &AzureBackupPolicyVmRetentionWeekly{
					Count:    12,
					Weekdays: []string{"Sunday"},
				}
				input.Spec.RetentionMonthly = &AzureBackupPolicyVmRetentionMonthly{
					Count:    12,
					Weeks:    []string{"First"},
					Weekdays: []string{"Sunday"},
				}
				input.Spec.RetentionYearly = &AzureBackupPolicyVmRetentionYearly{
					Count:    7,
					Months:   []string{"January"},
					Weeks:    []string{"First"},
					Weekdays: []string{"Sunday"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept monthly retention in the month-days form", func() {
				input := validResource()
				input.Spec.RetentionMonthly = &AzureBackupPolicyVmRetentionMonthly{
					Count:           12,
					Days:            []int32{1, 15},
					IncludeLastDays: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a V2 crash-consistent policy", func() {
				input := validResource()
				input.Spec.PolicyType = strPtr("V2")
				input.Spec.InstantRestoreRetentionDays = int32Ptr(2)
				input.Spec.ConsistencyType = "OnlyCrashConsistent"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept instant restore retention up to 5 on V1", func() {
				input := validResource()
				input.Spec.InstantRestoreRetentionDays = int32Ptr(5)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept instant restore retention up to 30 on V2", func() {
				input := validResource()
				input.Spec.PolicyType = strPtr("V2")
				input.Spec.InstantRestoreRetentionDays = int32Ptr(30)
				input.Spec.RetentionDaily = &AzureBackupPolicyVmRetentionDaily{Count: 31}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a TierRecommended archive rule", func() {
				input := validResource()
				input.Spec.TieringPolicy = &AzureBackupPolicyVmTieringPolicy{
					ArchivedRestorePoint: &AzureBackupPolicyVmArchivedRestorePoint{
						Mode: "TierRecommended",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a TierAfter archive rule with an age", func() {
				input := validResource()
				input.Spec.TieringPolicy = &AzureBackupPolicyVmTieringPolicy{
					ArchivedRestorePoint: &AzureBackupPolicyVmArchivedRestorePoint{
						Mode:         "TierAfter",
						Duration:     int32Ptr(3),
						DurationType: "Months",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an instant-restore resource group naming rule", func() {
				input := validResource()
				input.Spec.InstantRestoreResourceGroup = &AzureBackupPolicyVmInstantRestoreResourceGroup{
					Prefix: "backup-snapshots",
					Suffix: "prod",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_backup_policy_vm", func() {

			ginkgo.It("should reject a policy name shorter than 3 characters", func() {
				input := validResource()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a policy name starting with a digit", func() {
				input := validResource()
				input.Spec.Name = "1policy"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a backup time off the half hour", func() {
				input := validResource()
				input.Spec.Backup.Time = "23:15"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a daily retention of 2-6 days (Azure's floor)", func() {
				input := validResource()
				input.Spec.RetentionDaily = &AzureBackupPolicyVmRetentionDaily{Count: 5}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a daily policy without daily retention", func() {
				input := validResource()
				input.Spec.RetentionDaily = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a weekly policy keeping daily retention", func() {
				input := validResource()
				input.Spec.Backup = &AzureBackupPolicyVmSchedule{
					Frequency: "Weekly",
					Time:      "23:00",
					Weekdays:  []string{"Sunday"},
				}
				input.Spec.RetentionWeekly = &AzureBackupPolicyVmRetentionWeekly{
					Count:    12,
					Weekdays: []string{"Sunday"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a weekly policy without weekdays", func() {
				input := validResource()
				input.Spec.Backup = &AzureBackupPolicyVmSchedule{
					Frequency: "Weekly",
					Time:      "23:00",
				}
				input.Spec.RetentionDaily = nil
				input.Spec.RetentionWeekly = &AzureBackupPolicyVmRetentionWeekly{
					Count:    12,
					Weekdays: []string{"Sunday"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject weekdays on a daily schedule", func() {
				input := validResource()
				input.Spec.Backup.Weekdays = []string{"Sunday"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an hourly schedule on a V1 policy", func() {
				input := validResource()
				input.Spec.Backup = &AzureBackupPolicyVmSchedule{
					Frequency:    "Hourly",
					Time:         "08:00",
					HourInterval: int32Ptr(4),
					HourDuration: int32Ptr(12),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an hourly schedule without the hourly fields", func() {
				input := validResource()
				input.Spec.PolicyType = strPtr("V2")
				input.Spec.Backup = &AzureBackupPolicyVmSchedule{
					Frequency: "Hourly",
					Time:      "08:00",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an hourly window that is not a multiple of the interval", func() {
				input := validResource()
				input.Spec.PolicyType = strPtr("V2")
				input.Spec.Backup = &AzureBackupPolicyVmSchedule{
					Frequency:    "Hourly",
					Time:         "08:00",
					HourInterval: int32Ptr(6),
					HourDuration: int32Ptr(16),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject hourly fields on a daily schedule", func() {
				input := validResource()
				input.Spec.Backup.HourInterval = int32Ptr(4)
				input.Spec.Backup.HourDuration = int32Ptr(12)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject crash consistency on a V1 policy", func() {
				input := validResource()
				input.Spec.ConsistencyType = "OnlyCrashConsistent"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject instant restore retention above 5 on V1", func() {
				input := validResource()
				input.Spec.InstantRestoreRetentionDays = int32Ptr(6)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a V2 policy with daily count 7 and instant restore unset", func() {
				input := validResource()
				input.Spec.PolicyType = strPtr("V2")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject instant restore that meets or exceeds daily retention", func() {
				input := validResource()
				input.Spec.PolicyType = strPtr("V2")
				input.Spec.InstantRestoreRetentionDays = int32Ptr(7)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject mixing both monthly retention forms", func() {
				input := validResource()
				input.Spec.RetentionMonthly = &AzureBackupPolicyVmRetentionMonthly{
					Count:    12,
					Weeks:    []string{"First"},
					Weekdays: []string{"Sunday"},
					Days:     []int32{1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject monthly retention with no form configured", func() {
				input := validResource()
				input.Spec.RetentionMonthly = &AzureBackupPolicyVmRetentionMonthly{Count: 12}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject monthly weeks without weekdays", func() {
				input := validResource()
				input.Spec.RetentionMonthly = &AzureBackupPolicyVmRetentionMonthly{
					Count: 12,
					Weeks: []string{"First"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject yearly retention without months", func() {
				input := validResource()
				input.Spec.RetentionYearly = &AzureBackupPolicyVmRetentionYearly{
					Count:    7,
					Weeks:    []string{"First"},
					Weekdays: []string{"Sunday"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a TierAfter archive rule without an age", func() {
				input := validResource()
				input.Spec.TieringPolicy = &AzureBackupPolicyVmTieringPolicy{
					ArchivedRestorePoint: &AzureBackupPolicyVmArchivedRestorePoint{
						Mode: "TierAfter",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an age on a TierRecommended archive rule", func() {
				input := validResource()
				input.Spec.TieringPolicy = &AzureBackupPolicyVmTieringPolicy{
					ArchivedRestorePoint: &AzureBackupPolicyVmArchivedRestorePoint{
						Mode:         "TierRecommended",
						Duration:     int32Ptr(3),
						DurationType: "Months",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown weekday", func() {
				input := validResource()
				input.Spec.RetentionWeekly = &AzureBackupPolicyVmRetentionWeekly{
					Count:    12,
					Weekdays: []string{"Funday"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
