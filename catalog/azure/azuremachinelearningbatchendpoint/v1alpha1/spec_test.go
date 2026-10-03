package azuremachinelearningbatchendpointv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureMachineLearningBatchEndpointSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureMachineLearningBatchEndpointSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const (
	testWorkspaceId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.MachineLearningServices/workspaces/ml-workspace"
	testIdentityId  = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.ManagedIdentity/userAssignedIdentities/ml-uai"
)

// validResource returns a minimal valid batch endpoint that individual
// cases mutate into the shape under test. auth_mode is deliberately
// UNSET: the platform default (AADToken -- the only mode the batch
// service accepts) is the canonical manifest shape.
func validResource() *AzureMachineLearningBatchEndpoint {
	return &AzureMachineLearningBatchEndpoint{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureMachineLearningBatchEndpoint",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-ml-batch-endpoint",
		},
		Spec: &AzureMachineLearningBatchEndpointSpec{
			WorkspaceId: literal(testWorkspaceId),
			Name:        "nightly-scoring",
			Region:      "eastus",
		},
	}
}

var _ = ginkgo.Describe("AzureMachineLearningBatchEndpointSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_machine_learning_batch_endpoint", func() {

			ginkgo.It("should not return a validation error for a minimal endpoint with auth mode unset", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the explicit AADToken auth mode (the service's only mode)", func() {
				input := validResource()
				input.Spec.AuthMode = "AADToken"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an endpoint without an identity (optional here, unlike the online sibling)", func() {
				input := validResource()
				input.Spec.Identity = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a system-assigned identity", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningBatchEndpointIdentity{
					Type: AzureMachineLearningBatchEndpointIdentityType_SYSTEM_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a user-assigned identity with identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningBatchEndpointIdentity{
					Type:        AzureMachineLearningBatchEndpointIdentityType_USER_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a default-deployment pointer", func() {
				input := validResource()
				input.Spec.DefaultDeploymentName = "production"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept names on ARM's rule: digit start, underscores and hyphens", func() {
				for _, name := range []string{"0endpoint", "nightly_scoring-v2", "a"} {
					input := validResource()
					input.Spec.Name = name
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})

			ginkgo.It("should accept descriptive surface and ARM properties", func() {
				input := validResource()
				input.Spec.Description = "nightly batch scoring endpoint"
				input.Spec.Properties = map[string]string{"team": "ml-platform"}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_machine_learning_batch_endpoint", func() {

			ginkgo.It("should reject a missing workspace reference", func() {
				input := validResource()
				input.Spec.WorkspaceId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject the auth modes the batch service refuses (Key, AMLToken)", func() {
				for _, mode := range []string{"Key", "AMLToken"} {
					input := validResource()
					input.Spec.AuthMode = mode
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil())
				}
			})

			ginkgo.It("should reject an auth mode outside every vocabulary", func() {
				input := validResource()
				input.Spec.AuthMode = "Token"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a user-assigned identity without identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningBatchEndpointIdentity{
					Type: AzureMachineLearningBatchEndpointIdentityType_USER_ASSIGNED,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a system-assigned identity carrying identity ids", func() {
				input := validResource()
				input.Spec.Identity = &AzureMachineLearningBatchEndpointIdentity{
					Type:        AzureMachineLearningBatchEndpointIdentityType_SYSTEM_ASSIGNED,
					IdentityIds: []*foreignkeyv1.StringValueOrRef{literal(testIdentityId)},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject names ARM rejects", func() {
				for _, name := range []string{"-endpoint", "_endpoint", "nightly scoring", "nightly.scoring"} {
					input := validResource()
					input.Spec.Name = name
					err := protovalidate.Validate(input)
					gomega.Expect(err).NotTo(gomega.BeNil())
				}
			})

			ginkgo.It("should reject a default-deployment pointer ARM's name rule rejects", func() {
				input := validResource()
				input.Spec.DefaultDeploymentName = "-production"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing region", func() {
				input := validResource()
				input.Spec.Region = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a wrong kind literal", func() {
				input := validResource()
				input.Kind = "AzureMachineLearningEndpoint"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
