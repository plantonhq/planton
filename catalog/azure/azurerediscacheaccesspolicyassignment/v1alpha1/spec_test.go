package azurerediscacheaccesspolicyassignmentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureRedisCacheAccessPolicyAssignmentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureRedisCacheAccessPolicyAssignmentSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const cacheId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cache/redis/app-cache"

// minimal valid spec: grant the built-in read-only policy to a principal.
func minimalSpec() *AzureRedisCacheAccessPolicyAssignment {
	return &AzureRedisCacheAccessPolicyAssignment{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureRedisCacheAccessPolicyAssignment",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-assignment",
		},
		Spec: &AzureRedisCacheAccessPolicyAssignmentSpec{
			RedisCacheId:     literal(cacheId),
			AssignmentName:   "app-identity-data-reader",
			AccessPolicyName: literal("Data Reader"),
			ObjectId:         literal("11111111-2222-3333-4444-555555555555"),
			ObjectIdAlias:    "app-identity",
		},
	}
}

var _ = ginkgo.Describe("AzureRedisCacheAccessPolicyAssignmentSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal built-in policy grant", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept every built-in policy name", func() {
			for _, policy := range []string{"Data Owner", "Data Contributor", "Data Reader"} {
				input := minimalSpec()
				input.Spec.AccessPolicyName = literal(policy)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "policy %q must be accepted", policy)
			}
		})

		ginkgo.It("should accept a custom policy reference", func() {
			input := minimalSpec()
			input.Spec.AccessPolicyName = &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
					ValueFrom: &foreignkeyv1.ValueFromRef{
						Kind:      catalogkind.CatalogKind_AzureRedisCacheAccessPolicy,
						Name:      "app-read-only",
						FieldPath: "status.outputs.access_policy_name",
					},
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
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
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing cache reference", func() {
			input := minimalSpec()
			input.Spec.RedisCacheId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing assignment name", func() {
			input := minimalSpec()
			input.Spec.AssignmentName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an assignment name longer than 63 characters", func() {
			input := minimalSpec()
			input.Spec.AssignmentName = "a123456789b123456789c123456789d123456789e123456789f123456789g123"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing policy reference", func() {
			input := minimalSpec()
			input.Spec.AccessPolicyName = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing object id", func() {
			input := minimalSpec()
			input.Spec.ObjectId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing object id alias", func() {
			input := minimalSpec()
			input.Spec.ObjectIdAlias = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
