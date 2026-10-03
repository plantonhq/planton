package azuremssqlelasticpoolv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMssqlElasticPoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMssqlElasticPoolSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const serverId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Sql/servers/orders-sql"

// minimal valid spec: a small standard DTU pool.
func minimalSpec() *AzureMssqlElasticPool {
	return &AzureMssqlElasticPool{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMssqlElasticPool",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-pool",
		},
		Spec: &AzureMssqlElasticPoolSpec{
			ServerId: literal(serverId),
			Region:   "eastus",
			PoolName: "tenant-pool",
			SkuName:  "StandardPool",
			Capacity: 50,
			PerDatabaseSettings: &AzureMssqlElasticPoolPerDatabaseSettings{
				MinCapacity: 0,
				MaxCapacity: 50,
			},
		},
	}
}

var _ = ginkgo.Describe("AzureMssqlElasticPoolSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal DTU pool", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept every pool sku", func() {
			for _, sku := range []string{"BasicPool", "StandardPool", "PremiumPool", "GP_Gen5", "GP_Fsv2", "GP_DC", "BC_Gen5", "BC_DC", "HS_Gen5", "HS_PRMS", "HS_MOPRMS"} {
				input := minimalSpec()
				input.Spec.SkuName = sku
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "sku %q must be accepted", sku)
			}
		})

		ginkgo.It("should accept a vCore pool with license and fractional per-database bounds", func() {
			input := minimalSpec()
			input.Spec.SkuName = "GP_Gen5"
			input.Spec.Capacity = 4
			input.Spec.LicenseType = AzureMssqlElasticPoolLicenseType_BASE_PRICE
			input.Spec.PerDatabaseSettings = &AzureMssqlElasticPoolPerDatabaseSettings{
				MinCapacity: 0.25,
				MaxCapacity: 2,
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a Hyperscale pool with HA replicas", func() {
			replicas := int32(2)
			input := minimalSpec()
			input.Spec.SkuName = "HS_Gen5"
			input.Spec.Capacity = 4
			input.Spec.HighAvailabilityReplicaCount = &replicas
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a storage cap in gigabytes", func() {
			size := 100.0
			input := minimalSpec()
			input.Spec.MaxSizeGb = &size
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a storage cap in bytes", func() {
			bytes := int64(107374182400)
			input := minimalSpec()
			input.Spec.MaxSizeBytes = &bytes
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept zone redundancy, enclave, maintenance, and tags", func() {
			maintenance := "SQL_EastUS_DB_1"
			input := minimalSpec()
			input.Spec.SkuName = "BC_Gen5"
			input.Spec.Capacity = 4
			input.Spec.ZoneRedundant = true
			input.Spec.EnclaveType = AzureMssqlElasticPoolEnclaveType_VBS
			input.Spec.MaintenanceConfigurationName = &maintenance
			input.Spec.Tags = map[string]string{"team": "data"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a valueFrom reference for the server", func() {
			input := minimalSpec()
			input.Spec.ServerId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureMssqlServer,
						Name:      "orders-sql",
						FieldPath: "status.outputs.server_id",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing server", func() {
			input := minimalSpec()
			input.Spec.ServerId = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a missing pool name", func() {
			input := minimalSpec()
			input.Spec.PoolName = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a missing region", func() {
			input := minimalSpec()
			input.Spec.Region = ""
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown sku", func() {
			for _, sku := range []string{"Standard", "GP_Gen4", "BasicPool2", "gp_gen5"} {
				input := minimalSpec()
				input.Spec.SkuName = sku
				gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil(), "sku %q must be rejected", sku)
			}
		})

		ginkgo.It("should reject a zero capacity", func() {
			input := minimalSpec()
			input.Spec.Capacity = 0
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject missing per-database settings", func() {
			input := minimalSpec()
			input.Spec.PerDatabaseSettings = nil
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject per-database min above max", func() {
			input := minimalSpec()
			input.Spec.PerDatabaseSettings = &AzureMssqlElasticPoolPerDatabaseSettings{
				MinCapacity: 10,
				MaxCapacity: 5,
			}
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject setting both storage caps", func() {
			size := 100.0
			bytes := int64(107374182400)
			input := minimalSpec()
			input.Spec.MaxSizeGb = &size
			input.Spec.MaxSizeBytes = &bytes
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject HA replicas outside Hyperscale", func() {
			replicas := int32(2)
			input := minimalSpec()
			input.Spec.HighAvailabilityReplicaCount = &replicas
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})

		ginkgo.It("should reject a license type on a DTU pool", func() {
			input := minimalSpec()
			input.Spec.LicenseType = AzureMssqlElasticPoolLicenseType_LICENSE_INCLUDED
			gomega.Expect(protovalidate.Validate(input)).ToNot(gomega.BeNil())
		})
	})
})
