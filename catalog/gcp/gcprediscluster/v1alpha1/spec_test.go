package gcpredisclusterv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpRedisClusterSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

func nameRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{ValueFrom: &foreignkeyv1.ValueFromRef{Name: v}},
	}
}

var _ = ginkgo.Describe("GcpRedisClusterSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// The smallest valid cluster: one shard, Google-placed PSC endpoints,
	// everything else defaulted.
	minimal := func() *GcpRedisCluster {
		return &GcpRedisCluster{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpRedisCluster",
			Metadata:   &shared.CloudResourceMetadata{Name: "orders-cache"},
			Spec: &GcpRedisClusterSpec{
				Region:     "us-central1",
				ShardCount: 1,
				PscConfigs: []*GcpRedisClusterPscConfig{{Network: nameRef("prod-vpc")}},
			},
		}
	}

	ginkgo.It("should accept the minimal cluster", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a cluster without psc_configs (user-created connections)", func() {
		msg := minimal()
		msg.Spec.PscConfigs = nil
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should accept every optional lever together", func() {
		msg := minimal()
		msg.Spec.ProjectId = litRef("cache-project")
		msg.Spec.ClusterName = "orders-cache"
		msg.Spec.ShardCount = 3
		msg.Spec.ReplicaCount = 2
		msg.Spec.NodeType = "REDIS_HIGHMEM_MEDIUM"
		msg.Spec.RedisConfigs = map[string]string{"maxmemory-policy": "allkeys-lru"}
		msg.Spec.AuthorizationMode = "AUTH_MODE_IAM_AUTH"
		msg.Spec.TransitEncryptionMode = "TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION"
		msg.Spec.ServerCaMode = "SERVER_CA_MODE_CUSTOMER_MANAGED_CAS_CA"
		msg.Spec.ServerCaPool = "projects/p/locations/us-central1/caPools/redis-ca"
		msg.Spec.KmsKey = nameRef("cache-key")
		msg.Spec.PersistenceConfig = &GcpRedisClusterPersistenceConfig{
			Mode:      "AOF",
			AofConfig: &GcpRedisClusterAofConfig{AppendFsync: "EVERYSEC"},
		}
		msg.Spec.ZoneDistributionConfig = &GcpRedisClusterZoneDistributionConfig{Mode: "MULTI_ZONE"}
		msg.Spec.MaintenancePolicy = &GcpRedisClusterMaintenancePolicy{
			WeeklyMaintenanceWindow: &GcpRedisClusterMaintenanceWindow{Day: "SUNDAY", Hour: 3},
		}
		msg.Spec.AutomatedBackupConfig = &GcpRedisClusterAutomatedBackupConfig{StartHour: 2, Retention: "3024000s"}
		msg.Spec.CrossClusterReplicationConfig = &GcpRedisClusterCrossClusterReplicationConfig{
			ClusterRole:       "PRIMARY",
			SecondaryClusters: []*GcpRedisClusterSecondaryCluster{{Cluster: nameRef("orders-cache-eu")}},
		}
		msg.Spec.GcsSource = &GcpRedisClusterGcsSource{Uris: []string{"gs://seed/orders.rdb"}}
		msg.Spec.Labels = map[string]string{"team": "orders"}
		msg.Spec.DeletionProtectionEnabled = proto.Bool(false)
		msg.Spec.MaintenanceVersion = "20260901_00_00"
		msg.Spec.AclPolicy = "projects/p/locations/us-central1/aclPolicies/readers"
		msg.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require region and shard_count", func() {
		msg := minimal()
		msg.Spec.Region = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.ShardCount = 0
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should bound shard_count and replica_count", func() {
		msg := minimal()
		msg.Spec.ShardCount = 251
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = minimal()
		msg.Spec.ReplicaCount = 6
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a cluster_name that is not RFC 1035", func() {
		msg := minimal()
		msg.Spec.ClusterName = "Orders_Cache"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject unknown enum values", func() {
		for _, mutate := range []func(*GcpRedisClusterSpec){
			func(s *GcpRedisClusterSpec) { s.NodeType = "REDIS_GIANT" },
			func(s *GcpRedisClusterSpec) { s.AuthorizationMode = "IAM_AUTH" },
			func(s *GcpRedisClusterSpec) { s.TransitEncryptionMode = "SERVER_AUTHENTICATION" },
			func(s *GcpRedisClusterSpec) { s.ServerCaMode = "CUSTOMER_MANAGED_CAS_CA" },
			func(s *GcpRedisClusterSpec) { s.DeletionPolicy = "KEEP" },
		} {
			msg := minimal()
			mutate(msg.Spec)
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		}
	})

	ginkgo.It("should allow at most one psc_configs entry", func() {
		msg := minimal()
		msg.Spec.PscConfigs = append(msg.Spec.PscConfigs, &GcpRedisClusterPscConfig{Network: nameRef("other-vpc")})
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject two seed sources", func() {
		msg := minimal()
		msg.Spec.GcsSource = &GcpRedisClusterGcsSource{Uris: []string{"gs://seed/a.rdb"}}
		msg.Spec.ManagedBackupSource = &GcpRedisClusterManagedBackupSource{Backup: "projects/p/locations/us-central1/backupCollections/c/backups/b"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a GCS seed URI that is not gs://", func() {
		msg := minimal()
		msg.Spec.GcsSource = &GcpRedisClusterGcsSource{Uris: []string{"s3://seed/a.rdb"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should tie server_ca_pool to the customer-managed CA mode and TLS", func() {
		msg := minimal()
		msg.Spec.ServerCaPool = "projects/p/locations/us-central1/caPools/redis-ca"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg = minimal()
		msg.Spec.ServerCaMode = "SERVER_CA_MODE_GOOGLE_MANAGED_SHARED_CA"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a CA mode on a plaintext cluster")

		msg.Spec.TransitEncryptionMode = "TRANSIT_ENCRYPTION_MODE_SERVER_AUTHENTICATION"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())

		msg.Spec.ServerCaPool = "not-a-pool"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a malformed pool name")
	})

	ginkgo.It("should tie rdb_config and aof_config to their modes", func() {
		msg := minimal()
		msg.Spec.PersistenceConfig = &GcpRedisClusterPersistenceConfig{
			Mode:      "RDB",
			AofConfig: &GcpRedisClusterAofConfig{AppendFsync: "ALWAYS"},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg.Spec.PersistenceConfig = &GcpRedisClusterPersistenceConfig{
			Mode:      "RDB",
			RdbConfig: &GcpRedisClusterRdbConfig{RdbSnapshotPeriod: "SIX_HOURS"},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())

		msg.Spec.PersistenceConfig.RdbConfig.RdbSnapshotPeriod = "HOURLY"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())

		msg.Spec.PersistenceConfig = &GcpRedisClusterPersistenceConfig{
			Mode:      "AOF",
			AofConfig: &GcpRedisClusterAofConfig{AppendFsync: "EVERY_SEC"},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "the Valkey spelling is not Redis Cluster's")
	})

	ginkgo.It("should tie zone to SINGLE_ZONE", func() {
		msg := minimal()
		msg.Spec.ZoneDistributionConfig = &GcpRedisClusterZoneDistributionConfig{Mode: "SINGLE_ZONE"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.ZoneDistributionConfig.Zone = "us-central1-a"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		msg.Spec.ZoneDistributionConfig.Mode = "MULTI_ZONE"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should bound the maintenance window hour and require a valid day", func() {
		msg := minimal()
		msg.Spec.MaintenancePolicy = &GcpRedisClusterMaintenancePolicy{
			WeeklyMaintenanceWindow: &GcpRedisClusterMaintenanceWindow{Day: "SUNDAY", Hour: 24},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.MaintenancePolicy.WeeklyMaintenanceWindow = &GcpRedisClusterMaintenanceWindow{Day: "SUN", Hour: 3}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg.Spec.MaintenancePolicy.WeeklyMaintenanceWindow = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a policy without a window")
	})

	ginkgo.It("should require a seconds duration for backup retention", func() {
		msg := minimal()
		msg.Spec.AutomatedBackupConfig = &GcpRedisClusterAutomatedBackupConfig{StartHour: 2, Retention: "35d"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should enforce the cross-cluster role rules", func() {
		msg := minimal()
		msg.Spec.CrossClusterReplicationConfig = &GcpRedisClusterCrossClusterReplicationConfig{ClusterRole: "SECONDARY"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a SECONDARY without a primary")

		msg.Spec.CrossClusterReplicationConfig.PrimaryCluster = &GcpRedisClusterPrimaryCluster{Cluster: nameRef("orders-cache-us")}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())

		msg.Spec.CrossClusterReplicationConfig.ClusterRole = "PRIMARY"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "a PRIMARY naming a primary")

		msg.Spec.CrossClusterReplicationConfig = &GcpRedisClusterCrossClusterReplicationConfig{
			ClusterRole:       "NONE",
			SecondaryClusters: []*GcpRedisClusterSecondaryCluster{{Cluster: nameRef("x")}},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed(), "secondaries on a NONE role")
	})

	ginkgo.It("should require the ACL policy as a full resource name", func() {
		msg := minimal()
		msg.Spec.AclPolicy = "readers"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
