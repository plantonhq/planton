package azurecosmosdbsqlroledefinitionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureCosmosdbSqlRoleDefinitionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureCosmosdbSqlRoleDefinitionSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const accountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.DocumentDB/databaseAccounts/planton-cosmos"

// minimal valid spec: a custom read-only role assignable anywhere in the
// account.
func minimalSpec() *AzureCosmosdbSqlRoleDefinition {
	return &AzureCosmosdbSqlRoleDefinition{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureCosmosdbSqlRoleDefinition",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-sql-role-definition",
		},
		Spec: &AzureCosmosdbSqlRoleDefinitionSpec{
			CosmosdbAccountId: literal(accountId),
			RoleName:          "app-data-reader",
			AssignableScopes:  []*foreignkeyv1.StringValueOrRef{literal(accountId)},
			Permissions: []*AzureCosmosdbSqlRoleDefinitionPermission{
				{
					DataActions: []string{
						"Microsoft.DocumentDB/databaseAccounts/readMetadata",
						"Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/items/read",
					},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureCosmosdbSqlRoleDefinitionSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal custom role", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept multiple additive permission blocks", func() {
			input := minimalSpec()
			input.Spec.Permissions = append(input.Spec.Permissions,
				&AzureCosmosdbSqlRoleDefinitionPermission{
					DataActions: []string{"Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/executeQuery"},
				})
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a database-scoped assignable scope beside the account scope", func() {
			input := minimalSpec()
			input.Spec.AssignableScopes = append(input.Spec.AssignableScopes,
				literal(accountId+"/dbs/app-data"))
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept an explicit type", func() {
			for _, roleType := range []AzureCosmosdbSqlRoleDefinitionType{
				AzureCosmosdbSqlRoleDefinitionType_CUSTOM_ROLE,
				AzureCosmosdbSqlRoleDefinitionType_BUILT_IN_ROLE,
			} {
				input := minimalSpec()
				input.Spec.Type = roleType
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "type %s must be accepted", roleType)
			}
		})

		ginkgo.It("should accept a pinned UUID role_definition_id", func() {
			input := minimalSpec()
			input.Spec.RoleDefinitionId = "9b7f3f6a-2f0e-4b9a-8f0d-2f6a8f0d2f6a"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing account reference", func() {
			input := minimalSpec()
			input.Spec.CosmosdbAccountId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing role name", func() {
			input := minimalSpec()
			input.Spec.RoleName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject empty assignable_scopes", func() {
			input := minimalSpec()
			input.Spec.AssignableScopes = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a role with no permission blocks", func() {
			input := minimalSpec()
			input.Spec.Permissions = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a permission block with no data actions", func() {
			input := minimalSpec()
			input.Spec.Permissions[0].DataActions = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an empty data action", func() {
			input := minimalSpec()
			input.Spec.Permissions[0].DataActions = []string{""}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a non-UUID pinned role_definition_id", func() {
			input := minimalSpec()
			input.Spec.RoleDefinitionId = "not-a-uuid"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
