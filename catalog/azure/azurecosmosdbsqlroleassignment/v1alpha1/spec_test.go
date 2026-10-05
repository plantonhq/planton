package azurecosmosdbsqlroleassignmentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureCosmosdbSqlRoleAssignmentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureCosmosdbSqlRoleAssignmentSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	accountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.DocumentDB/databaseAccounts/planton-cosmos"
	// The built-in Data Contributor role's well-known ID inside the account.
	dataContributorId = accountId + "/sqlRoleDefinitions/00000000-0000-0000-0000-000000000002"
	principalObjectId = "c3b2a190-8f7e-4d6c-b5a4-93d2c1b0a987"
)

// minimal valid spec: the built-in Data Contributor granted to a managed
// identity's object ID across the whole account.
func minimalSpec() *AzureCosmosdbSqlRoleAssignment {
	return &AzureCosmosdbSqlRoleAssignment{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureCosmosdbSqlRoleAssignment",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-sql-role-assignment",
		},
		Spec: &AzureCosmosdbSqlRoleAssignmentSpec{
			CosmosdbAccountId: literal(accountId),
			RoleDefinitionId:  literal(dataContributorId),
			PrincipalId:       literal(principalObjectId),
			Scope:             literal(accountId),
		},
	}
}

var _ = ginkgo.Describe("AzureCosmosdbSqlRoleAssignmentSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal account-wide built-in grant", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept a database-scoped grant", func() {
			input := minimalSpec()
			input.Spec.Scope = literal(accountId + "/dbs/app-data")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a container-scoped grant", func() {
			input := minimalSpec()
			input.Spec.Scope = literal(accountId + "/dbs/app-data/colls/orders")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a pinned UUID name", func() {
			input := minimalSpec()
			input.Spec.Name = "7c1de3f8-5a4b-4c2d-9e8f-1a2b3c4d5e6f"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing account reference", func() {
			input := minimalSpec()
			input.Spec.CosmosdbAccountId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing role definition reference", func() {
			input := minimalSpec()
			input.Spec.RoleDefinitionId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing principal reference", func() {
			input := minimalSpec()
			input.Spec.PrincipalId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing scope", func() {
			input := minimalSpec()
			input.Spec.Scope = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a non-UUID pinned name", func() {
			input := minimalSpec()
			input.Spec.Name = "not-a-uuid"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
