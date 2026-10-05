package azuremanagedredisgeoreplicationv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureManagedRedisGeoReplicationSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureManagedRedisGeoReplicationSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	eastClusterId   = "/subscriptions/s/resourceGroups/rg-east/providers/Microsoft.Cache/redisEnterprise/app-cache-east"
	westClusterId   = "/subscriptions/s/resourceGroups/rg-west/providers/Microsoft.Cache/redisEnterprise/app-cache-west"
	europeClusterId = "/subscriptions/s/resourceGroups/rg-europe/providers/Microsoft.Cache/redisEnterprise/app-cache-europe"
	asiaClusterId   = "/subscriptions/s/resourceGroups/rg-asia/providers/Microsoft.Cache/redisEnterprise/app-cache-asia"
	southClusterId  = "/subscriptions/s/resourceGroups/rg-south/providers/Microsoft.Cache/redisEnterprise/app-cache-south"
)

// minimal valid spec: a two-region group.
func minimalSpec() *AzureManagedRedisGeoReplication {
	return &AzureManagedRedisGeoReplication{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureManagedRedisGeoReplication",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-geo-replication",
		},
		Spec: &AzureManagedRedisGeoReplicationSpec{
			ManagedRedisId:        literal(eastClusterId),
			LinkedManagedRedisIds: []*foreignkeyv1.StringValueOrRef{literal(westClusterId)},
		},
	}
}

var _ = ginkgo.Describe("AzureManagedRedisGeoReplicationSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a two-region group", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept the maximum of four linked members", func() {
			input := minimalSpec()
			input.Spec.LinkedManagedRedisIds = []*foreignkeyv1.StringValueOrRef{
				literal(westClusterId),
				literal(europeClusterId),
				literal(asiaClusterId),
				literal(southClusterId),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept cluster references by kind output", func() {
			input := minimalSpec()
			input.Spec.LinkedManagedRedisIds = []*foreignkeyv1.StringValueOrRef{
				{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
						ValueFrom: &foreignkeyv1.ValueFromRef{
							Kind:      catalogkind.CatalogKind_AzureManagedRedis,
							Name:      "app-cache-west",
							FieldPath: "status.outputs.managed_redis_id",
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing managing cluster", func() {
			input := minimalSpec()
			input.Spec.ManagedRedisId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty linked list", func() {
			input := minimalSpec()
			input.Spec.LinkedManagedRedisIds = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject more than four linked members", func() {
			input := minimalSpec()
			input.Spec.LinkedManagedRedisIds = []*foreignkeyv1.StringValueOrRef{
				literal(westClusterId),
				literal(europeClusterId),
				literal(asiaClusterId),
				literal(southClusterId),
				literal("/subscriptions/s/resourceGroups/rg-x/providers/Microsoft.Cache/redisEnterprise/app-cache-x"),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
