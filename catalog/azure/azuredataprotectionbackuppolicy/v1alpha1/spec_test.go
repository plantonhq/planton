package azuredataprotectionbackuppolicyv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureDataProtectionBackupPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureDataProtectionBackupPolicySpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func int32Ptr(v int32) *int32 { return &v }

const testVaultId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/backup-rg/providers/Microsoft.DataProtection/backupVaults/app-backup-vault"

// diskVariant returns a minimal valid disk-variant policy -- the
// simplest of the six shapes -- that cases mutate under test.
func diskVariant() *AzureDataProtectionBackupPolicy {
	return &AzureDataProtectionBackupPolicy{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureDataProtectionBackupPolicy",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-backup-policy",
		},
		Spec: &AzureDataProtectionBackupPolicySpec{
			VaultId: literal(testVaultId),
			Name:    "daily-disk-policy",
			Disk: &AzureDataProtectionBackupPolicyDisk{
				BackupRepeatingTimeIntervals: []string{"R/2024-01-01T02:00:00+00:00/P1D"},
				DefaultRetentionDuration:     "P7D",
			},
		},
	}
}

// kubernetesVariant returns a minimal valid AKS-variant policy.
func kubernetesVariant() *AzureDataProtectionBackupPolicy {
	input := diskVariant()
	input.Spec.Disk = nil
	input.Spec.KubernetesCluster = &AzureDataProtectionBackupPolicyKubernetesCluster{
		BackupRepeatingTimeIntervals: []string{"R/2024-01-01T00:00:00+00:00/PT4H"},
		DefaultRetentionRule: &AzureDataProtectionBackupPolicyKubernetesClusterDefaultRetentionRule{
			LifeCycles: []*AzureDataProtectionBackupPolicyKubernetesClusterLifeCycle{
				{DataStoreType: "OperationalStore", Duration: "P14D"},
			},
		},
	}
	return input
}

