package azurefederatedidentitycredentialv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureFederatedIdentityCredentialSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFederatedIdentityCredentialSpec Validation Tests")
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

// validResource returns a minimal valid AzureFederatedIdentityCredential that
// individual cases then mutate into the shape under test.
func validResource() *AzureFederatedIdentityCredential {
	return &AzureFederatedIdentityCredential{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFederatedIdentityCredential",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-federated-credential",
		},
		Spec: &AzureFederatedIdentityCredentialSpec{
			Name:                 "github-main-branch",
			UserAssignedIdentity: literal("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/test-identity"),
			Issuer:               literal("https://token.actions.githubusercontent.com"),
			Subject:              "repo:acme/platform:ref:refs/heads/main",
		},
	}
}

var _ = ginkgo.Describe("AzureFederatedIdentityCredentialSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_federated_identity_credential", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the parent identity as a reference", func() {
				input := validResource()
				input.Spec.UserAssignedIdentity = ref("platform-identity")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a GitHub environment subject", func() {
				input := validResource()
				input.Spec.Subject = "repo:acme/platform:environment:production"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an AKS workload-identity trust with a literal issuer URL", func() {
				input := validResource()
				input.Spec.Issuer = literal("https://eastus.oic.prod-aks.azure.com/00000000-0000-0000-0000-000000000000/11111111-1111-1111-1111-111111111111/")
				input.Spec.Subject = "system:serviceaccount:payments:payments-api"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the issuer as a reference to a cluster's OIDC issuer output", func() {
				input := validResource()
				input.Spec.Issuer = ref("platform-aks-cluster")
				input.Spec.Subject = "system:serviceaccount:payments:payments-api"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit audience override", func() {
				input := validResource()
				input.Spec.Audience = proto.String("api://AzureADTokenExchange")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a name at the 3-character minimum", func() {
				input := validResource()
				input.Spec.Name = "cid"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_federated_identity_credential", func() {

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

			ginkgo.It("should return a validation error when name exceeds 120 characters", func() {
				input := validResource()
				long := make([]byte, 121)
				for i := range long {
					long[i] = 'a'
				}
				input.Spec.Name = string(long)
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when the parent identity is missing", func() {
				input := validResource()
				input.Spec.UserAssignedIdentity = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when issuer is missing", func() {
				input := validResource()
				input.Spec.Issuer = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when subject is missing", func() {
				input := validResource()
				input.Spec.Subject = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when an explicit audience is empty", func() {
				input := validResource()
				input.Spec.Audience = proto.String("")
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
