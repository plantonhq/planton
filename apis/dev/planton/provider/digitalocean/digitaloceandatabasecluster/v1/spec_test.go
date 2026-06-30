package digitaloceandatabaseclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	digitalocean "github.com/plantonhq/planton/apis/dev/planton/provider/digitalocean"
	fk "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestDigitalOceanDatabaseClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanDatabaseClusterSpec Validation Suite")
}

var _ = ginkgo.Describe("DigitalOceanDatabaseClusterSpec validations", func() {

	// Helper to create VPC reference
	newVpcRef := func(vpcId string) *fk.StringValueOrRef {
		return &fk.StringValueOrRef{
			LiteralOrRef: &fk.StringValueOrRef_Value{Value: vpcId},
		}
	}

	// Helper function to create a minimal valid PostgreSQL spec
	makeValidPostgresSpec := func() *DigitalOceanDatabaseClusterSpec {
		return &DigitalOceanDatabaseClusterSpec{
			ClusterName:              "test-postgres",
			Engine:                   DigitalOceanDatabaseEngine_pg,
			EngineVersion:            "16",
			Region:                   digitalocean.DigitalOceanRegion_nyc3,
			SizeSlug:                 "db-s-1vcpu-1gb",
			NodeCount:                1,
			EnablePublicConnectivity: false,
		}
	}

	// Helper function to create a minimal valid MySQL spec
	makeValidMysqlSpec := func() *DigitalOceanDatabaseClusterSpec {
		return &DigitalOceanDatabaseClusterSpec{
			ClusterName:              "test-mysql",
			Engine:                   DigitalOceanDatabaseEngine_mysql,
			EngineVersion:            "8",
			Region:                   digitalocean.DigitalOceanRegion_sfo3,
			SizeSlug:                 "db-s-2vcpu-4gb",
			NodeCount:                2,
			EnablePublicConnectivity: false,
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

		ginkgo.It("accepts cluster_name with hyphens", func() {
			spec := makeValidPostgresSpec()
			spec.ClusterName = "prod-postgres-cluster-2025"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("engine_version validation", func() {
		ginkgo.It("accepts major version only (e.g., '16')", func() {
			spec := makeValidPostgresSpec()
			spec.EngineVersion = "16"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts major.minor version (e.g., '8.0')", func() {
			spec := makeValidMysqlSpec()
			spec.EngineVersion = "8.0"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects invalid version format (text)", func() {
			spec := makeValidPostgresSpec()
			spec.EngineVersion = "latest"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects version with patch number (e.g., '16.1.2')", func() {
			spec := makeValidPostgresSpec()
			spec.EngineVersion = "16.1.2"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("node_count validation", func() {
		ginkgo.It("accepts node_count = 1 (single node)", func() {
			spec := makeValidPostgresSpec()
			spec.NodeCount = 1
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts node_count = 2 (HA minimum)", func() {
			spec := makeValidPostgresSpec()
			spec.NodeCount = 2
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts node_count = 3 (HA maximum)", func() {
			spec := makeValidPostgresSpec()
			spec.NodeCount = 3
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects node_count = 0", func() {
			spec := makeValidPostgresSpec()
			spec.NodeCount = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects node_count = 4 (exceeds max)", func() {
			spec := makeValidPostgresSpec()
			spec.NodeCount = 4
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

		ginkgo.It("accepts spec without VPC (nil)", func() {
			spec := makeValidPostgresSpec()
			spec.Vpc = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts spec with custom storage_gib", func() {
			spec := makeValidPostgresSpec()
			spec.StorageGib = 100
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts spec with storage_gib = 0 (use default)", func() {
			spec := makeValidPostgresSpec()
			spec.StorageGib = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts spec with enable_public_connectivity = true", func() {
			spec := makeValidPostgresSpec()
			spec.EnablePublicConnectivity = true
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts spec with enable_public_connectivity = false", func() {
			spec := makeValidPostgresSpec()
			spec.EnablePublicConnectivity = false
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Engine-specific configurations", func() {
		ginkgo.It("accepts PostgreSQL with version 14", func() {
			spec := makeValidPostgresSpec()
			spec.Engine = DigitalOceanDatabaseEngine_pg
			spec.EngineVersion = "14"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts MySQL with version 8", func() {
			spec := makeValidMysqlSpec()
			spec.Engine = DigitalOceanDatabaseEngine_mysql
			spec.EngineVersion = "8"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts Redis with version 7", func() {
			spec := makeValidPostgresSpec()
			spec.Engine = DigitalOceanDatabaseEngine_redis
			spec.EngineVersion = "7"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts MongoDB with version 7.0", func() {
			spec := makeValidPostgresSpec()
			spec.Engine = DigitalOceanDatabaseEngine_mongodb
			spec.EngineVersion = "7.0"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Production configurations", func() {
		ginkgo.It("accepts production HA PostgreSQL with VPC", func() {
			spec := &DigitalOceanDatabaseClusterSpec{
				ClusterName:              "prod-postgres",
				Engine:                   DigitalOceanDatabaseEngine_pg,
				EngineVersion:            "16",
				Region:                   digitalocean.DigitalOceanRegion_nyc3,
				SizeSlug:                 "db-s-4vcpu-8gb",
				NodeCount:                3,
				Vpc:                      newVpcRef("vpc-prod-12345"),
				StorageGib:               200,
				EnablePublicConnectivity: false,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts production MySQL cluster", func() {
			spec := &DigitalOceanDatabaseClusterSpec{
				ClusterName:              "prod-mysql",
				Engine:                   DigitalOceanDatabaseEngine_mysql,
				EngineVersion:            "8",
				Region:                   digitalocean.DigitalOceanRegion_sfo3,
				SizeSlug:                 "db-s-2vcpu-4gb",
				NodeCount:                2,
				Vpc:                      newVpcRef("vpc-prod-67890"),
				StorageGib:               150,
				EnablePublicConnectivity: false,
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("Edge cases", func() {
		ginkgo.It("accepts different regions", func() {
			spec := makeValidPostgresSpec()
			spec.Region = digitalocean.DigitalOceanRegion_lon1
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts large storage values", func() {
			spec := makeValidPostgresSpec()
			spec.StorageGib = 1000
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts various size slugs", func() {
			spec := makeValidPostgresSpec()
			spec.SizeSlug = "db-s-8vcpu-16gb"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
