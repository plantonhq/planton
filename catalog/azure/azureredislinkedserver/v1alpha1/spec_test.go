package azureredislinkedserverv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureRedisLinkedServerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureRedisLinkedServerSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func cacheRef(name string, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      catalogkind.CatalogKind_AzureRedisCache,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

const (
	primaryCacheId   = "/subscriptions/s/resourceGroups/rg-east/providers/Microsoft.Cache/redis/app-cache-east"
	secondaryCacheId = "/subscriptions/s/resourceGroups/rg-west/providers/Microsoft.Cache/redis/app-cache-west"
)

// minimal valid spec: link a west-region secondary to an east-region primary.
func minimalSpec() *AzureRedisLinkedServer {
	return &AzureRedisLinkedServer{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureRedisLinkedServer",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-link",
		},
		Spec: &AzureRedisLinkedServerSpec{
			TargetRedisCacheId:       literal(primaryCacheId),
			LinkedRedisCacheId:       literal(secondaryCacheId),
			LinkedRedisCacheLocation: literal("westus2"),
			ServerRole:               AzureRedisLinkedServerRole_SECONDARY,
		},
	}
}

var _ = ginkgo.Describe("AzureRedisLinkedServerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal secondary link", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept the PRIMARY role", func() {
			input := minimalSpec()
			input.Spec.ServerRole = AzureRedisLinkedServerRole_PRIMARY
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept cache references for both ends and the location", func() {
			input := minimalSpec()
			input.Spec.TargetRedisCacheId = cacheRef("app-cache-east", "status.outputs.redis_cache_id")
			input.Spec.LinkedRedisCacheId = cacheRef("app-cache-west", "status.outputs.redis_cache_id")
			input.Spec.LinkedRedisCacheLocation = cacheRef("app-cache-west", "status.outputs.region")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing target cache", func() {
			input := minimalSpec()
			input.Spec.TargetRedisCacheId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing linked cache", func() {
			input := minimalSpec()
			input.Spec.LinkedRedisCacheId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing linked cache location", func() {
			input := minimalSpec()
			input.Spec.LinkedRedisCacheLocation = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unspecified server role", func() {
			input := minimalSpec()
			input.Spec.ServerRole = AzureRedisLinkedServerRole_azure_redis_linked_server_role_unspecified
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an undefined server role", func() {
			input := minimalSpec()
			input.Spec.ServerRole = AzureRedisLinkedServerRole(99)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
