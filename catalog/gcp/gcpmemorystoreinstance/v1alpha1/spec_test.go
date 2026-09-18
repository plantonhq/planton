package gcpmemorystoreinstancev1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
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
			ApiVersion: "gcp.planton.dev/v1alpha1",
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
		deletionProtection := true
		msg.Spec.DeletionProtectionEnabled = &deletionProtection
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept user labels", func() {
		msg := minimal()
		msg.Spec.Labels = map[string]string{"team": "platform", "tier": "cache"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a DR PRIMARY listing its secondaries", func() {
		msg := minimal()
		msg.Spec.CrossInstanceReplicationConfig = &GcpMemorystoreInstanceCrossInstanceReplicationConfig{
			InstanceRole: "PRIMARY",
			SecondaryInstances: []*GcpMemorystoreInstanceSecondaryInstance{
				{Instance: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "projects/my-gcp-project/locations/us-east1/instances/my-cache-dr",
					},
				}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a DR SECONDARY pointing at its primary", func() {
		msg := minimal()
		msg.Spec.CrossInstanceReplicationConfig = &GcpMemorystoreInstanceCrossInstanceReplicationConfig{
			InstanceRole: "SECONDARY",
			PrimaryInstance: &GcpMemorystoreInstancePrimaryInstance{
				Instance: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "projects/my-gcp-project/locations/us-central1/instances/my-cache-01",
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept role NONE with no replication references", func() {
		msg := minimal()
		msg.Spec.CrossInstanceReplicationConfig = &GcpMemorystoreInstanceCrossInstanceReplicationConfig{
			InstanceRole: "NONE",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept seeding from GCS RDB files", func() {
		msg := minimal()
		msg.Spec.GcsSource = &GcpMemorystoreInstanceGcsSource{
			Uris: []string{"gs://my-bucket/dumps/cache.rdb"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept seeding from a managed backup", func() {
		msg := minimal()
		msg.Spec.ManagedBackupSource = &GcpMemorystoreInstanceManagedBackupSource{
			Backup: "projects/my-gcp-project/locations/us-central1/backupCollections/col-1/backups/b-1",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a SECONDARY without a primary_instance reference", func() {
		msg := minimal()
		msg.Spec.CrossInstanceReplicationConfig = &GcpMemorystoreInstanceCrossInstanceReplicationConfig{
			InstanceRole: "SECONDARY",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("primary"))
	})

	ginkgo.It("should reject a primary_instance reference on a PRIMARY", func() {
		msg := minimal()
		msg.Spec.CrossInstanceReplicationConfig = &GcpMemorystoreInstanceCrossInstanceReplicationConfig{
			InstanceRole: "PRIMARY",
			PrimaryInstance: &GcpMemorystoreInstancePrimaryInstance{
				Instance: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/p/locations/l/instances/i"},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject secondary_instances on a SECONDARY", func() {
		msg := minimal()
		msg.Spec.CrossInstanceReplicationConfig = &GcpMemorystoreInstanceCrossInstanceReplicationConfig{
			InstanceRole: "SECONDARY",
			PrimaryInstance: &GcpMemorystoreInstancePrimaryInstance{
				Instance: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/p/locations/l/instances/i"},
				},
			},
			SecondaryInstances: []*GcpMemorystoreInstanceSecondaryInstance{
				{Instance: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/p/locations/l2/instances/i2"},
				}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid replication instance_role", func() {
		msg := minimal()
		msg.Spec.CrossInstanceReplicationConfig = &GcpMemorystoreInstanceCrossInstanceReplicationConfig{
			InstanceRole: "REPLICA",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject both seed sources set together", func() {
		msg := minimal()
		msg.Spec.GcsSource = &GcpMemorystoreInstanceGcsSource{
			Uris: []string{"gs://my-bucket/dumps/cache.rdb"},
		}
		msg.Spec.ManagedBackupSource = &GcpMemorystoreInstanceManagedBackupSource{
			Backup: "projects/p/locations/l/backupCollections/c/backups/b",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("mutually exclusive"))
	})

	ginkgo.It("should reject a non-gs:// seed URI", func() {
		msg := minimal()
		msg.Spec.GcsSource = &GcpMemorystoreInstanceGcsSource{
			Uris: []string{"s3://wrong-cloud/dump.rdb"},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty seed URI list", func() {
		msg := minimal()
		msg.Spec.GcsSource = &GcpMemorystoreInstanceGcsSource{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a zone where a region is expected", func() {
		msg := minimal()
		msg.Spec.Location = "us-central1-a"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an omitted project_id (ambient-project contract)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
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

	ginkgo.It("should accept a PSC connection without project_id (rides the effective project)", func() {
		msg := minimal()
		msg.Spec.PscAutoConnections = []*GcpMemorystoreInstancePscAutoConnection{
			{
				Network: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "projects/p/global/networks/vpc"},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Maintenance Window, CA Mode, Deletion Policy ────────────────
	// The window has no minute field by API truth: Memorystore rejects a
	// start time carrying minutes (live 400 "Invalid start time, only
	// hours are supported"), so the spec models the start hour-only.

	ginkgo.It("should accept a maintenance window on the hour", func() {
		msg := minimal()
		msg.Spec.MaintenancePolicy = &GcpMemorystoreInstanceMaintenancePolicy{
			WeeklyMaintenanceWindow: &GcpMemorystoreInstanceMaintenanceWindow{
				Day:  "SUNDAY",
				Hour: 3,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each documented node_type value", func() {
		for _, nodeType := range []string{"", "SHARED_CORE_NANO", "CUSTOM_PICO", "CUSTOM_MICRO", "CUSTOM_MINI", "STANDARD_SMALL", "STANDARD_LARGE", "HIGHCPU_MEDIUM", "HIGHMEM_MEDIUM", "HIGHMEM_XLARGE", "HIGHMEM_2XLARGE"} {
			msg := minimal()
			msg.Spec.NodeType = nodeType
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "node_type %q should be accepted", nodeType)
		}
	})

	ginkgo.It("should reject an unknown node_type value", func() {
		msg := minimal()
		msg.Spec.NodeType = "MEGA_LARGE"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each valid server_ca_mode value", func() {
		for _, mode := range []string{"", "GOOGLE_MANAGED_PER_INSTANCE_CA", "GOOGLE_MANAGED_SHARED_CA", "CUSTOMER_MANAGED_CAS_CA"} {
			msg := minimal()
			msg.Spec.ServerCaMode = mode
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "server_ca_mode %q should be accepted", mode)
		}
	})

	ginkgo.It("should reject an invalid server_ca_mode value", func() {
		msg := minimal()
		msg.Spec.ServerCaMode = "SELF_SIGNED"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept server_ca_pool paired with CUSTOMER_MANAGED_CAS_CA", func() {
		msg := minimal()
		msg.Spec.ServerCaMode = "CUSTOMER_MANAGED_CAS_CA"
		msg.Spec.ServerCaPool = "projects/my-project/locations/us-central1/caPools/my-pool"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject server_ca_pool without CUSTOMER_MANAGED_CAS_CA mode", func() {
		msg := minimal()
		msg.Spec.ServerCaPool = "projects/my-project/locations/us-central1/caPools/my-pool"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a maintenance_version", func() {
		msg := minimal()
		msg.Spec.MaintenanceVersion = "20260801_00_00"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each valid deletion_policy value", func() {
		for _, policy := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = policy
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "deletion_policy %q should be accepted", policy)
		}
	})

	ginkgo.It("should reject an invalid deletion_policy value", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a full-resource-name acl_policy", func() {
		msg := minimal()
		msg.Spec.AclPolicy = "projects/my-gcp-project/locations/us-central1/aclPolicies/readers"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an acl_policy that is not a full resource name", func() {
		msg := minimal()
		msg.Spec.AclPolicy = "readers"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("acl_policy"))
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
