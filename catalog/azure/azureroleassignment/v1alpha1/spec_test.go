package azureroleassignmentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureRoleAssignmentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureRoleAssignmentSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// ref builds a StringValueOrRef carrying a value_from reference.
func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

// validResource returns a minimal valid AzureRoleAssignment that individual
// cases then mutate into the shape under test.
func validResource() *AzureRoleAssignment {
	return &AzureRoleAssignment{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureRoleAssignment",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-role-assignment",
		},
		Spec: &AzureRoleAssignmentSpec{
			Scope:              literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg"),
			RoleDefinitionName: "Reader",
			PrincipalId:        literal("11111111-1111-1111-1111-111111111111"),
		},
	}
}

var _ = ginkgo.Describe("AzureRoleAssignmentSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_role_assignment", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a role referenced by role_definition_id instead of name", func() {
				input := validResource()
				input.Spec.RoleDefinitionName = ""
				input.Spec.RoleDefinitionId = literal("/subscriptions/00000000-0000-0000-0000-000000000000/providers/Microsoft.Authorization/roleDefinitions/acdd72a7-3385-48ef-bd42-f606fba81ae7")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a role_definition_id referencing an AzureRoleDefinition", func() {
				// The composed-environment shape: the custom role deploys in
				// the same run and its fully-scoped ID flows in by reference.
				input := validResource()
				input.Spec.RoleDefinitionName = ""
				input.Spec.RoleDefinitionId = ref("velero-backup-role")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept scope and principal_id as references", func() {
				input := validResource()
				input.Spec.Scope = ref("platform-rg")
				input.Spec.PrincipalId = ref("workload-identity")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit principal_type", func() {
				input := validResource()
				input.Spec.PrincipalType = AzureRoleAssignmentPrincipalType_SERVICE_PRINCIPAL
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an ABAC condition with condition_version 2.0", func() {
				input := validResource()
				input.Spec.Condition = "((!(ActionMatches{'Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read'})))"
				input.Spec.ConditionVersion = "2.0"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an ABAC condition without an explicit version", func() {
				input := validResource()
				input.Spec.Condition = "((!(ActionMatches{'Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read'})))"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a pinned UUID name", func() {
				input := validResource()
				input.Spec.Name = "a67e1183-4b2d-4b6e-93f1-2b2b8d2e1c11"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept skip_service_principal_aad_check and description", func() {
				input := validResource()
				input.Spec.SkipServicePrincipalAadCheck = true
				input.Spec.Description = "CI deploy identity needs read access to the platform resource group"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a delegated managed identity resource id", func() {
				input := validResource()
				input.Spec.DelegatedManagedIdentityResourceId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/mi-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/lighthouse-mi"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_role_assignment", func() {

			ginkgo.It("should return a validation error when scope is missing", func() {
				input := validResource()
				input.Spec.Scope = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when scope is an empty literal", func() {
				input := validResource()
				input.Spec.Scope = literal("")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when principal_id is missing", func() {
				input := validResource()
				input.Spec.PrincipalId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when neither role name nor role id is set", func() {
				input := validResource()
				input.Spec.RoleDefinitionName = ""
				input.Spec.RoleDefinitionId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when both role name and role id are set", func() {
				input := validResource()
				input.Spec.RoleDefinitionId = literal("/subscriptions/00000000-0000-0000-0000-000000000000/providers/Microsoft.Authorization/roleDefinitions/acdd72a7-3385-48ef-bd42-f606fba81ae7")
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when condition_version is set without condition", func() {
				input := validResource()
				input.Spec.ConditionVersion = "2.0"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an unsupported condition_version", func() {
				input := validResource()
				input.Spec.Condition = "((!(ActionMatches{'Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read'})))"
				input.Spec.ConditionVersion = "3.0"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is not a UUID", func() {
				input := validResource()
				input.Spec.Name = "not-a-uuid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when api_version is incorrect", func() {
				input := validResource()
				input.ApiVersion = "wrong.version/v1"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when kind is incorrect", func() {
				input := validResource()
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when metadata is missing", func() {
				input := validResource()
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when spec is missing", func() {
				input := validResource()
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
