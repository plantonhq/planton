package awsrdsclusterv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAwsRdsClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsRdsClusterSpec Custom Validation Tests")
}

// twoSubnets returns the minimum valid subnet list (two literal subnet IDs).
func twoSubnets() []*foreignkeyv1.StringValueOrRef {
	return []*foreignkeyv1.StringValueOrRef{
		{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-12345678"}},
		{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "subnet-87654321"}},
	}
}

// validAuroraCluster returns a minimal valid Aurora Serverless v2 cluster
// that individual tests mutate into specific scenarios.
func validAuroraCluster() *AwsRdsCluster {
	return &AwsRdsCluster{
		ApiVersion: "aws.planton.dev/v1alpha1",
		Kind:       "AwsRdsCluster",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-rds-cluster",
		},
		Spec: &AwsRdsClusterSpec{
			Region:                   "us-west-2",
			SubnetIds:                twoSubnets(),
			Engine:                   "aurora-postgresql",
			MasterUsername:           "dbadmin",
			ManageMasterUserPassword: true,
			SkipFinalSnapshot:        true,
			ServerlessV2Scaling: &AwsRdsClusterServerlessV2Scaling{
				MinCapacity: 0,
				MaxCapacity: 1,
			},
			Instances: []*AwsRdsClusterInstance{
				{Name: "writer", InstanceClass: "db.serverless"},
			},
		},
	}
}

