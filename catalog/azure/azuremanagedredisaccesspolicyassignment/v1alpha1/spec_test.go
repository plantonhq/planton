package azuremanagedredisaccesspolicyassignmentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureManagedRedisAccessPolicyAssignmentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureManagedRedisAccessPolicyAssignmentSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const clusterId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cache/redisEnterprise/app-cache"

// minimal valid spec: grant data-plane access to a principal.
func minimalSpec() *AzureManagedRedisAccessPolicyAssignment {
	return &AzureManagedRedisAccessPolicyAssignment{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureManagedRedisAccessPolicyAssignment",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-grant",
		},
		Spec: &AzureManagedRedisAccessPolicyAssignmentSpec{
			ManagedRedisId: literal(clusterId),
			ObjectId:       literal("11111111-2222-3333-4444-555555555555"),
		},
	}
}

var _ = ginkgo.Describe("AzureManagedRedisAccessPolicyAssignmentSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal grant", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept an identity principal reference for the object id", func() {
			input := minimalSpec()
			input.Spec.ObjectId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureUserAssignedIdentity,
						Name:      "app-identity",
						FieldPath: "status.outputs.principal_id",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a cluster reference by kind output", func() {
			input := minimalSpec()
			input.Spec.ManagedRedisId = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureManagedRedis,
						Name:      "app-cache",
						FieldPath: "status.outputs.managed_redis_id",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing cluster reference", func() {
			input := minimalSpec()
			input.Spec.ManagedRedisId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing object id", func() {
			input := minimalSpec()
			input.Spec.ObjectId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