var _ = ginkgo.Describe("AzureDataProtectionBackupPolicySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_data_protection_backup_policy", func() {

			ginkgo.It("should not return a validation error for a minimal disk policy", func() {
				err := protovalidate.Validate(diskVariant())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a disk policy with an absolute retention rule", func() {
				input := diskVariant()
				input.Spec.Disk.RetentionRules = []*AzureDataProtectionBackupPolicyDiskRetentionRule{
					{
						Name:     "weekly",
						Duration: "P90D",
						Criteria: &AzureDataProtectionBackupPolicyDiskCriteria{AbsoluteCriteria: "FirstOfWeek"},
						Priority: int32Ptr(25),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an operational-only blob policy", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.BlobStorage = &AzureDataProtectionBackupPolicyBlobStorage{
					OperationalDefaultRetentionDuration: "P30D",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a dual-tier blob policy with a vault retention rule", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.BlobStorage = &AzureDataProtectionBackupPolicyBlobStorage{
					OperationalDefaultRetentionDuration: "P30D",
					VaultDefaultRetentionDuration:       "P90D",
					BackupRepeatingTimeIntervals:        []string{"R/2024-01-01T00:00:00+00:00/P1D"},
					TimeZone:                            "UTC",
					RetentionRules: []*AzureDataProtectionBackupPolicyBlobStorageRetentionRule{
						{
							Name: "monthly",
							Criteria: &AzureDataProtectionBackupPolicyBlobStorageCriteria{
								AbsoluteCriteria: "FirstOfMonth",
							},
							LifeCycle: &AzureDataProtectionBackupPolicyBlobStorageLifeCycle{
								DataStoreType: "VaultStore",
								Duration:      "P12M",
							},
							Priority: int32Ptr(20),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept blob calendar criteria including the last-day-of-month encoding", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.BlobStorage = &AzureDataProtectionBackupPolicyBlobStorage{
					VaultDefaultRetentionDuration: "P90D",
					BackupRepeatingTimeIntervals:  []string{"R/2024-01-01T00:00:00+00:00/P1D"},
					RetentionRules: []*AzureDataProtectionBackupPolicyBlobStorageRetentionRule{
						{
							Name: "month-end",
							Criteria: &AzureDataProtectionBackupPolicyBlobStorageCriteria{
								DaysOfMonth:  []int32{0, 1, 28},
								WeeksOfMonth: []string{"Last"},
								MonthsOfYear: []string{"December"},
							},
							LifeCycle: &AzureDataProtectionBackupPolicyBlobStorageLifeCycle{
								DataStoreType: "VaultStore",
								Duration:      "P1Y",
							},
							Priority: int32Ptr(15),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a minimal kubernetes policy", func() {
				err := protovalidate.Validate(kubernetesVariant())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a kubernetes policy with a calendar retention rule", func() {
				input := kubernetesVariant()
				input.Spec.KubernetesCluster.RetentionRules = []*AzureDataProtectionBackupPolicyKubernetesClusterRetentionRule{
					{
						Name: "sundays",
						Criteria: &AzureDataProtectionBackupPolicyKubernetesClusterCriteria{
							DaysOfWeek:           []string{"Sunday"},
							ScheduledBackupTimes: []string{"2024-01-01T00:00:00Z"},
						},
						LifeCycles: []*AzureDataProtectionBackupPolicyKubernetesClusterLifeCycle{
							{DataStoreType: "OperationalStore", Duration: "P8W"},
						},
						Priority: int32Ptr(20),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a minimal mysql flexible-server policy with a valid time zone", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.MysqlFlexibleServer = &AzureDataProtectionBackupPolicyMysqlFlexibleServer{
					BackupRepeatingTimeIntervals: []string{"R/2024-01-01T00:00:00+00:00/P1W"},
					DefaultRetentionRule: &AzureDataProtectionBackupPolicyFlexibleServerDefaultRetentionRule{
						LifeCycles: []*AzureDataProtectionBackupPolicyFlexibleServerLifeCycle{
							{DataStoreType: "VaultStore", Duration: "P3M"},
						},
					},
					TimeZone: "India Standard Time",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a minimal postgresql flexible-server policy", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.PostgresqlFlexibleServer = &AzureDataProtectionBackupPolicyPostgresqlFlexibleServer{
					BackupRepeatingTimeIntervals: []string{"R/2024-01-01T00:00:00+00:00/P1W"},
					DefaultRetentionRule: &AzureDataProtectionBackupPolicyFlexibleServerDefaultRetentionRule{
						LifeCycles: []*AzureDataProtectionBackupPolicyFlexibleServerLifeCycle{
							{DataStoreType: "VaultStore", Duration: "P3M"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a data-lake policy with a days-of-week rule", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.DataLakeStorage = &AzureDataProtectionBackupPolicyDataLakeStorage{
					BackupSchedule:           []string{"R/2024-01-01T00:00:00+00:00/P1D"},
					DefaultRetentionDuration: "P30D",
					TimeZone:                 "Coordinated Universal Time",
					RetentionRules: []*AzureDataProtectionBackupPolicyDataLakeStorageRetentionRule{
						{
							Name:       "sundays",
							Duration:   "P12W",
							DaysOfWeek: []string{"Sunday"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_data_protection_backup_policy", func() {

			ginkgo.It("should reject a policy with no variant", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a policy with two variants", func() {
				input := diskVariant()
				input.Spec.BlobStorage = &AzureDataProtectionBackupPolicyBlobStorage{
					OperationalDefaultRetentionDuration: "P30D",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a two-character name", func() {
				input := diskVariant()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a name with invalid characters", func() {
				input := diskVariant()
				input.Spec.Name = "policy_with_underscores"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a data-lake policy whose name starts with a digit", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.Name = "1-adls-policy"
				input.Spec.DataLakeStorage = &AzureDataProtectionBackupPolicyDataLakeStorage{
					BackupSchedule:           []string{"R/2024-01-01T00:00:00+00:00/P1D"},
					DefaultRetentionDuration: "P30D",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a blob policy with neither tier configured", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.BlobStorage = &AzureDataProtectionBackupPolicyBlobStorage{}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a blob vault tier without schedule intervals", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.BlobStorage = &AzureDataProtectionBackupPolicyBlobStorage{
					VaultDefaultRetentionDuration: "P90D",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject blob retention rules without the vault tier", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.BlobStorage = &AzureDataProtectionBackupPolicyBlobStorage{
					OperationalDefaultRetentionDuration: "P30D",
					RetentionRules: []*AzureDataProtectionBackupPolicyBlobStorageRetentionRule{
						{
							Name:     "orphan",
							Criteria: &AzureDataProtectionBackupPolicyBlobStorageCriteria{AbsoluteCriteria: "FirstOfDay"},
							LifeCycle: &AzureDataProtectionBackupPolicyBlobStorageLifeCycle{
								DataStoreType: "VaultStore",
								Duration:      "P12W",
							},
							Priority: int32Ptr(20),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a malformed ISO-8601 duration", func() {
				input := diskVariant()
				input.Spec.Disk.DefaultRetentionDuration = "7days"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a malformed repeating interval", func() {
				input := diskVariant()
				input.Spec.Disk.BackupRepeatingTimeIntervals = []string{"2024-01-01T02:00:00+00:00/P1D"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a disk policy without schedule intervals", func() {
				input := diskVariant()
				input.Spec.Disk.BackupRepeatingTimeIntervals = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown absolute criteria value", func() {
				input := diskVariant()
				input.Spec.Disk.RetentionRules = []*AzureDataProtectionBackupPolicyDiskRetentionRule{
					{
						Name:     "bad",
						Duration: "P90D",
						Criteria: &AzureDataProtectionBackupPolicyDiskCriteria{AbsoluteCriteria: "LastOfWeek"},
						Priority: int32Ptr(25),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a disk rule without a priority", func() {
				input := diskVariant()
				input.Spec.Disk.RetentionRules = []*AzureDataProtectionBackupPolicyDiskRetentionRule{
					{
						Name:     "no-priority",
						Duration: "P90D",
						Criteria: &AzureDataProtectionBackupPolicyDiskCriteria{AbsoluteCriteria: "FirstOfWeek"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a kubernetes life cycle on the vault store", func() {
				input := kubernetesVariant()
				input.Spec.KubernetesCluster.DefaultRetentionRule.LifeCycles[0].DataStoreType = "VaultStore"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a kubernetes policy without a default retention rule", func() {
				input := kubernetesVariant()
				input.Spec.KubernetesCluster.DefaultRetentionRule = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a flexible-server life cycle on the operational store", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.MysqlFlexibleServer = &AzureDataProtectionBackupPolicyMysqlFlexibleServer{
					BackupRepeatingTimeIntervals: []string{"R/2024-01-01T00:00:00+00:00/P1W"},
					DefaultRetentionRule: &AzureDataProtectionBackupPolicyFlexibleServerDefaultRetentionRule{
						LifeCycles: []*AzureDataProtectionBackupPolicyFlexibleServerLifeCycle{
							{DataStoreType: "OperationalStore", Duration: "P3M"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown mysql time zone", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.MysqlFlexibleServer = &AzureDataProtectionBackupPolicyMysqlFlexibleServer{
					BackupRepeatingTimeIntervals: []string{"R/2024-01-01T00:00:00+00:00/P1W"},
					DefaultRetentionRule: &AzureDataProtectionBackupPolicyFlexibleServerDefaultRetentionRule{
						LifeCycles: []*AzureDataProtectionBackupPolicyFlexibleServerLifeCycle{
							{DataStoreType: "VaultStore", Duration: "P3M"},
						},
					},
					TimeZone: "Asia/Kolkata",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject more than five data-lake schedules", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.DataLakeStorage = &AzureDataProtectionBackupPolicyDataLakeStorage{
					BackupSchedule: []string{
						"R/2024-01-01T00:00:00+00:00/P1D",
						"R/2024-01-01T04:00:00+00:00/P1D",
						"R/2024-01-01T08:00:00+00:00/P1D",
						"R/2024-01-01T12:00:00+00:00/P1D",
						"R/2024-01-01T16:00:00+00:00/P1D",
						"R/2024-01-01T20:00:00+00:00/P1D",
					},
					DefaultRetentionDuration: "P30D",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a data-lake rule with neither absolute criteria nor days of week", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.DataLakeStorage = &AzureDataProtectionBackupPolicyDataLakeStorage{
					BackupSchedule:           []string{"R/2024-01-01T00:00:00+00:00/P1D"},
					DefaultRetentionDuration: "P30D",
					RetentionRules: []*AzureDataProtectionBackupPolicyDataLakeStorageRetentionRule{
						{
							Name:         "months-only",
							Duration:     "P12W",
							MonthsOfYear: []string{"January"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a day of month above 28", func() {
				input := diskVariant()
				input.Spec.Disk = nil
				input.Spec.BlobStorage = &AzureDataProtectionBackupPolicyBlobStorage{
					VaultDefaultRetentionDuration: "P90D",
					BackupRepeatingTimeIntervals:  []string{"R/2024-01-01T00:00:00+00:00/P1D"},
					RetentionRules: []*AzureDataProtectionBackupPolicyBlobStorageRetentionRule{
						{
							Name: "bad-day",
							Criteria: &AzureDataProtectionBackupPolicyBlobStorageCriteria{
								DaysOfMonth: []int32{29},
							},
							LifeCycle: &AzureDataProtectionBackupPolicyBlobStorageLifeCycle{
								DataStoreType: "VaultStore",
								Duration:      "P12W",
							},
							Priority: int32Ptr(20),
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown day of week", func() {
				input := kubernetesVariant()
				input.Spec.KubernetesCluster.RetentionRules = []*AzureDataProtectionBackupPolicyKubernetesClusterRetentionRule{
					{
						Name: "bad-day",
						Criteria: &AzureDataProtectionBackupPolicyKubernetesClusterCriteria{
							DaysOfWeek: []string{"Sundays"},
						},
						LifeCycles: []*AzureDataProtectionBackupPolicyKubernetesClusterLifeCycle{
							{DataStoreType: "OperationalStore", Duration: "P8W"},
						},
						Priority: int32Ptr(20),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing vault reference", func() {
				input := diskVariant()
				input.Spec.VaultId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
