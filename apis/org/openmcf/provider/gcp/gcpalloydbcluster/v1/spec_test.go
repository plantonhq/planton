package gcpalloydbclusterv1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpAlloydbClusterSpec Suite")
}

var _ = ginkgo.Describe("GcpAlloydbClusterSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpAlloydbCluster.
	minimal := func() *GcpAlloydbCluster {
		return &GcpAlloydbCluster{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpAlloydbCluster",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-alloydb",
			},
			Spec: &GcpAlloydbClusterSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				ClusterName: "my-alloydb-cluster",
				Location:    "us-central1",
				Network: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "projects/my-gcp-project/global/networks/default",
					},
				},
				PrimaryInstance: &GcpAlloydbClusterPrimaryInstance{
					InstanceId: "my-primary",
				},
			},
		}
	}

	// Helper for StringValueOrRef with a literal value.
	svr := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a cluster with database_version set", func() {
		msg := minimal()
		msg.Spec.DatabaseVersion = "POSTGRES_15"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept all valid database versions", func() {
		for _, version := range []string{"POSTGRES_14", "POSTGRES_15", "POSTGRES_16"} {
			msg := minimal()
			msg.Spec.DatabaseVersion = version
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "version: %s", version)
		}
	})

	ginkgo.It("should accept a cluster with display_name", func() {
		msg := minimal()
		msg.Spec.DisplayName = "My Production AlloyDB"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a cluster with initial_user", func() {
		msg := minimal()
		msg.Spec.InitialUser = &GcpAlloydbClusterInitialUser{
			Password: "strongpassword123",
			User:     "admin",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept initial_user with default username (empty)", func() {
		msg := minimal()
		msg.Spec.InitialUser = &GcpAlloydbClusterInitialUser{
			Password: "strongpassword123",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a cluster with automated_backup_policy (quantity-based)", func() {
		msg := minimal()
		msg.Spec.AutomatedBackupPolicy = &GcpAlloydbClusterAutomatedBackupPolicy{
			Enabled:                       true,
			BackupWindow:                  "3600s",
			QuantityBasedRetentionCount:   7,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a cluster with automated_backup_policy (time-based)", func() {
		msg := minimal()
		msg.Spec.AutomatedBackupPolicy = &GcpAlloydbClusterAutomatedBackupPolicy{
			Enabled:                    true,
			TimeBasedRetentionPeriod:   "1209600s",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept automated_backup_policy with weekly schedule", func() {
		msg := minimal()
		msg.Spec.AutomatedBackupPolicy = &GcpAlloydbClusterAutomatedBackupPolicy{
			Enabled: true,
			WeeklySchedule: &GcpAlloydbClusterBackupSchedule{
				DaysOfWeek: []string{"MONDAY", "WEDNESDAY", "FRIDAY"},
				StartHour:  2,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept automated_backup_policy with CMEK encryption", func() {
		msg := minimal()
		msg.Spec.AutomatedBackupPolicy = &GcpAlloydbClusterAutomatedBackupPolicy{
			Enabled:             true,
			EncryptionKmsKeyName: svr("projects/p/locations/l/keyRings/kr/cryptoKeys/k"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept continuous_backup_config", func() {
		msg := minimal()
		msg.Spec.ContinuousBackupConfig = &GcpAlloydbClusterContinuousBackupConfig{
			Enabled:            true,
			RecoveryWindowDays: 21,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept continuous_backup_config at boundary (1 day)", func() {
		msg := minimal()
		msg.Spec.ContinuousBackupConfig = &GcpAlloydbClusterContinuousBackupConfig{
			Enabled:            true,
			RecoveryWindowDays: 1,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept continuous_backup_config at boundary (35 days)", func() {
		msg := minimal()
		msg.Spec.ContinuousBackupConfig = &GcpAlloydbClusterContinuousBackupConfig{
			Enabled:            true,
			RecoveryWindowDays: 35,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept cluster-level CMEK encryption", func() {
		msg := minimal()
		msg.Spec.KmsKeyName = svr("projects/p/locations/l/keyRings/kr/cryptoKeys/k")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept maintenance_window", func() {
		msg := minimal()
		msg.Spec.MaintenanceWindow = &GcpAlloydbClusterMaintenanceWindow{
			Day:       "TUESDAY",
			StartHour: 3,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept primary_instance with cpu_count", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.CpuCount = 4
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept primary_instance with machine_type", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.MachineType = "n2-highmem-4"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept primary_instance with REGIONAL availability", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.AvailabilityType = "REGIONAL"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept primary_instance with ZONAL availability", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.AvailabilityType = "ZONAL"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept primary_instance with database_flags", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.DatabaseFlags = map[string]string{
			"max_connections": "200",
			"work_mem":        "64000",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept primary_instance with query_insights_config", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.QueryInsightsConfig = &GcpAlloydbClusterQueryInsightsConfig{
			QueryPlansPerMinute:  10,
			QueryStringLength:    2048,
			RecordApplicationTags: true,
			RecordClientAddress:   true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept primary_instance with ssl_mode ENCRYPTED_ONLY", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.SslMode = "ENCRYPTED_ONLY"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept primary_instance with require_connectors", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.RequireConnectors = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully-featured spec", func() {
		msg := minimal()
		msg.Spec.DatabaseVersion = "POSTGRES_16"
		msg.Spec.DisplayName = "Production AlloyDB"
		msg.Spec.AllocatedIpRange = "my-ip-range"
		msg.Spec.DeletionProtection = true
		msg.Spec.KmsKeyName = svr("projects/p/locations/l/keyRings/kr/cryptoKeys/k")
		msg.Spec.InitialUser = &GcpAlloydbClusterInitialUser{
			Password: "securepassword123",
			User:     "dbadmin",
		}
		msg.Spec.AutomatedBackupPolicy = &GcpAlloydbClusterAutomatedBackupPolicy{
			Enabled:                     true,
			BackupWindow:                "7200s",
			QuantityBasedRetentionCount: 14,
			WeeklySchedule: &GcpAlloydbClusterBackupSchedule{
				DaysOfWeek: []string{"MONDAY", "THURSDAY"},
				StartHour:  3,
			},
			EncryptionKmsKeyName: svr("projects/p/locations/l/keyRings/kr/cryptoKeys/backup-key"),
		}
		msg.Spec.ContinuousBackupConfig = &GcpAlloydbClusterContinuousBackupConfig{
			Enabled:              true,
			RecoveryWindowDays:   21,
			EncryptionKmsKeyName: svr("projects/p/locations/l/keyRings/kr/cryptoKeys/pitr-key"),
		}
		msg.Spec.MaintenanceWindow = &GcpAlloydbClusterMaintenanceWindow{
			Day:       "SUNDAY",
			StartHour: 4,
		}
		msg.Spec.PrimaryInstance = &GcpAlloydbClusterPrimaryInstance{
			InstanceId:       "prod-primary",
			CpuCount:         8,
			AvailabilityType: "REGIONAL",
			DisplayName:      "Production Primary",
			DatabaseFlags: map[string]string{
				"max_connections": "500",
			},
			QueryInsightsConfig: &GcpAlloydbClusterQueryInsightsConfig{
				QueryPlansPerMinute:  10,
				QueryStringLength:    4096,
				RecordApplicationTags: true,
				RecordClientAddress:   true,
			},
			RequireConnectors: true,
			SslMode:           "ENCRYPTED_ONLY",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept cluster_name at minimum length (2 chars)", func() {
		msg := minimal()
		msg.Spec.ClusterName = "ab"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept cluster_name with hyphens", func() {
		msg := minimal()
		msg.Spec.ClusterName = "my-alloydb-cluster-2026"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject missing project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing cluster_name", func() {
		msg := minimal()
		msg.Spec.ClusterName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cluster_name starting with digit", func() {
		msg := minimal()
		msg.Spec.ClusterName = "1-bad-name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cluster_name with uppercase", func() {
		msg := minimal()
		msg.Spec.ClusterName = "MyBadCluster"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cluster_name ending with hyphen", func() {
		msg := minimal()
		msg.Spec.ClusterName = "bad-name-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cluster_name with single char", func() {
		msg := minimal()
		msg.Spec.ClusterName = "a"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing location", func() {
		msg := minimal()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing network", func() {
		msg := minimal()
		msg.Spec.Network = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing primary_instance", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing primary_instance.instance_id", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.InstanceId = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid database_version", func() {
		msg := minimal()
		msg.Spec.DatabaseVersion = "MYSQL_8_0"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("database_version"))
	})

	ginkgo.It("should reject invalid availability_type", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.AvailabilityType = "MULTI_REGIONAL"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("availability_type"))
	})

	ginkgo.It("should reject invalid ssl_mode", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.SslMode = "TLS_REQUIRED"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("ssl_mode"))
	})

	ginkgo.It("should reject both cpu_count and machine_type set", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.CpuCount = 4
		msg.Spec.PrimaryInstance.MachineType = "n2-highmem-4"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("only one of cpu_count or machine_type may be set"))
	})

	ginkgo.It("should reject both quantity_based and time_based retention", func() {
		msg := minimal()
		msg.Spec.AutomatedBackupPolicy = &GcpAlloydbClusterAutomatedBackupPolicy{
			Enabled:                       true,
			QuantityBasedRetentionCount:   7,
			TimeBasedRetentionPeriod:      "1209600s",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("only one of quantity_based_retention_count or time_based_retention_period may be set"))
	})

	ginkgo.It("should reject invalid backup_window format", func() {
		msg := minimal()
		msg.Spec.AutomatedBackupPolicy = &GcpAlloydbClusterAutomatedBackupPolicy{
			Enabled:      true,
			BackupWindow: "1h",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("backup_window"))
	})

	ginkgo.It("should reject invalid time_based_retention_period format", func() {
		msg := minimal()
		msg.Spec.AutomatedBackupPolicy = &GcpAlloydbClusterAutomatedBackupPolicy{
			Enabled:                  true,
			TimeBasedRetentionPeriod: "14d",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("time_based_retention_period"))
	})

	ginkgo.It("should reject recovery_window_days out of range (36)", func() {
		msg := minimal()
		msg.Spec.ContinuousBackupConfig = &GcpAlloydbClusterContinuousBackupConfig{
			Enabled:            true,
			RecoveryWindowDays: 36,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("recovery_window_days"))
	})

	ginkgo.It("should reject recovery_window_days out of range (negative)", func() {
		msg := minimal()
		msg.Spec.ContinuousBackupConfig = &GcpAlloydbClusterContinuousBackupConfig{
			Enabled:            true,
			RecoveryWindowDays: -1,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid days_of_week in weekly_schedule", func() {
		msg := minimal()
		msg.Spec.AutomatedBackupPolicy = &GcpAlloydbClusterAutomatedBackupPolicy{
			Enabled: true,
			WeeklySchedule: &GcpAlloydbClusterBackupSchedule{
				DaysOfWeek: []string{"MONDAY", "FUNDAY"},
				StartHour:  2,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject start_hour out of range (24)", func() {
		msg := minimal()
		msg.Spec.AutomatedBackupPolicy = &GcpAlloydbClusterAutomatedBackupPolicy{
			Enabled: true,
			WeeklySchedule: &GcpAlloydbClusterBackupSchedule{
				DaysOfWeek: []string{"MONDAY"},
				StartHour:  24,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject initial_user password too short", func() {
		msg := minimal()
		msg.Spec.InitialUser = &GcpAlloydbClusterInitialUser{
			Password: "short",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject initial_user without password", func() {
		msg := minimal()
		msg.Spec.InitialUser = &GcpAlloydbClusterInitialUser{
			User: "admin",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid maintenance_window day", func() {
		msg := minimal()
		msg.Spec.MaintenanceWindow = &GcpAlloydbClusterMaintenanceWindow{
			Day:       "FUNDAY",
			StartHour: 3,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject maintenance_window without day", func() {
		msg := minimal()
		msg.Spec.MaintenanceWindow = &GcpAlloydbClusterMaintenanceWindow{
			StartHour: 3,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject query_plans_per_minute out of range (21)", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.QueryInsightsConfig = &GcpAlloydbClusterQueryInsightsConfig{
			QueryPlansPerMinute: 21,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject query_string_length out of range (255)", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.QueryInsightsConfig = &GcpAlloydbClusterQueryInsightsConfig{
			QueryStringLength: 255,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject query_string_length out of range (4501)", func() {
		msg := minimal()
		msg.Spec.PrimaryInstance.QueryInsightsConfig = &GcpAlloydbClusterQueryInsightsConfig{
			QueryStringLength: 4501,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong kind", func() {
		msg := minimal()
		msg.Kind = "WrongKind"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing spec", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	// Ensure strings import is used (for potential future use).
	_ = strings.ToLower
})
