package gcpmemorystoreinstancev1

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
	ginkgo.RunSpecs(t, "GcpMemorystoreInstanceSpec Suite")
}

var _ = ginkgo.Describe("GcpMemorystoreInstanceSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpMemorystoreInstance.
	minimal := func() *GcpMemorystoreInstance {
		return &GcpMemorystoreInstance{
			ApiVersion: "gcp.planton.dev/v1",
			Kind:       "GcpMemorystoreInstance",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-memorystore",
			},
			Spec: &GcpMemorystoreInstanceSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				InstanceName: "my-cache-01",
				Location:     "us-central1",
				ShardCount:   1,
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept instance_name at minimum boundary (4 chars)", func() {
		msg := minimal()
		msg.Spec.InstanceName = "ab0c"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept instance_name at maximum boundary (63 chars)", func() {
		msg := minimal()
		msg.Spec.InstanceName = "a" + strings.Repeat("b", 61) + "c"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept instance_name with hyphens and numbers", func() {
		msg := minimal()
		msg.Spec.InstanceName = "cache-prod-01-us"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept mode CLUSTER", func() {
		msg := minimal()
		msg.Spec.Mode = "CLUSTER"
		msg.Spec.ShardCount = 3
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept mode CLUSTER_DISABLED", func() {
		msg := minimal()
		msg.Spec.Mode = "CLUSTER_DISABLED"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty mode (defaults to GCP default)", func() {
		msg := minimal()
		msg.Spec.Mode = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept all valid node types", func() {
		for _, nodeType := range []string{"SHARED_CORE_NANO", "STANDARD_SMALL", "HIGHMEM_MEDIUM", "HIGHMEM_XLARGE"} {
			msg := minimal()
			msg.Spec.NodeType = nodeType
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should accept empty node_type (defaults to GCP default)", func() {
		msg := minimal()
		msg.Spec.NodeType = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept engine_version", func() {
		msg := minimal()
		msg.Spec.EngineVersion = "VALKEY_8_0"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept engine_configs", func() {
		msg := minimal()
		msg.Spec.EngineConfigs = map[string]string{
			"maxmemory-policy": "volatile-ttl",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept PSC auto connections", func() {
		msg := minimal()
		msg.Spec.PscAutoConnections = []*GcpMemorystoreInstancePscAutoConnection{
			{
				Network: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "projects/my-project/global/networks/my-vpc",
					},
				},
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-project",
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept multiple PSC connections (multi-VPC)", func() {
		msg := minimal()
		msg.Spec.PscAutoConnections = []*GcpMemorystoreInstancePscAutoConnection{
			{
				Network: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/p1/global/networks/vpc1"},
				},
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "p1"},
				},
			},
			{
				Network: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/p2/global/networks/vpc2"},
				},
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "p2"},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept authorization_mode AUTH_DISABLED", func() {
		msg := minimal()
		msg.Spec.AuthorizationMode = "AUTH_DISABLED"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept authorization_mode IAM_AUTH", func() {
		msg := minimal()
		msg.Spec.AuthorizationMode = "IAM_AUTH"
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
		msg.Spec.PersistenceConfig = &GcpMemorystoreInstancePersistenceConfig{
			Mode: "DISABLED",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept persistence_config with RDB mode and all snapshot periods", func() {
		for _, period := range []string{"ONE_HOUR", "SIX_HOURS", "TWELVE_HOURS", "TWENTY_FOUR_HOURS"} {
			msg := minimal()
			msg.Spec.PersistenceConfig = &GcpMemorystoreInstancePersistenceConfig{
				Mode: "RDB",
				RdbConfig: &GcpMemorystoreInstanceRdbConfig{
					RdbSnapshotPeriod: period,
				},
			}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should accept persistence_config with AOF mode and all fsync options", func() {
		for _, fsync := range []string{"NEVER", "EVERY_SEC", "ALWAYS"} {
			msg := minimal()
			msg.Spec.PersistenceConfig = &GcpMemorystoreInstancePersistenceConfig{
				Mode: "AOF",
				AofConfig: &GcpMemorystoreInstanceAofConfig{
					AppendFsync: fsync,
				},
			}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should accept zone_distribution_config MULTI_ZONE", func() {
		msg := minimal()
		msg.Spec.ZoneDistributionConfig = &GcpMemorystoreInstanceZoneDistributionConfig{
			Mode: "MULTI_ZONE",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept zone_distribution_config SINGLE_ZONE with zone", func() {
		msg := minimal()
		msg.Spec.ZoneDistributionConfig = &GcpMemorystoreInstanceZoneDistributionConfig{
			Mode: "SINGLE_ZONE",
			Zone: "us-central1-a",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept all valid maintenance window days", func() {
		days := []string{"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"}
		for _, day := range days {
			msg := minimal()
			msg.Spec.MaintenancePolicy = &GcpMemorystoreInstanceMaintenancePolicy{
				WeeklyMaintenanceWindow: &GcpMemorystoreInstanceMaintenanceWindow{
					Day:  day,
					Hour: 0,
				},
			}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		}
	})

	ginkgo.It("should accept maintenance_window with hour at boundary (23)", func() {
		msg := minimal()
		msg.Spec.MaintenancePolicy = &GcpMemorystoreInstanceMaintenancePolicy{
			WeeklyMaintenanceWindow: &GcpMemorystoreInstanceMaintenanceWindow{
				Day:  "MONDAY",
				Hour: 23,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept automated_backup_config with valid retention", func() {
		msg := minimal()
		msg.Spec.AutomatedBackupConfig = &GcpMemorystoreInstanceAutomatedBackupConfig{
			StartHour: 2,
			Retention: "3024000s",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept full spec with all optional fields set", func() {
		msg := minimal()
		msg.Spec.Mode = "CLUSTER"
		msg.Spec.ShardCount = 3
		msg.Spec.NodeType = "HIGHMEM_MEDIUM"
		msg.Spec.EngineVersion = "VALKEY_8_0"
		msg.Spec.EngineConfigs = map[string]string{"maxmemory-policy": "allkeys-lru"}
		msg.Spec.ReplicaCount = 2
		msg.Spec.PscAutoConnections = []*GcpMemorystoreInstancePscAutoConnection{
			{
				Network: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "projects/p/global/networks/vpc",
					},
				},
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "p"},
				},
			},
		}
		msg.Spec.AuthorizationMode = "IAM_AUTH"
		msg.Spec.TransitEncryptionMode = "SERVER_AUTHENTICATION"
		msg.Spec.KmsKey = &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
				Value: "projects/p/locations/us-central1/keyRings/kr/cryptoKeys/k",
			},
		}
		msg.Spec.PersistenceConfig = &GcpMemorystoreInstancePersistenceConfig{
			Mode: "AOF",
			AofConfig: &GcpMemorystoreInstanceAofConfig{
				AppendFsync: "EVERY_SEC",
			},
		}
		msg.Spec.ZoneDistributionConfig = &GcpMemorystoreInstanceZoneDistributionConfig{
			Mode: "MULTI_ZONE",
		}
		msg.Spec.MaintenancePolicy = &GcpMemorystoreInstanceMaintenancePolicy{
			WeeklyMaintenanceWindow: &GcpMemorystoreInstanceMaintenanceWindow{
				Day:  "SUNDAY",
				Hour: 3,
			},
		}
		msg.Spec.AutomatedBackupConfig = &GcpMemorystoreInstanceAutomatedBackupConfig{
			StartHour: 4,
			Retention: "3024000s",
		}
		msg.Spec.DeletionProtectionEnabled = true
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

	ginkgo.It("should reject instance_name shorter than 4 chars", func() {
		msg := minimal()
		msg.Spec.InstanceName = "abc"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name longer than 63 chars", func() {
		msg := minimal()
		msg.Spec.InstanceName = "a" + strings.Repeat("b", 63)
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
		msg.Spec.InstanceName = "MyCache01"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name ending with a hyphen", func() {
		msg := minimal()
		msg.Spec.InstanceName = "my-cache-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_name with underscores", func() {
		msg := minimal()
		msg.Spec.InstanceName = "my_cache_01"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when location is empty", func() {
		msg := minimal()
		msg.Spec.Location = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject when shard_count is zero", func() {
		msg := minimal()
		msg.Spec.ShardCount = 0
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid mode value", func() {
		msg := minimal()
		msg.Spec.Mode = "STANDALONE"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid node_type value", func() {
		msg := minimal()
		msg.Spec.NodeType = "MEGA_LARGE"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid authorization_mode value", func() {
		msg := minimal()
		msg.Spec.AuthorizationMode = "PASSWORD"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid transit_encryption_mode value", func() {
		msg := minimal()
		msg.Spec.TransitEncryptionMode = "MUTUAL_TLS"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid persistence_config mode", func() {
		msg := minimal()
		msg.Spec.PersistenceConfig = &GcpMemorystoreInstancePersistenceConfig{
			Mode: "SNAPSHOT",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject RDB mode without rdb_config", func() {
		msg := minimal()
		msg.Spec.PersistenceConfig = &GcpMemorystoreInstancePersistenceConfig{
			Mode: "RDB",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject AOF mode without aof_config", func() {
		msg := minimal()
		msg.Spec.PersistenceConfig = &GcpMemorystoreInstancePersistenceConfig{
			Mode: "AOF",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid rdb_snapshot_period value", func() {
		msg := minimal()
		msg.Spec.PersistenceConfig = &GcpMemorystoreInstancePersistenceConfig{
			Mode: "RDB",
			RdbConfig: &GcpMemorystoreInstanceRdbConfig{
				RdbSnapshotPeriod: "EVERY_MINUTE",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid aof append_fsync value", func() {
		msg := minimal()
		msg.Spec.PersistenceConfig = &GcpMemorystoreInstancePersistenceConfig{
			Mode: "AOF",
			AofConfig: &GcpMemorystoreInstanceAofConfig{
				AppendFsync: "SOMETIMES",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid zone_distribution_config mode", func() {
		msg := minimal()
		msg.Spec.ZoneDistributionConfig = &GcpMemorystoreInstanceZoneDistributionConfig{
			Mode: "ANY_ZONE",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject SINGLE_ZONE without zone", func() {
		msg := minimal()
		msg.Spec.ZoneDistributionConfig = &GcpMemorystoreInstanceZoneDistributionConfig{
			Mode: "SINGLE_ZONE",
			Zone: "",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid maintenance_window day", func() {
		msg := minimal()
		msg.Spec.MaintenancePolicy = &GcpMemorystoreInstanceMaintenancePolicy{
			WeeklyMaintenanceWindow: &GcpMemorystoreInstanceMaintenanceWindow{
				Day:  "FUNDAY",
				Hour: 10,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject maintenance_window hour > 23", func() {
		msg := minimal()
		msg.Spec.MaintenancePolicy = &GcpMemorystoreInstanceMaintenancePolicy{
			WeeklyMaintenanceWindow: &GcpMemorystoreInstanceMaintenanceWindow{
				Day:  "MONDAY",
				Hour: 24,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid automated_backup_config retention format", func() {
		msg := minimal()
		msg.Spec.AutomatedBackupConfig = &GcpMemorystoreInstanceAutomatedBackupConfig{
			StartHour: 2,
			Retention: "35days",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject PSC connection without network", func() {
		msg := minimal()
		msg.Spec.PscAutoConnections = []*GcpMemorystoreInstancePscAutoConnection{
			{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "p"},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject PSC connection without project_id", func() {
		msg := minimal()
		msg.Spec.PscAutoConnections = []*GcpMemorystoreInstancePscAutoConnection{
			{
				Network: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/p/global/networks/vpc"},
				},
			},
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
