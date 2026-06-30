package awsrediselasticachev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestAwsRedisElasticacheSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsRedisElasticacheSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

var _ = ginkgo.Describe("AwsRedisElasticacheSpec validations", func() {
	var spec *AwsRedisElasticacheSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: single-node Redis cluster.
		spec = &AwsRedisElasticacheSpec{
			Region:           "us-west-2",
			Engine:           "redis",
			Description:      "test cluster",
			NodeType:         "cache.t3.micro",
			NumCacheClusters: 1,
		}
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal single-node Redis cluster", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a Valkey engine", func() {
		spec.Engine = "valkey"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a non-clustered HA setup with failover and multi-AZ", func() {
		spec.NumCacheClusters = 3
		spec.AutomaticFailoverEnabled = true
		spec.MultiAzEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a clustered (sharded) topology", func() {
		spec.NumCacheClusters = 0
		spec.NumNodeGroups = 3
		spec.ReplicasPerNodeGroup = 2
		spec.AutomaticFailoverEnabled = true
		spec.MultiAzEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts encryption settings", func() {
		spec.AtRestEncryptionEnabled = true
		spec.TransitEncryptionEnabled = true
		spec.TransitEncryptionMode = "required"
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts auth_token authentication", func() {
		spec.TransitEncryptionEnabled = true
		spec.AuthToken = strRef("my-secret-auth-token-at-least-16-chars")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts user_group_ids authentication", func() {
		spec.UserGroupIds = []string{"my-user-group"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts maintenance and snapshot settings", func() {
		spec.MaintenanceWindow = "sun:05:00-sun:06:00"
		spec.SnapshotRetentionLimit = 7
		spec.SnapshotWindow = "03:00-04:00"
		spec.FinalSnapshotIdentifier = "my-final-snapshot"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts custom parameters", func() {
		spec.ParameterGroupFamily = "redis7"
		spec.Parameters = []*AwsRedisElasticacheParameter{
			{Name: "maxmemory-policy", Value: "volatile-lru"},
			{Name: "timeout", Value: "300"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts log delivery configurations", func() {
		spec.LogDeliveryConfigurations = []*AwsRedisElasticacheLogDeliveryConfig{
			{
				DestinationType: "cloudwatch-logs",
				Destination:     strRef("/aws/elasticache/my-cluster"),
				LogFormat:       "json",
				LogType:         "slow-log",
			},
			{
				DestinationType: "kinesis-firehose",
				Destination:     strRef("my-firehose-stream"),
				LogFormat:       "json",
				LogType:         "engine-log",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a production-ready configuration", func() {
		port := int32(6379)
		spec.EngineVersion = "7.1"
		spec.Port = &port
		spec.NumCacheClusters = 3
		spec.AutomaticFailoverEnabled = true
		spec.MultiAzEnabled = true
		spec.AtRestEncryptionEnabled = true
		spec.TransitEncryptionEnabled = true
		spec.TransitEncryptionMode = "required"
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/mrk-abc123")
		spec.SubnetIds = []*foreignkeyv1.StringValueOrRef{
			strRef("subnet-aaa"), strRef("subnet-bbb"),
		}
		spec.SecurityGroupIds = []*foreignkeyv1.StringValueOrRef{
			strRef("sg-123"),
		}
		spec.MaintenanceWindow = "sun:05:00-sun:06:00"
		spec.SnapshotRetentionLimit = 7
		spec.SnapshotWindow = "03:00-04:00"
		spec.NotificationTopicArn = strRef("arn:aws:sns:us-east-1:123456789012:my-topic")
		spec.ParameterGroupFamily = "redis7"
		spec.Parameters = []*AwsRedisElasticacheParameter{
			{Name: "maxmemory-policy", Value: "volatile-lru"},
		}
		spec.LogDeliveryConfigurations = []*AwsRedisElasticacheLogDeliveryConfig{
			{
				DestinationType: "cloudwatch-logs",
				Destination:     strRef("/aws/elasticache/my-cluster"),
				LogFormat:       "json",
				LogType:         "slow-log",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts data tiering enabled", func() {
		spec.NodeType = "cache.r6gd.xlarge"
		spec.DataTieringEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts notification_topic_arn", func() {
		spec.NotificationTopicArn = strRef("arn:aws:sns:us-east-1:123456789012:my-topic")
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Required field validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when engine is missing", func() {
		spec.Engine = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when description is missing", func() {
		spec.Description = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when node_type is missing", func() {
		spec.NodeType = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: engine valid values
	// -------------------------------------------------------------------------

	ginkgo.It("fails when engine is an invalid value", func() {
		spec.Engine = "memcached"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: topology mode selection
	// -------------------------------------------------------------------------

	ginkgo.It("fails when neither num_cache_clusters nor num_node_groups is set", func() {
		spec.NumCacheClusters = 0
		spec.NumNodeGroups = 0
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when both num_cache_clusters and num_node_groups are set", func() {
		spec.NumCacheClusters = 2
		spec.NumNodeGroups = 3
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: num_cache_clusters range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when num_cache_clusters exceeds 6", func() {
		spec.NumCacheClusters = 7
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: replicas_per_node_group requires num_node_groups
	// -------------------------------------------------------------------------

	ginkgo.It("fails when replicas_per_node_group is set without num_node_groups", func() {
		spec.NumCacheClusters = 1
		spec.ReplicasPerNodeGroup = 2
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level: replicas_per_node_group range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when replicas_per_node_group exceeds 5", func() {
		spec.NumCacheClusters = 0
		spec.NumNodeGroups = 3
		spec.ReplicasPerNodeGroup = 6
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: failover requires multi-node
	// -------------------------------------------------------------------------

	ginkgo.It("fails when automatic_failover_enabled with single node", func() {
		spec.NumCacheClusters = 1
		spec.AutomaticFailoverEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("accepts automatic_failover_enabled with clustered mode", func() {
		spec.NumCacheClusters = 0
		spec.NumNodeGroups = 2
		spec.AutomaticFailoverEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: multi_az requires failover
	// -------------------------------------------------------------------------

	ginkgo.It("fails when multi_az_enabled without automatic_failover_enabled", func() {
		spec.NumCacheClusters = 3
		spec.MultiAzEnabled = true
		spec.AutomaticFailoverEnabled = false
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: auth mutual exclusion
	// -------------------------------------------------------------------------

	ginkgo.It("fails when both auth_token and user_group_ids are set", func() {
		spec.AuthToken = strRef("my-secret-auth-token-at-least-16-chars")
		spec.UserGroupIds = []string{"my-user-group"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: transit_encryption_mode requires transit_encryption_enabled
	// -------------------------------------------------------------------------

	ginkgo.It("fails when transit_encryption_mode is set without transit_encryption_enabled", func() {
		spec.TransitEncryptionMode = "required"
		spec.TransitEncryptionEnabled = false
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: transit_encryption_mode valid values
	// -------------------------------------------------------------------------

	ginkgo.It("fails when transit_encryption_mode has an invalid value", func() {
		spec.TransitEncryptionEnabled = true
		spec.TransitEncryptionMode = "optional"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: parameters require parameter_group_family
	// -------------------------------------------------------------------------

	ginkgo.It("fails when parameters are set without parameter_group_family", func() {
		spec.Parameters = []*AwsRedisElasticacheParameter{
			{Name: "timeout", Value: "300"},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: log delivery max 2
	// -------------------------------------------------------------------------

	ginkgo.It("fails when more than 2 log delivery configurations are set", func() {
		cfg := &AwsRedisElasticacheLogDeliveryConfig{
			DestinationType: "cloudwatch-logs",
			Destination:     strRef("log-group"),
			LogFormat:       "json",
			LogType:         "slow-log",
		}
		spec.LogDeliveryConfigurations = []*AwsRedisElasticacheLogDeliveryConfig{cfg, cfg, cfg}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// CEL: log delivery unique log_type
	// -------------------------------------------------------------------------

	ginkgo.It("fails when two log delivery configs have the same log_type", func() {
		spec.LogDeliveryConfigurations = []*AwsRedisElasticacheLogDeliveryConfig{
			{
				DestinationType: "cloudwatch-logs",
				Destination:     strRef("log-group-1"),
				LogFormat:       "json",
				LogType:         "slow-log",
			},
			{
				DestinationType: "cloudwatch-logs",
				Destination:     strRef("log-group-2"),
				LogFormat:       "text",
				LogType:         "slow-log",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level: maintenance_window pattern
	// -------------------------------------------------------------------------

	ginkgo.It("fails when maintenance_window has an invalid format", func() {
		spec.MaintenanceWindow = "sunday:05:00-sunday:06:00"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level: snapshot_window pattern
	// -------------------------------------------------------------------------

	ginkgo.It("fails when snapshot_window has an invalid format", func() {
		spec.SnapshotWindow = "3am-4am"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level: snapshot_retention_limit range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when snapshot_retention_limit exceeds 35", func() {
		spec.SnapshotRetentionLimit = 36
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Field-level: port range
	// -------------------------------------------------------------------------

	ginkgo.It("fails when port is 0", func() {
		port := int32(0)
		spec.Port = &port
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when port exceeds 65535", func() {
		port := int32(65536)
		spec.Port = &port
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Nested message: log delivery config validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when log delivery destination_type is invalid", func() {
		spec.LogDeliveryConfigurations = []*AwsRedisElasticacheLogDeliveryConfig{
			{
				DestinationType: "s3-bucket",
				Destination:     strRef("my-bucket"),
				LogFormat:       "json",
				LogType:         "slow-log",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when log delivery log_format is invalid", func() {
		spec.LogDeliveryConfigurations = []*AwsRedisElasticacheLogDeliveryConfig{
			{
				DestinationType: "cloudwatch-logs",
				Destination:     strRef("log-group"),
				LogFormat:       "xml",
				LogType:         "slow-log",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when log delivery log_type is invalid", func() {
		spec.LogDeliveryConfigurations = []*AwsRedisElasticacheLogDeliveryConfig{
			{
				DestinationType: "cloudwatch-logs",
				Destination:     strRef("log-group"),
				LogFormat:       "json",
				LogType:         "access-log",
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
