package azurerediscacheaccesspolicyv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureRedisCacheAccessPolicySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureRedisCacheAccessPolicySpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const cacheId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cache/redis/app-cache"

// minimal valid spec: a read-only policy scoped to one key prefix.
func minimalSpec() *AzureRedisCacheAccessPolicy {
	return &AzureRedisCacheAccessPolicy{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureRedisCacheAccessPolicy",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-policy",
		},
		Spec: &AzureRedisCacheAccessPolicySpec{
			RedisCacheId: literal(cacheId),
			PolicyName:   "app-read-only",
			Permissions:  "+@read +@connection ~app:*",
		},
	}
}

var _ = ginkgo.Describe("AzureRedisCacheAccessPolicySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal custom policy", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a cache reference", func() {
			input := minimalSpec()
			input.Spec.RedisCacheId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureRedisCache,
						Name:      "app-cache",
						FieldPath: "status.outputs.redis_cache_id",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept complex ACL permission sets", func() {
			input := minimalSpec()
			input.Spec.Permissions = "+@all -@dangerous ~app1:* ~app2:*"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing cache reference", func() {
			input := minimalSpec()
			input.Spec.RedisCacheId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing policy name", func() {
			input := minimalSpec()
			input.Spec.PolicyName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject the built-in policy names", func() {
			for _, name := range []string{"Data Owner", "Data Contributor", "Data Reader"} {
				input := minimalSpec()
				input.Spec.PolicyName = name
				gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), "built-in name %q must be rejected", name)
			}
		})

		ginkgo.It("should reject a policy name longer than 63 characters", func() {
			input := minimalSpec()
			input.Spec.PolicyName = "a123456789b123456789c123456789d123456789e123456789f123456789g123"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject missing permissions", func() {
			input := minimalSpec()
			input.Spec.Permissions = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
