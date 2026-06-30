package gcpredisinstancev1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpRedisInstanceSpec Suite")
}

var _ = ginkgo.Describe("GcpRedisInstanceSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpRedisInstance.
	minimal := func() *GcpRedisInstance {
		return &GcpRedisInstance{
			ApiVersion: "gcp.planton.dev/v1",
			Kind:       "GcpRedisInstance",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-redis",
			},
			Spec: &GcpRedisInstanceSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				InstanceName: "my-redis-cache",
				Region:       "us-central1",
				Tier:         "BASIC",
				MemorySizeGb: 1,
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid BASIC tier spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a minimal valid STANDARD_HA tier spec", func() {
		msg := minimal()
		msg.Spec.Tier = "STANDARD_HA"
		msg.Spec.MemorySizeGb = 5
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept instance_name at minimum boundary (2 chars)", func() {
		msg := minimal()
		msg.Spec.InstanceName = "ab"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept instance_name at maximum boundary (40 chars)", func() {
		msg := minimal()
		msg.Spec.InstanceName = "a" + strings.Repeat("b", 38) + "c"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept instance_name with hyphens and numbers", func() {
		msg := minimal()
		msg.Spec.InstanceName = "cache-prod-01"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with all optional fields set", func() {
		msg := minimal()
		msg.Spec.Tier = "STANDARD_HA"
		msg.Spec.MemorySizeGb = 10
		msg.Spec.RedisVersion = "REDIS_7_0"
		msg.Spec.DisplayName = "Production Cache"
		msg.Spec.LocationId = "us-central1-a"
		msg.Spec.AuthorizedNetwork = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/my-project/global/networks/my-vpc",
			},
		}
		msg.Spec.ConnectMode = "DIRECT_PEERING"
		msg.Spec.ReservedIpRange = "10.0.0.0/29"
		msg.Spec.AuthEnabled = true
		msg.Spec.TransitEncryptionMode = "SERVER_AUTHENTICATION"
		msg.Spec.RedisConfigs = map[string]string{
			"maxmemory-policy": "allkeys-lru",
		}
		msg.Spec.MaintenanceWindow = &GcpRedisInstanceMaintenanceWindow{
			Day:  "SUNDAY",
			Hour: 3,
		}
		msg.Spec.ReadReplicasMode = "READ_REPLICAS_ENABLED"
		msg.Spec.ReplicaCount = 3
		msg.Spec.PersistenceConfig = &GcpRedisInstancePersistenceConfig{
			PersistenceMode:   "RDB",
			RdbSnapshotPeriod: "TWELVE_HOURS",
		}
		msg.Spec.CustomerManagedKey = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/p/locations/us-central1/keyRings/kr/cryptoKeys/k",
			},
		}
		msg.Spec.DeletionProtection = true
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept connect_mode PRIVATE_SERVICE_ACCESS", func() {
		msg := minimal()
		msg.Spec.ConnectMode = "PRIVATE_SERVICE_ACCESS"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept transit_encryption_mode SERVER_AUTHENTICATION", func() {
		msg := minimal()
		msg.Spec.TransitEncryptionMode = "SERVER_AUTHENTICATION"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept persistence_config with DISABLED mode", func() {
		msg := minimal()
		msg.Spec.PersistenceConfig = &GcpRedisInstancePersistenceConfig{
			PersistenceMode: "DISABLED",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept persistence_config with RDB mode and all snapshot periods", func() {
		for _, period := range []string{"ONE_HOUR", "SIX_HOURS", "TWELVE_HOURS", "TWENTY_FOUR_HOURS"} {
			msg := minimal()
			msg.Spec.PersistenceConfig = &GcpRedisInstancePersistenceConfig{
				PersistenceMode:   "RDB",
				RdbSnapshotPeriod: period,
			}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should accept all valid maintenance window days", func() {
		days := []string{"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"}
		for _, day := range days {
			msg := minimal()
			msg.Spec.MaintenanceWindow = &GcpRedisInstanceMaintenanceWindow{
				Day:  day,
				Hour: 0,
			}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should accept maintenance_window with hour at boundary (23)", func() {
		msg := minimal()
		msg.Spec.MaintenanceWindow = &GcpRedisInstanceMaintenanceWindow{
			Day:  "MONDAY",
			Hour: 23,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept STANDARD_HA with read replicas enabled and replica_count 1", func() {
		msg := minimal()
		msg.Spec.Tier = "STANDARD_HA"
		msg.Spec.ReadReplicasMode = "READ_REPLICAS_ENABLED"
		msg.Spec.ReplicaCount = 1
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept STANDARD_HA with read replicas enabled and replica_count 5", func() {
		msg := minimal()
		msg.Spec.Tier = "STANDARD_HA"
		msg.Spec.ReadReplicasMode = "READ_REPLICAS_ENABLED"
		msg.Spec.ReplicaCount = 5
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty connect_mode (defaults to DIRECT_PEERING)", func() {
		msg := minimal()
		msg.Spec.ConnectMode = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty transit_encryption_mode (defaults to DISABLED)", func() {
		msg := minimal()
		msg.Spec.TransitEncryptionMode = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty read_replicas_mode (defaults to DISABLED)", func() {
		msg := minimal()
		msg.Spec.ReadReplicasMode = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject when project_id is missing", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when instance_name is empty", func() {
		msg := minimal()
		msg.Spec.InstanceName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name starting with a digit", func() {
		msg := minimal()
		msg.Spec.InstanceName = "1my-cache"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name starting with a hyphen", func() {
		msg := minimal()
		msg.Spec.InstanceName = "-my-cache"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name with uppercase letters", func() {
		msg := minimal()
		msg.Spec.InstanceName = "MyCache"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name ending with a hyphen", func() {
		msg := minimal()
		msg.Spec.InstanceName = "my-cache-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name longer than 40 characters", func() {
		msg := minimal()
		msg.Spec.InstanceName = "a" + strings.Repeat("b", 40)
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name with underscores", func() {
		msg := minimal()
		msg.Spec.InstanceName = "my_cache"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when region is empty", func() {
		msg := minimal()
		msg.Spec.Region = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when tier is empty", func() {
		msg := minimal()
		msg.Spec.Tier = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid tier value", func() {
		msg := minimal()
		msg.Spec.Tier = "PREMIUM"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when memory_size_gb is zero", func() {
		msg := minimal()
		msg.Spec.MemorySizeGb = 0
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid connect_mode value", func() {
		msg := minimal()
		msg.Spec.ConnectMode = "PUBLIC"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid transit_encryption_mode value", func() {
		msg := minimal()
		msg.Spec.TransitEncryptionMode = "MUTUAL_TLS"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid read_replicas_mode value", func() {
		msg := minimal()
		msg.Spec.ReadReplicasMode = "AUTO"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject read_replicas_mode ENABLED with BASIC tier", func() {
		msg := minimal()
		msg.Spec.Tier = "BASIC"
		msg.Spec.ReadReplicasMode = "READ_REPLICAS_ENABLED"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject replica_count with BASIC tier", func() {
		msg := minimal()
		msg.Spec.Tier = "BASIC"
		msg.Spec.ReplicaCount = 2
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject replica_count > 5 with STANDARD_HA and read replicas", func() {
		msg := minimal()
		msg.Spec.Tier = "STANDARD_HA"
		msg.Spec.ReadReplicasMode = "READ_REPLICAS_ENABLED"
		msg.Spec.ReplicaCount = 6
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject replica_count without read_replicas_mode ENABLED", func() {
		msg := minimal()
		msg.Spec.Tier = "STANDARD_HA"
		msg.Spec.ReadReplicasMode = "READ_REPLICAS_DISABLED"
		msg.Spec.ReplicaCount = 2
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid persistence_mode", func() {
		msg := minimal()
		msg.Spec.PersistenceConfig = &GcpRedisInstancePersistenceConfig{
			PersistenceMode: "AOF",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject RDB mode without rdb_snapshot_period", func() {
		msg := minimal()
		msg.Spec.PersistenceConfig = &GcpRedisInstancePersistenceConfig{
			PersistenceMode: "RDB",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid rdb_snapshot_period value", func() {
		msg := minimal()
		msg.Spec.PersistenceConfig = &GcpRedisInstancePersistenceConfig{
			PersistenceMode:   "RDB",
			RdbSnapshotPeriod: "EVERY_MINUTE",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid maintenance_window day", func() {
		msg := minimal()
		msg.Spec.MaintenanceWindow = &GcpRedisInstanceMaintenanceWindow{
			Day:  "FUNDAY",
			Hour: 10,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject maintenance_window hour > 23", func() {
		msg := minimal()
		msg.Spec.MaintenanceWindow = &GcpRedisInstanceMaintenanceWindow{
			Day:  "MONDAY",
			Hour: 24,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when metadata is missing", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when spec is missing", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