var _ = ginkgo.Describe("AwsRdsClusterSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("accepts a minimal Aurora Serverless v2 cluster", func() {
			err := protovalidate.Validate(validAuroraCluster())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a provisioned Aurora cluster with writer and reader", func() {
			input := validAuroraCluster()
			input.Spec.ServerlessV2Scaling = nil
			input.Spec.Engine = "aurora-mysql"
			input.Spec.EngineVersion = "8.0.mysql_aurora.3.08.0"
			input.Spec.Instances = []*AwsRdsClusterInstance{
				{Name: "writer", InstanceClass: "db.r6g.large", PromotionTier: 0},
				{Name: "reader-1", InstanceClass: "db.r6g.large", PromotionTier: 1},
			}
			input.Spec.EnabledCloudwatchLogsExports = []string{"audit", "error", "slowquery"}
			input.Spec.BacktrackWindowSeconds = 3600
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a Multi-AZ RDS cluster on the postgres engine", func() {
			input := validAuroraCluster()
			input.Spec.Engine = "postgres"
			input.Spec.ServerlessV2Scaling = nil
			input.Spec.Instances = nil
			input.Spec.DbClusterInstanceClass = "db.m6gd.large"
			input.Spec.AllocatedStorageGb = 100
			input.Spec.Iops = 1000
			input.Spec.StorageType = "io1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts an Aurora Serverless v1 cluster with no instances", func() {
			input := validAuroraCluster()
			input.Spec.EngineMode = "serverless"
			input.Spec.ServerlessV2Scaling = nil
			input.Spec.Instances = nil
			input.Spec.ServerlessV1Scaling = &AwsRdsClusterServerlessV1Scaling{
				MinCapacity: 1,
				MaxCapacity: 4,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a direct master password without managed password", func() {
			input := validAuroraCluster()
			input.Spec.ManageMasterUserPassword = false
			input.Spec.MasterPassword = "super-secret-password"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts an existing subnet group instead of subnet ids", func() {
			input := validAuroraCluster()
			input.Spec.SubnetIds = nil
			input.Spec.DbSubnetGroupName = "existing-group"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a point-in-time restore with latest restorable time", func() {
			input := validAuroraCluster()
			input.Spec.RestoreToPointInTime = &AwsRdsClusterRestoreToPointInTime{
				SourceClusterIdentifier: "source-cluster",
				UseLatestRestorableTime: true,
				RestoreType:             "copy-on-write",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts inline cluster parameters with a pinned engine version", func() {
			input := validAuroraCluster()
			input.Spec.EngineVersion = "16.4"
			input.Spec.Parameters = []*AwsRdsClusterParameter{
				{Name: "rds.force_ssl", Value: "1", ApplyMethod: "immediate"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts enhanced monitoring with a role", func() {
			input := validAuroraCluster()
			input.Spec.MonitoringInterval = 30
			input.Spec.MonitoringRoleArn = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/rds-monitoring"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("networking and engine validations", func() {

		ginkgo.It("rejects a spec with neither subnets nor a subnet group", func() {
			input := validAuroraCluster()
			input.Spec.SubnetIds = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("provide at least two subnet_ids (distinct AZs) or an"))
		})

		ginkgo.It("rejects a single subnet without a subnet group", func() {
			input := validAuroraCluster()
			input.Spec.SubnetIds = input.Spec.SubnetIds[:1]
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("provide at least two subnet_ids (distinct AZs) or an"))
		})

		ginkgo.It("rejects an unknown engine", func() {
			input := validAuroraCluster()
			input.Spec.Engine = "mariadb"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid engine_mode", func() {
			input := validAuroraCluster()
			input.Spec.EngineMode = "multimaster"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("engine_mode must be 'provisioned' or 'serverless' when set"))
		})

		ginkgo.It("rejects an invalid network_type", func() {
			input := validAuroraCluster()
			input.Spec.NetworkType = "IPV6"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("network_type must be 'IPV4' or 'DUAL' when set"))
		})

		ginkgo.It("rejects an invalid engine_lifecycle_support value", func() {
			input := validAuroraCluster()
			input.Spec.EngineLifecycleSupport = "extended"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("engine_lifecycle_support must be"))
		})
	})

	ginkgo.Describe("password strategy validations", func() {

		ginkgo.It("rejects a master password alongside the managed password", func() {
			input := validAuroraCluster()
			input.Spec.MasterPassword = "also-a-password"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("master_password cannot be set when"))
		})

		ginkgo.It("rejects a new cluster without a master username", func() {
			input := validAuroraCluster()
			input.Spec.MasterUsername = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("master_username is required"))
		})

		ginkgo.It("accepts a blank master username on a snapshot restore", func() {
			input := validAuroraCluster()
			input.Spec.MasterUsername = ""
			input.Spec.SnapshotIdentifier = "my-snapshot"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("deletion safety validations", func() {

		ginkgo.It("rejects skipping the final snapshot name while not skipping the snapshot", func() {
			input := validAuroraCluster()
			input.Spec.SkipFinalSnapshot = false
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("final_snapshot_identifier is required when"))
		})
	})

	ginkgo.Describe("cluster shape validations", func() {

		ginkgo.It("rejects serverless engine mode on a community engine", func() {
			input := validAuroraCluster()
			input.Spec.Engine = "postgres"
			input.Spec.EngineMode = "serverless"
			input.Spec.ServerlessV2Scaling = nil
			input.Spec.Instances = nil
			input.Spec.DbClusterInstanceClass = "db.m6gd.large"
			input.Spec.AllocatedStorageGb = 100
			input.Spec.Iops = 1000
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("engine_mode 'serverless' (Aurora Serverless v1) requires an"))
		})

		ginkgo.It("rejects v1 scaling outside serverless engine mode", func() {
			input := validAuroraCluster()
			input.Spec.ServerlessV1Scaling = &AwsRdsClusterServerlessV1Scaling{MinCapacity: 1, MaxCapacity: 2}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("serverless_v1_scaling only applies when engine_mode is"))
		})

		ginkgo.It("rejects v2 scaling in serverless engine mode", func() {
			input := validAuroraCluster()
			input.Spec.EngineMode = "serverless"
			input.Spec.Instances = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("serverless_v2_scaling requires provisioned engine mode"))
		})

		ginkgo.It("rejects instances on an Aurora Serverless v1 cluster", func() {
			input := validAuroraCluster()
			input.Spec.EngineMode = "serverless"
			input.Spec.ServerlessV2Scaling = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("instances cannot be set with engine_mode 'serverless' --"))
		})

		ginkgo.It("rejects instances on a Multi-AZ RDS cluster", func() {
			input := validAuroraCluster()
			input.Spec.Engine = "postgres"
			input.Spec.ServerlessV2Scaling = nil
			input.Spec.DbClusterInstanceClass = "db.m6gd.large"
			input.Spec.AllocatedStorageGb = 100
			input.Spec.Iops = 1000
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("instances cannot be set with db_cluster_instance_class -- a"))
		})

		ginkgo.It("rejects db_cluster_instance_class on an Aurora engine", func() {
			input := validAuroraCluster()
			input.Spec.Instances = nil
			input.Spec.DbClusterInstanceClass = "db.m6gd.large"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("db_cluster_instance_class (Multi-AZ RDS cluster) requires"))
		})

		ginkgo.It("rejects a community engine without the Multi-AZ cluster trio", func() {
			input := validAuroraCluster()
			input.Spec.Engine = "postgres"
			input.Spec.ServerlessV2Scaling = nil
			input.Spec.Instances = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("engine 'mysql' or 'postgres' creates a Multi-AZ RDS cluster"))
		})

		ginkgo.It("rejects io1 storage on an Aurora engine", func() {
			input := validAuroraCluster()
			input.Spec.StorageType = "io1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("Aurora engines use storage_type '' (standard) or"))
		})

		ginkgo.It("rejects aurora-iopt1 storage on a Multi-AZ RDS cluster", func() {
			input := validAuroraCluster()
			input.Spec.Engine = "postgres"
			input.Spec.ServerlessV2Scaling = nil
			input.Spec.Instances = nil
			input.Spec.DbClusterInstanceClass = "db.m6gd.large"
			input.Spec.AllocatedStorageGb = 100
			input.Spec.Iops = 1000
			input.Spec.StorageType = "aurora-iopt1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("Multi-AZ RDS clusters (mysql/postgres) use storage_type"))
		})

		ginkgo.It("rejects backtrack on a non-mysql Aurora engine", func() {
			input := validAuroraCluster()
			input.Spec.BacktrackWindowSeconds = 3600
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("backtrack_window_seconds is an Aurora MySQL feature"))
		})
	})

	ginkgo.Describe("log export validations", func() {

		ginkgo.It("rejects mysql log types on a postgresql-family engine", func() {
			input := validAuroraCluster()
			input.Spec.EnabledCloudwatchLogsExports = []string{"slowquery"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("MySQL-family log types are audit/error/general/slowquery"))
		})

		ginkgo.It("rejects postgresql log types on a mysql-family engine", func() {
			input := validAuroraCluster()
			input.Spec.Engine = "aurora-mysql"
			input.Spec.EnabledCloudwatchLogsExports = []string{"postgresql"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("MySQL-family log types are audit/error/general/slowquery"))
		})
	})

	ginkgo.Describe("observability validations", func() {

		ginkgo.It("rejects an invalid performance insights retention", func() {
			input := validAuroraCluster()
			input.Spec.PerformanceInsightsRetentionPeriod = 30
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("performance_insights_retention_period must be 7, 731, or a"))
		})

		ginkgo.It("accepts a monthly-granularity performance insights retention", func() {
			input := validAuroraCluster()
			input.Spec.PerformanceInsightsRetentionPeriod = 93
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid monitoring interval", func() {
			input := validAuroraCluster()
			input.Spec.MonitoringInterval = 45
			input.Spec.MonitoringRoleArn = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/rds-monitoring"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("monitoring_interval must be 0 (disabled), 1, 5, 10, 15, 30,"))
		})

		ginkgo.It("rejects enhanced monitoring without a role", func() {
			input := validAuroraCluster()
			input.Spec.MonitoringInterval = 30
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("monitoring_role_arn is required when monitoring_interval is"))
		})

		ginkgo.It("rejects an invalid database insights mode", func() {
			input := validAuroraCluster()
			input.Spec.DatabaseInsightsMode = "premium"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("database_insights_mode must be 'standard' or 'advanced'"))
		})
	})

	ginkgo.Describe("restore and parameter validations", func() {

		ginkgo.It("rejects snapshot and point-in-time restore together", func() {
			input := validAuroraCluster()
			input.Spec.SnapshotIdentifier = "my-snapshot"
			input.Spec.RestoreToPointInTime = &AwsRdsClusterRestoreToPointInTime{
				SourceClusterIdentifier: "source-cluster",
				UseLatestRestorableTime: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("snapshot_identifier, restore_to_point_in_time, and s3_import"))
		})

		ginkgo.It("rejects inline parameters alongside an existing parameter group", func() {
			input := validAuroraCluster()
			input.Spec.DbClusterParameterGroupName = "existing-group"
			input.Spec.Parameters = []*AwsRdsClusterParameter{{Name: "rds.force_ssl", Value: "1"}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("parameters and db_cluster_parameter_group_name are mutually"))
		})

		ginkgo.It("rejects a restore with both source fields", func() {
			input := validAuroraCluster()
			input.Spec.RestoreToPointInTime = &AwsRdsClusterRestoreToPointInTime{
				SourceClusterIdentifier: "source-cluster",
				SourceClusterResourceId: "cluster-ABC123",
				UseLatestRestorableTime: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one of source_cluster_identifier or"))
		})

		ginkgo.It("rejects a restore with both a time and latest-restorable", func() {
			input := validAuroraCluster()
			input.Spec.RestoreToPointInTime = &AwsRdsClusterRestoreToPointInTime{
				SourceClusterIdentifier: "source-cluster",
				RestoreToTime:           "2026-07-01T09:45:00Z",
				UseLatestRestorableTime: true,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one of restore_to_time or"))
		})

		ginkgo.It("rejects an invalid restore type", func() {
			input := validAuroraCluster()
			input.Spec.RestoreToPointInTime = &AwsRdsClusterRestoreToPointInTime{
				SourceClusterIdentifier: "source-cluster",
				UseLatestRestorableTime: true,
				RestoreType:             "shallow-copy",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("restore_type must be 'full-copy' or 'copy-on-write' when set"))
		})

		ginkgo.It("rejects inline parameters without a pinned engine version", func() {
			input := validAuroraCluster()
			input.Spec.Parameters = []*AwsRdsClusterParameter{
				{Name: "rds.force_ssl", Value: "1"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("inline parameters require a pinned engine_version"))
		})

		ginkgo.It("rejects an invalid parameter apply method", func() {
			input := validAuroraCluster()
			input.Spec.Parameters = []*AwsRdsClusterParameter{
				{Name: "rds.force_ssl", Value: "1", ApplyMethod: "on-restart"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("apply_method must be 'immediate' or 'pending-reboot' when"))
		})
	})

	ginkgo.Describe("instance entry validations", func() {

		ginkgo.It("rejects an instance name with invalid characters", func() {
			input := validAuroraCluster()
			input.Spec.Instances[0].Name = "Writer_1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an instance class without the db. prefix", func() {
			input := validAuroraCluster()
			input.Spec.Instances[0].InstanceClass = "r6g.large"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a promotion tier above 15", func() {
			input := validAuroraCluster()
			input.Spec.Instances[0].PromotionTier = 16
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid per-instance monitoring interval", func() {
			input := validAuroraCluster()
			input.Spec.Instances[0].MonitoringInterval = 20
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("monitoring_interval must be 0 (disabled), 1, 5, 10, 15, 30,"))
		})
	})

	ginkgo.Describe("serverless scaling validations", func() {

		ginkgo.It("rejects v2 max capacity below min capacity", func() {
			input := validAuroraCluster()
			input.Spec.ServerlessV2Scaling = &AwsRdsClusterServerlessV2Scaling{
				MinCapacity: 4,
				MaxCapacity: 2,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_capacity must be greater than or equal to min_capacity"))
		})

		ginkgo.It("rejects v2 auto-pause seconds with a nonzero min capacity", func() {
			input := validAuroraCluster()
			input.Spec.ServerlessV2Scaling = &AwsRdsClusterServerlessV2Scaling{
				MinCapacity:           0.5,
				MaxCapacity:           2,
				SecondsUntilAutoPause: 600,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("seconds_until_auto_pause only applies when min_capacity is"))
		})

		ginkgo.It("rejects v2 auto-pause seconds out of range", func() {
			input := validAuroraCluster()
			input.Spec.ServerlessV2Scaling.SecondsUntilAutoPause = 100
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("seconds_until_auto_pause must be between 300 (5 minutes)"))
		})

		ginkgo.It("rejects a v2 block without max capacity", func() {
			input := validAuroraCluster()
			input.Spec.ServerlessV2Scaling = &AwsRdsClusterServerlessV2Scaling{MinCapacity: 0.5}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects v1 bounds out of order", func() {
			input := validAuroraCluster()
			input.Spec.EngineMode = "serverless"
			input.Spec.ServerlessV2Scaling = nil
			input.Spec.Instances = nil
			input.Spec.ServerlessV1Scaling = &AwsRdsClusterServerlessV1Scaling{
				MinCapacity: 8,
				MaxCapacity: 2,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_capacity must be greater than or equal to min_capacity"))
		})

		ginkgo.It("rejects an invalid v1 timeout action", func() {
			input := validAuroraCluster()
			input.Spec.EngineMode = "serverless"
			input.Spec.ServerlessV2Scaling = nil
			input.Spec.Instances = nil
			input.Spec.ServerlessV1Scaling = &AwsRdsClusterServerlessV1Scaling{
				TimeoutAction: "CancelCapacityChange",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("timeout_action must be 'RollbackCapacityChange' or"))
		})
	})

	ginkgo.Describe("directory join validations", func() {

		ginkgo.It("accepts the managed-directory pair", func() {
			input := validAuroraCluster()
			input.Spec.Domain = "d-1234567890"
			input.Spec.DomainIamRoleName = "rds-directory-service-role"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects a domain without its IAM role name", func() {
			input := validAuroraCluster()
			input.Spec.Domain = "d-1234567890"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("domain and domain_iam_role_name are required together"))
		})
	})

	ginkgo.Describe("s3 import validations", func() {

		validS3Import := func() *AwsRdsClusterS3Import {
			return &AwsRdsClusterS3Import{
				BucketName: "my-backup-bucket",
				IngestionRole: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/rds-s3-import"},
				},
				SourceEngine:        "mysql",
				SourceEngineVersion: "8.0",
			}
		}

		ginkgo.It("accepts an S3 import on aurora-mysql", func() {
			input := validAuroraCluster()
			input.Spec.Engine = "aurora-mysql"
			input.Spec.S3Import = validS3Import()
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects an S3 import on a non-mysql Aurora engine", func() {
			input := validAuroraCluster()
			input.Spec.S3Import = validS3Import()
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("s3_import (Percona XtraBackup restore) is an Aurora MySQL"))
		})

		ginkgo.It("rejects an S3 import alongside a snapshot restore", func() {
			input := validAuroraCluster()
			input.Spec.Engine = "aurora-mysql"
			input.Spec.S3Import = validS3Import()
			input.Spec.SnapshotIdentifier = "my-snapshot"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("snapshot_identifier, restore_to_point_in_time, and s3_import"))
		})

		ginkgo.It("rejects an S3 import with an unsupported source engine", func() {
			input := validAuroraCluster()
			input.Spec.Engine = "aurora-mysql"
			input.Spec.S3Import = validS3Import()
			input.Spec.S3Import.SourceEngine = "postgres"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("iam role association validations", func() {

		ginkgo.It("accepts a role with a feature name and one without", func() {
			input := validAuroraCluster()
			input.Spec.IamRoles = []*AwsRdsClusterIamRole{
				{
					Role: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/s3-import"},
					},
					FeatureName: "s3Import",
				},
				{
					Role: &foreignkeyv1.StringValueOrRef{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:iam::123456789012:role/general"},
					},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects a role entry without the role reference", func() {
			input := validAuroraCluster()
			input.Spec.IamRoles = []*AwsRdsClusterIamRole{{FeatureName: "s3Import"}}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("custom endpoint validations", func() {

		ginkgo.It("accepts a reader endpoint with static members", func() {
			input := validAuroraCluster()
			input.Spec.CustomEndpoints = []*AwsRdsClusterCustomEndpoint{
				{Name: "analytics", Type: "READER", StaticMembers: []string{"writer"}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects static and excluded members together", func() {
			input := validAuroraCluster()
			input.Spec.CustomEndpoints = []*AwsRdsClusterCustomEndpoint{
				{Name: "analytics", Type: "READER", StaticMembers: []string{"writer"}, ExcludedMembers: []string{"reader-1"}},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("static_members and excluded_members are mutually exclusive"))
		})

		ginkgo.It("rejects an invalid endpoint type", func() {
			input := validAuroraCluster()
			input.Spec.CustomEndpoints = []*AwsRdsClusterCustomEndpoint{
				{Name: "analytics", Type: "WRITER"},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("activity stream validations", func() {

		validActivityStream := func() *AwsRdsClusterActivityStream {
			return &AwsRdsClusterActivityStream{
				Mode: "async",
				KmsKeyId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "arn:aws:kms:us-west-2:123456789012:key/abcd"},
				},
			}
		}

		ginkgo.It("accepts an async activity stream on an Aurora engine", func() {
			input := validAuroraCluster()
			input.Spec.ActivityStream = validActivityStream()
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects an activity stream on a Multi-AZ RDS cluster", func() {
			input := validAuroraCluster()
			input.Spec.Engine = "postgres"
			input.Spec.ServerlessV2Scaling = nil
			input.Spec.Instances = nil
			input.Spec.DbClusterInstanceClass = "db.m6gd.large"
			input.Spec.AllocatedStorageGb = 100
			input.Spec.Iops = 1000
			input.Spec.StorageType = "io1"
			input.Spec.ActivityStream = validActivityStream()
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("activity_stream (Database Activity Streams) is an Aurora"))
		})

		ginkgo.It("rejects an invalid activity stream mode", func() {
			input := validAuroraCluster()
			input.Spec.ActivityStream = validActivityStream()
			input.Spec.ActivityStream.Mode = "batch"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("per-instance override validations", func() {

		ginkgo.It("accepts per-instance windows, PI settings, and snapshot tag copy", func() {
			input := validAuroraCluster()
			input.Spec.Instances[0].CopyTagsToSnapshot = true
			input.Spec.Instances[0].PerformanceInsightsRetentionPeriod = 93
			input.Spec.Instances[0].PreferredBackupWindow = "04:00-05:00"
			input.Spec.Instances[0].PreferredMaintenanceWindow = "sun:05:00-sun:06:00"
			input.Spec.Instances[0].ApplyImmediately = true
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid per-instance performance insights retention", func() {
			input := validAuroraCluster()
			input.Spec.Instances[0].PerformanceInsightsRetentionPeriod = 30
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("performance_insights_retention_period must be 7, 731, or a"))
		})

		ginkgo.It("rejects a malformed per-instance backup window", func() {
			input := validAuroraCluster()
			input.Spec.Instances[0].PreferredBackupWindow = "4am-5am"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("final snapshot identifier format", func() {

		ginkgo.It("rejects consecutive hyphens", func() {
			input := validAuroraCluster()
			input.Spec.SkipFinalSnapshot = false
			input.Spec.FinalSnapshotIdentifier = "my--final"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a leading digit", func() {
			input := validAuroraCluster()
			input.Spec.SkipFinalSnapshot = false
			input.Spec.FinalSnapshotIdentifier = "1st-final"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts a well-formed identifier", func() {
			input := validAuroraCluster()
			input.Spec.SkipFinalSnapshot = false
			input.Spec.FinalSnapshotIdentifier = "orders-db-final-2026"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("window format validations", func() {

		ginkgo.It("rejects a malformed backup window", func() {
			input := validAuroraCluster()
			input.Spec.PreferredBackupWindow = "5am-6am"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a malformed maintenance window", func() {
			input := validAuroraCluster()
			input.Spec.PreferredMaintenanceWindow = "monday:03:00-monday:04:00"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a backup retention above 35 days", func() {
			input := validAuroraCluster()
			input.Spec.BackupRetentionPeriod = 36
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
