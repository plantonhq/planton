package civodatabasev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	civo "github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestCivoDatabaseSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CivoDatabaseSpec Validation Suite")
}

var _ = ginkgo.Describe("CivoDatabaseSpec validations", func() {

	// Helper function to create a minimal valid spec
	makeValidSpec := func() *CivoDatabaseSpec {
		return &CivoDatabaseSpec{
			DbInstanceName: "test-db",
			Engine:         CivoDatabaseEngine_postgres,
			EngineVersion:  "16",
			Region:         civo.CivoRegion_lon1,
			SizeSlug:       "g3.db.small",
			NetworkId: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "net-12345"},
			},
		}
	}

	ginkgo.Context("Required fields", func() {
		ginkgo.It("accepts a minimal valid spec", func() {
			spec := makeValidSpec()
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing db_instance_name", func() {
			spec := makeValidSpec()
			spec.DbInstanceName = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing engine", func() {
			spec := makeValidSpec()
			spec.Engine = CivoDatabaseEngine_civo_database_engine_unspecified
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing engine_version", func() {
			spec := makeValidSpec()
			spec.EngineVersion = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing region", func() {
			spec := makeValidSpec()
			spec.Region = civo.CivoRegion_civo_region_unspecified
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing size_slug", func() {
			spec := makeValidSpec()
			spec.SizeSlug = ""
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects spec with missing network_id", func() {
			spec := makeValidSpec()
			spec.NetworkId = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("db_instance_name validation", func() {
		ginkgo.It("accepts name within 64 character limit", func() {
			spec := makeValidSpec()
			spec.DbInstanceName = "a-valid-database-name-with-64-characters-exactly-01234567890123"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects name exceeding 64 characters", func() {
			spec := makeValidSpec()
			spec.DbInstanceName = "a-very-long-database-name-that-exceeds-the-maximum-allowed-length-of-64-characters-01234567890"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts short name", func() {
			spec := makeValidSpec()
			spec.DbInstanceName = "db"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("engine_version validation", func() {
		ginkgo.It("accepts major version only", func() {
			spec := makeValidSpec()
			spec.EngineVersion = "16"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts major.minor version", func() {
			spec := makeValidSpec()
			spec.EngineVersion = "8.0"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts major.minor with multiple digits", func() {
			spec := makeValidSpec()
			spec.EngineVersion = "14.10"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects version with patch", func() {
			spec := makeValidSpec()
			spec.EngineVersion = "16.2.1"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects version with letters", func() {
			spec := makeValidSpec()
			spec.EngineVersion = "16a"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("rejects version with spaces", func() {
			spec := makeValidSpec()
			spec.EngineVersion = "16 .0"
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("replicas validation", func() {
		ginkgo.It("accepts 0 replicas (master only)", func() {
			spec := makeValidSpec()
			spec.Replicas = 0
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts 1 replica", func() {
			spec := makeValidSpec()
			spec.Replicas = 1
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts 2 replicas", func() {
			spec := makeValidSpec()
			spec.Replicas = 2
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts 4 replicas (max allowed)", func() {
			spec := makeValidSpec()
			spec.Replicas = 4
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects 5 replicas (exceeds max)", func() {
			spec := makeValidSpec()
			spec.Replicas = 5
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("engine enum validation", func() {
		ginkgo.It("accepts mysql engine", func() {
			spec := makeValidSpec()
			spec.Engine = CivoDatabaseEngine_mysql
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("accepts postgres engine", func() {
			spec := makeValidSpec()
			spec.Engine = CivoDatabaseEngine_postgres
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects unspecified engine", func() {
			spec := makeValidSpec()
			spec.Engine = CivoDatabaseEngine_civo_database_engine_unspecified
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("complete production configuration", func() {
		ginkgo.It("accepts full spec with all optional fields", func() {
			spec := &CivoDatabaseSpec{
				DbInstanceName: "production-database",
				Engine:         CivoDatabaseEngine_postgres,
				EngineVersion:  "16",
				Region:         civo.CivoRegion_lon1,
				SizeSlug:       "g3.db.large",
				Replicas:       2,
				NetworkId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "network-uuid-12345",
					},
				},
				FirewallIds: []*foreignkeyv1.StringValueOrRef{
					{
						LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
							Value: "firewall-uuid-67890",
						},
					},
				},
				StorageGib: 200,
				Tags:       []string{"production", "backend", "primary"},
			}

			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("network_id foreign key", func() {
		ginkgo.It("accepts literal network ID", func() {
			spec := makeValidSpec()
			spec.NetworkId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "net-uuid-12345",
				},
			}
			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("rejects nil network ID", func() {
			spec := makeValidSpec()
			spec.NetworkId = nil
			err := protovalidate.Validate(spec)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("optional fields", func() {
		ginkgo.It("accepts spec without optional fields", func() {
			spec := &CivoDatabaseSpec{
				DbInstanceName: "test-db",
				Engine:         CivoDatabaseEngine_mysql,
				EngineVersion:  "8.0",
				Region:         civo.CivoRegion_nyc1,
				SizeSlug:       "g3.db.medium",
				Replicas:       0,
				NetworkId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "net-12345",
					},
				},
				// Omitting: FirewallIds, StorageGib, Tags
			}

			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Context("multiple firewalls", func() {
		ginkgo.It("accepts multiple firewall IDs", func() {
			spec := makeValidSpec()
			spec.FirewallIds = []*foreignkeyv1.StringValueOrRef{
				{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "fw-uuid-1",
					},
				},
				{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "fw-uuid-2",
					},
				},
			}

			err := protovalidate.Validate(spec)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
