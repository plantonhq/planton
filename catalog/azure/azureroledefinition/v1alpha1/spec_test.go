package azureroledefinitionv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureRoleDefinitionSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureRoleDefinitionSpec Validation Tests")
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

// validResource returns a minimal valid AzureRoleDefinition that individual
// cases then mutate into the shape under test.
func validResource() *AzureRoleDefinition {
	return &AzureRoleDefinition{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureRoleDefinition",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-role-definition",
		},
		Spec: &AzureRoleDefinitionSpec{
			Name:  "acme-test-operator",
			Scope: literal("/subscriptions/00000000-0000-0000-0000-000000000000"),
			Permissions: []*AzureRoleDefinitionPermission{
				{
					Actions: []string{"Microsoft.Resources/subscriptions/resourceGroups/read"},
				},
			},
		},
	}
}

var _ = ginkgo.Describe("AzureRoleDefinitionSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_role_definition", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a resource-group creation scope", func() {
				input := validResource()
				input.Spec.Scope = literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a management-group creation scope", func() {
				input := validResource()
				input.Spec.Scope = literal("/providers/Microsoft.Management/managementGroups/platform-mg")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the scope as a reference", func() {
				input := validResource()
				input.Spec.Scope = ref("platform-rg")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a definition with no permission blocks", func() {
				input := validResource()
				input.Spec.Permissions = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a permission block with all four operation lists", func() {
				input := validResource()
				input.Spec.Permissions = []*AzureRoleDefinitionPermission{
					{
						Actions:        []string{"*"},
						NotActions:     []string{"Microsoft.Authorization/*/write"},
						DataActions:    []string{"Microsoft.Storage/storageAccounts/blobServices/containers/blobs/*"},
						NotDataActions: []string{"Microsoft.Storage/storageAccounts/blobServices/containers/blobs/delete"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept multiple permission blocks", func() {
				input := validResource()
				input.Spec.Permissions = []*AzureRoleDefinitionPermission{
					{Actions: []string{"Microsoft.Compute/virtualMachines/read"}},
					{DataActions: []string{"Microsoft.Storage/storageAccounts/blobServices/containers/blobs/read"}},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept assignable scopes as literals", func() {
				input := validResource()
				input.Spec.AssignableScopes = []*foreignkeyv1.StringValueOrRef{
					literal("/subscriptions/00000000-0000-0000-0000-000000000000"),
					literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept assignable scopes mixing references and literals", func() {
				input := validResource()
				input.Spec.AssignableScopes = []*foreignkeyv1.StringValueOrRef{
					ref("platform-rg"),
					literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/other-rg"),
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a description", func() {
				input := validResource()
				input.Spec.Description = "Operate existing VMs: start/stop/restart, no create or delete"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a pinned UUID role_definition_id", func() {
				input := validResource()
				input.Spec.RoleDefinitionId = "b24988ac-6180-42a0-ab88-20f7382dd24c"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_role_definition", func() {

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when scope is missing", func() {
				input := validResource()
				input.Spec.Scope = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when role_definition_id is not a UUID", func() {
				input := validResource()
				input.Spec.RoleDefinitionId = "not-a-uuid"
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
