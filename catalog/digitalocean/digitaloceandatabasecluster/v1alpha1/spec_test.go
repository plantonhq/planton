package digitaloceandatabaseclusterv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	digitalocean "github.com/plantonhq/planton/catalog/digitalocean"
	fk "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestDigitalOceanDatabaseClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanDatabaseClusterSpec Validation Suite")
}

var _ = ginkgo.Describe("DigitalOceanDatabaseClusterSpec validations", func() {

	newVpcRef := func(vpcId string) *fk.StringValueOrRef {
		return &fk.StringValueOrRef{
			LiteralOrRef: &fk.StringValueOrRef_Value{Value: vpcId},
		}
	}

	makeValidPostgresSpec := func() *DigitalOceanDatabaseClusterSpec {
		return &DigitalOceanDatabaseClusterSpec{
			ClusterName:   "test-postgres",
			Engine:        DigitalOceanDatabaseEngine_pg,
			EngineVersion: "16",
			Region:        digitalocean.DigitalOceanRegion_nyc3,
			SizeSlug:      "db-s-1vcpu-1gb",
			NodeCount:     1,
		}
	}

	makeValidMysqlSpec := func() *DigitalOceanDatabaseClusterSpec {
		return &DigitalOceanDatabaseClusterSpec{
			ClusterName:   "test-mysql",
			Engine:        DigitalOceanDatabaseEngine_mysql,
			EngineVersion: "8",
			Region:        digitalocean.DigitalOceanRegion_sfo3,
			SizeSlug:      "db-s-2vcpu-4gb",
			NodeCount:     2,
		}
	}

	ginkgo.Context("Required fields", func() {
		ginkgo.It("accepts a minimal valid PostgreSQL spec", func() {
			spec := makeValidPostgresSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a minimal valid MySQL spec", func() {
			spec := makeValidMysqlSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing cluster_name", func() {
			spec := makeValidPostgresSpec()
			spec.ClusterName = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing engine", func() {
			spec := makeValidPostgresSpec()
			spec.Engine = DigitalOceanDatabaseEngine_digital_ocean_database_engine_unspecified
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing engine_version", func() {
			spec := makeValidPostgresSpec()
			spec.EngineVersion = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing size_slug", func() {
			spec := makeValidPostgresSpec()
			spec.SizeSlug = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing region", func() {
			spec := makeValidPostgresSpec()
			spec.Region = digitalocean.DigitalOceanRegion_digital_ocean_region_unspecified
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("cluster_name validation", func() {
		ginkgo.It("accepts cluster_name with 64 characters (max)", func() {
			spec := makeValidPostgresSpec()
			spec.ClusterName = "a123456789b123456789c123456789d123456789e123456789f123456789abcd" // 64 chars
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects cluster_name exceeding 64 characters", func() {
			spec := makeValidPostgresSpec()
			spec.ClusterName = "a123456789b123456789c123456789d123456789e123456789f123456789abcde" // 65 chars
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("engine coverage", func() {
		ginkgo.It("accepts every DigitalOcean engine slug", func() {
			cases := map[DigitalOceanDatabaseEngine]string{
				DigitalOceanDatabaseEngine_pg:         "16",
				DigitalOceanDatabaseEngine_mysql:      "8",
				DigitalOceanDatabaseEngine_redis:      "7",
				DigitalOceanDatabaseEngine_mongodb:    "7.0",
				DigitalOceanDatabaseEngine_kafka:      "3.5",
				DigitalOceanDatabaseEngine_opensearch: "2",
				DigitalOceanDatabaseEngine_valkey:     "8",
			}
			for engine, version := range cases {
				spec := makeValidPostgresSpec()
				spec.Engine = engine
				spec.EngineVersion = version
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil(), "engine %s", engine)
			}
		})
	})

	ginkgo.Context("engine_version validation", func() {
		ginkgo.It("accepts major version only (e.g. '16')", func() {
			spec := makeValidPostgresSpec()
			spec.EngineVersion = "16"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts major.minor version (e.g. '3.5')", func() {
			spec := makeValidPostgresSpec()
			spec.Engine = DigitalOceanDatabaseEngine_kafka
			spec.EngineVersion = "3.5"
			spec.NodeCount = 3
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects invalid version format (text)", func() {
			spec := makeValidPostgresSpec()
			spec.EngineVersion = "latest"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects version with patch number (e.g. '16.1.2')", func() {
			spec := makeValidPostgresSpec()
			spec.EngineVersion = "16.1.2"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("node_count validation", func() {
		ginkgo.It("accepts node_count = 1", func() {
			spec := makeValidPostgresSpec()
			spec.NodeCount = 1
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts node_count = 3 (Kafka minimum)", func() {
			spec := makeValidPostgresSpec()
			spec.Engine = DigitalOceanDatabaseEngine_kafka
			spec.EngineVersion = "3.5"
			spec.NodeCount = 3
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts node_count = 15 (OpenSearch maximum)", func() {
			spec := makeValidPostgresSpec()
			spec.Engine = DigitalOceanDatabaseEngine_opensearch
			spec.EngineVersion = "2"
			spec.NodeCount = 15
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects node_count = 0", func() {
			spec := makeValidPostgresSpec()
			spec.NodeCount = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("sql_mode engine guard", func() {
		ginkgo.It("accepts sql_mode on MySQL", func() {
			spec := makeValidMysqlSpec()
			spec.SqlMode = "ANSI,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects sql_mode on PostgreSQL", func() {
			spec := makeValidPostgresSpec()
			spec.SqlMode = "ANSI"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects sql_mode on Redis", func() {
			spec := makeValidPostgresSpec()
			spec.Engine = DigitalOceanDatabaseEngine_redis
			spec.EngineVersion = "7"
			spec.SqlMode = "ANSI"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("eviction_policy engine guard", func() {
		ginkgo.It("accepts eviction_policy on Redis", func() {
			spec := makeValidPostgresSpec()
			spec.Engine = DigitalOceanDatabaseEngine_redis
			spec.EngineVersion = "7"
			spec.EvictionPolicy = "allkeys_lru"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts eviction_policy on Valkey", func() {
			spec := makeValidPostgresSpec()
			spec.Engine = DigitalOceanDatabaseEngine_valkey
			spec.EngineVersion = "8"
			spec.EvictionPolicy = "noeviction"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects eviction_policy on PostgreSQL", func() {
			spec := makeValidPostgresSpec()
			spec.EvictionPolicy = "allkeys_lru"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects eviction_policy on MySQL", func() {
			spec := makeValidMysqlSpec()
			spec.EvictionPolicy = "allkeys_lru"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("maintenance_window validation", func() {
		ginkgo.It("accepts a valid window", func() {
			spec := makeValidPostgresSpec()
			spec.MaintenanceWindow = &DigitalOceanDatabaseClusterMaintenanceWindow{
				Day:  "sunday",
				Hour: "02:00",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a case-insensitive day and seconds in hour", func() {
			spec := makeValidPostgresSpec()
			spec.MaintenanceWindow = &DigitalOceanDatabaseClusterMaintenanceWindow{
				Day:  "Monday",
				Hour: "13:00:00",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid day", func() {
			spec := makeValidPostgresSpec()
			spec.MaintenanceWindow = &DigitalOceanDatabaseClusterMaintenanceWindow{
				Day:  "someday",
				Hour: "02:00",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid hour format", func() {
			spec := makeValidPostgresSpec()
			spec.MaintenanceWindow = &DigitalOceanDatabaseClusterMaintenanceWindow{
				Day:  "sunday",
				Hour: "2am",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects a window missing the day", func() {
			spec := makeValidPostgresSpec()
			spec.MaintenanceWindow = &DigitalOceanDatabaseClusterMaintenanceWindow{
				Hour: "02:00",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("backup_restore validation", func() {
		ginkgo.It("accepts a restore with database_name only", func() {
			spec := makeValidPostgresSpec()
			spec.BackupRestore = &DigitalOceanDatabaseClusterBackupRestore{
				DatabaseName: "prod-postgres",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts a restore pinned to a backup timestamp", func() {
			spec := makeValidPostgresSpec()
			spec.BackupRestore = &DigitalOceanDatabaseClusterBackupRestore{
				DatabaseName:    "prod-postgres",
				BackupCreatedAt: "2026-08-01T00:00:00Z",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects a restore without database_name", func() {
			spec := makeValidPostgresSpec()
			spec.BackupRestore = &DigitalOceanDatabaseClusterBackupRestore{
				BackupCreatedAt: "2026-08-01T00:00:00Z",
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("storage_autoscale validation", func() {
		ginkgo.It("accepts enabled autoscale with defaults", func() {
			spec := makeValidPostgresSpec()
			spec.StorageAutoscale = &DigitalOceanDatabaseClusterStorageAutoscale{
				Enabled: true,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts autoscale with threshold and increment", func() {
			spec := makeValidPostgresSpec()
			spec.StorageAutoscale = &DigitalOceanDatabaseClusterStorageAutoscale{
				Enabled:          true,
				ThresholdPercent: 80,
				IncrementGib:     50,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects threshold_percent below 15", func() {
			spec := makeValidPostgresSpec()
			spec.StorageAutoscale = &DigitalOceanDatabaseClusterStorageAutoscale{
				Enabled:          true,
				ThresholdPercent: 10,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects threshold_percent above 95", func() {
			spec := makeValidPostgresSpec()
			spec.StorageAutoscale = &DigitalOceanDatabaseClusterStorageAutoscale{
				Enabled:          true,
				ThresholdPercent: 96,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects increment_gib below 10", func() {
			spec := makeValidPostgresSpec()
			spec.StorageAutoscale = &DigitalOceanDatabaseClusterStorageAutoscale{
				Enabled:      true,
				IncrementGib: 5,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("tags validation", func() {
		ginkgo.It("accepts valid tags", func() {
			spec := makeValidPostgresSpec()
			spec.Tags = []string{"env:prod", "team-data", "planton_managed"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects a tag with invalid characters", func() {
			spec := makeValidPostgresSpec()
			spec.Tags = []string{"has spaces"}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("Optional fields", func() {
		ginkgo.It("accepts spec with VPC reference", func() {
			spec := makeValidPostgresSpec()
			spec.Vpc = newVpcRef("12345678-1234-1234-1234-123456789012")
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts spec with custom storage_gib", func() {
			spec := makeValidPostgresSpec()
			spec.StorageGib = 100
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts spec with project_id", func() {
			spec := makeValidPostgresSpec()
			spec.ProjectId = &fk.StringValueOrRef{
				LiteralOrRef: &fk.StringValueOrRef_Value{Value: "12345678-1234-1234-1234-123456789012"},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Production configurations", func() {
		ginkgo.It("accepts production HA PostgreSQL with the full surface", func() {
			spec := &DigitalOceanDatabaseClusterSpec{
				ClusterName:   "prod-postgres",
				Engine:        DigitalOceanDatabaseEngine_pg,
				EngineVersion: "16",
				Region:        digitalocean.DigitalOceanRegion_nyc3,
				SizeSlug:      "db-s-4vcpu-8gb",
				NodeCount:     3,
				Vpc:           newVpcRef("12345678-1234-1234-1234-123456789012"),
				StorageGib:    200,
				MaintenanceWindow: &DigitalOceanDatabaseClusterMaintenanceWindow{
					Day:  "sunday",
					Hour: "02:00",
				},
				StorageAutoscale: &DigitalOceanDatabaseClusterStorageAutoscale{
					Enabled:          true,
					ThresholdPercent: 80,
					IncrementGib:     50,
				},
				Tags: []string{"env:prod"},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
