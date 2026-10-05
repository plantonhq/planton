package azurebackuppolicyfilesharev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureBackupPolicyFileShareSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureBackupPolicyFileShareSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func strPtr(v string) *string { return &v }
func int32Ptr(v int32) *int32 { return &v }

// validResource returns a minimal valid daily snapshot-tier policy
// that individual cases mutate into the shape under test.
func validResource() *AzureBackupPolicyFileShare {
	return &AzureBackupPolicyFileShare{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureBackupPolicyFileShare",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-backup-policy-file-share",
		},
		Spec: &AzureBackupPolicyFileShareSpec{
			ResourceGroup:     literal("backup-rg"),
			RecoveryVaultName: literal("backup-vault"),
			Name:              "daily-share-policy",
			Backup: &AzureBackupPolicyFileShareSchedule{
				Frequency: "Daily",
				Time:      "23:00",
			},
			RetentionDaily: &AzureBackupPolicyFileShareRetentionDaily{Count: 30},
		},
	}
}

var _ = ginkgo.Describe("AzureBackupPolicyFileShareSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_backup_policy_file_share", func() {

			ginkgo.It("should not return a validation error for a minimal daily policy", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an hourly policy with the window block", func() {
				input := validResource()
				input.Spec.Backup = &AzureBackupPolicyFileShareSchedule{
					Frequency: "Hourly",
					Hourly: &AzureBackupPolicyFileShareHourlySchedule{
						Interval:       4,
						StartTime:      "06:00",
						WindowDuration: 12,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the vault-standard tier with snapshot retention below the daily count", func() {
				input := validResource()
				input.Spec.BackupTier = strPtr("vault-standard")
				input.Spec.SnapshotRetentionInDays = int32Ptr(5)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the vault-standard tier without snapshot retention", func() {
				input := validResource()
				input.Spec.BackupTier = strPtr("vault-standard")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit snapshot tier", func() {
				input := validResource()
				input.Spec.BackupTier = strPtr("snapshot")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the full retention ladder", func() {
				input := validResource()
				input.Spec.RetentionWeekly = &AzureBackupPolicyFileShareRetentionWeekly{
					Count:    12,
					Weekdays: []string{"Sunday"},
				}
				input.Spec.RetentionMonthly = &AzureBackupPolicyFileShareRetentionMonthly{
					Count:    12,
					Weeks:    []string{"First"},
					Weekdays: []string{"Sunday"},
				}
				input.Spec.RetentionYearly = &AzureBackupPolicyFileShareRetentionYearly{
					Count:    10,
					Months:   []string{"January"},
					Weeks:    []string{"First"},
					Weekdays: []string{"Sunday"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept monthly retention in the month-days form", func() {
				input := validResource()
				input.Spec.RetentionMonthly = &AzureBackupPolicyFileShareRetentionMonthly{
					Count:           12,
					Days:            []int32{1, 15},
					IncludeLastDays: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept yearly retention with include_last_days only", func() {
				input := validResource()
				input.Spec.RetentionYearly = &AzureBackupPolicyFileShareRetentionYearly{
					Count:           5,
					Months:          []string{"December"},
					IncludeLastDays: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a custom timezone and a half-past time", func() {
				input := validResource()
				input.Spec.Timezone = strPtr("Pacific Standard Time")
				input.Spec.Backup.Time = "02:30"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept every hourly interval with a matching window", func() {
				for _, interval := range []int32{4, 6, 8, 12} {
					input := validResource()
					input.Spec.Backup = &AzureBackupPolicyFileShareSchedule{
						Frequency: "Hourly",
						Hourly: &AzureBackupPolicyFileShareHourlySchedule{
							Interval:       interval,
							StartTime:      "20:30",
							WindowDuration: 24,
						},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept the retention bounds at their edges", func() {
				input := validResource()
				input.Spec.RetentionDaily = &AzureBackupPolicyFileShareRetentionDaily{Count: 200}
				input.Spec.RetentionWeekly = &AzureBackupPolicyFileShareRetentionWeekly{
					Count:    200,
					Weekdays: []string{"Saturday"},
				}
				input.Spec.RetentionMonthly = &AzureBackupPolicyFileShareRetentionMonthly{
					Count: 120,
					Days:  []int32{31},
				}
				input.Spec.RetentionYearly = &AzureBackupPolicyFileShareRetentionYearly{
					Count:  10,
					Months: []string{"June"},
					Days:   []int32{1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("schedule shape", func() {

			ginkgo.It("should reject a daily schedule without a time", func() {
				input := validResource()
				input.Spec.Backup = &AzureBackupPolicyFileShareSchedule{Frequency: "Daily"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a daily schedule carrying an hourly block", func() {
				input := validResource()
				input.Spec.Backup = &AzureBackupPolicyFileShareSchedule{
					Frequency: "Daily",
					Time:      "23:00",
					Hourly: &AzureBackupPolicyFileShareHourlySchedule{
						Interval:       4,
						StartTime:      "06:00",
						WindowDuration: 12,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an hourly schedule without the hourly block", func() {
				input := validResource()
				input.Spec.Backup = &AzureBackupPolicyFileShareSchedule{Frequency: "Hourly"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an hourly schedule carrying a daily time", func() {
				input := validResource()
				input.Spec.Backup = &AzureBackupPolicyFileShareSchedule{
					Frequency: "Hourly",
					Time:      "23:00",
					Hourly: &AzureBackupPolicyFileShareHourlySchedule{
						Interval:       4,
						StartTime:      "06:00",
						WindowDuration: 12,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a Weekly frequency (file shares have no weekly schedule)", func() {
				input := validResource()
				input.Spec.Backup.Frequency = "Weekly"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a time off the half-hour grid", func() {
				input := validResource()
				input.Spec.Backup.Time = "23:15"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an hourly interval outside the provider's set", func() {
				input := validResource()
				input.Spec.Backup = &AzureBackupPolicyFileShareSchedule{
					Frequency: "Hourly",
					Hourly: &AzureBackupPolicyFileShareHourlySchedule{
						Interval:       5,
						StartTime:      "06:00",
						WindowDuration: 12,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a window duration above 24 hours", func() {
				input := validResource()
				input.Spec.Backup = &AzureBackupPolicyFileShareSchedule{
					Frequency: "Hourly",
					Hourly: &AzureBackupPolicyFileShareHourlySchedule{
						Interval:       4,
						StartTime:      "06:00",
						WindowDuration: 25,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("backup tier", func() {

			ginkgo.It("should reject snapshot retention on the snapshot tier", func() {
				input := validResource()
				input.Spec.SnapshotRetentionInDays = int32Ptr(5)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject snapshot retention equal to the daily count", func() {
				input := validResource()
				input.Spec.BackupTier = strPtr("vault-standard")
				input.Spec.SnapshotRetentionInDays = int32Ptr(30)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject snapshot retention above the daily count", func() {
				input := validResource()
				input.Spec.BackupTier = strPtr("vault-standard")
				input.Spec.SnapshotRetentionInDays = int32Ptr(45)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown backup tier", func() {
				input := validResource()
				input.Spec.BackupTier = strPtr("archive")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("retention", func() {

			ginkgo.It("should reject a policy without daily retention", func() {
				input := validResource()
				input.Spec.RetentionDaily = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a daily count above 200", func() {
				input := validResource()
				input.Spec.RetentionDaily = &AzureBackupPolicyFileShareRetentionDaily{Count: 201}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a monthly count above 120", func() {
				input := validResource()
				input.Spec.RetentionMonthly = &AzureBackupPolicyFileShareRetentionMonthly{
					Count: 121,
					Days:  []int32{1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a yearly count above 10", func() {
				input := validResource()
				input.Spec.RetentionYearly = &AzureBackupPolicyFileShareRetentionYearly{
					Count:  11,
					Months: []string{"January"},
					Days:   []int32{1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject weekly retention without weekdays", func() {
				input := validResource()
				input.Spec.RetentionWeekly = &AzureBackupPolicyFileShareRetentionWeekly{Count: 12}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject monthly retention mixing both selection forms", func() {
				input := validResource()
				input.Spec.RetentionMonthly = &AzureBackupPolicyFileShareRetentionMonthly{
					Count:    12,
					Weeks:    []string{"First"},
					Weekdays: []string{"Sunday"},
					Days:     []int32{1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject monthly retention with neither selection form", func() {
				input := validResource()
				input.Spec.RetentionMonthly = &AzureBackupPolicyFileShareRetentionMonthly{Count: 12}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject monthly weeks without weekdays", func() {
				input := validResource()
				input.Spec.RetentionMonthly = &AzureBackupPolicyFileShareRetentionMonthly{
					Count: 12,
					Weeks: []string{"First"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject yearly retention without months", func() {
				input := validResource()
				input.Spec.RetentionYearly = &AzureBackupPolicyFileShareRetentionYearly{
					Count: 5,
					Days:  []int32{1},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject yearly retention mixing both selection forms", func() {
				input := validResource()
				input.Spec.RetentionYearly = &AzureBackupPolicyFileShareRetentionYearly{
					Count:           5,
					Months:          []string{"January"},
					Weeks:           []string{"Last"},
					Weekdays:        []string{"Friday"},
					IncludeLastDays: true,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("identity", func() {

			ginkgo.It("should reject a policy name that does not start with a letter", func() {
				input := validResource()
				input.Spec.Name = "1daily-policy"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a policy name shorter than 3 characters", func() {
				input := validResource()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

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
		})
	})
})
