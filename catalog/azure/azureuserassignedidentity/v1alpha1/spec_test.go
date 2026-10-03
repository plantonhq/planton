package azureuserassignedidentityv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureUserAssignedIdentitySpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureUserAssignedIdentitySpec Validation Tests")
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

// validResource returns a minimal valid AzureUserAssignedIdentity that
// individual cases then mutate into the shape under test.
func validResource() *AzureUserAssignedIdentity {
	return &AzureUserAssignedIdentity{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureUserAssignedIdentity",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-identity",
		},
		Spec: &AzureUserAssignedIdentitySpec{
			Region:        "eastus",
			ResourceGroup: literal("test-rg"),
			Name:          "test-managed-identity",
		},
	}
}

var _ = ginkgo.Describe("AzureUserAssignedIdentitySpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_user_assigned_identity", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the resource group as a reference", func() {
				input := validResource()
				input.Spec.ResourceGroup = ref("platform-rg")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept regional isolation", func() {
				input := validResource()
				input.Spec.IsolationScope = AzureUserAssignedIdentityIsolationScope_REGIONAL
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept user tags", func() {
				input := validResource()
				input.Spec.Tags = map[string]string{
					"cost-center": "platform",
					"owner":       "identity-team",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a name at the 3-character minimum", func() {
				input := validResource()
				input.Spec.Name = "cid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a name at the 128-character maximum", func() {
				input := validResource()
				long := make([]byte, 128)
				for i := range long {
					long[i] = 'a'
				}
				input.Spec.Name = string(long)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_user_assigned_identity", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when resource_group is missing", func() {
				input := validResource()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is missing", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name is shorter than 3 characters", func() {
				input := validResource()
				input.Spec.Name = "ab"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when name exceeds 128 characters", func() {
				input := validResource()
				long := make([]byte, 129)
				for i := range long {
					long[i] = 'a'
				}
				input.Spec.Name = string(long)
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
