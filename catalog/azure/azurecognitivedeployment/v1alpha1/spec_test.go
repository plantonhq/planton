package azurecognitivedeploymentv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureCognitiveDeploymentSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureCognitiveDeploymentSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const testAccountId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.CognitiveServices/accounts/openai-prod"

// validResource returns a minimal valid mini-model deployment that
// individual cases mutate into the shape under test.
func validResource() *AzureCognitiveDeployment {
	return &AzureCognitiveDeployment{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureCognitiveDeployment",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-cognitive-deployment",
		},
		Spec: &AzureCognitiveDeploymentSpec{
			CognitiveAccountId: literal(testAccountId),
			Name:               "chat",
			Model: &AzureCognitiveDeploymentModel{
				Format: "OpenAI",
				Name:   "gpt-5.4-mini",
			},
			Sku: &AzureCognitiveDeploymentSku{
				Name: "GlobalStandard",
			},
		},
	}
}

var _ = ginkgo.Describe("AzureCognitiveDeploymentSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_cognitive_deployment", func() {

			ginkgo.It("should not return a validation error for a minimal GlobalStandard deployment", func() {
				err := protovalidate.Validate(validResource())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a pinned model version with explicit capacity", func() {
				input := validResource()
				input.Spec.Model.Version = "2024-07-18"
				input.Spec.Sku.Capacity = proto.Int32(50)
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a provisioned-throughput deployment with tier and no-auto-upgrade", func() {
				input := validResource()
				input.Spec.Sku = &AzureCognitiveDeploymentSku{
					Name:     "ProvisionedManaged",
					Tier:     AzureCognitiveDeploymentSkuTier_STANDARD,
					Capacity: proto.Int32(100),
				}
				input.Spec.VersionUpgradeOption = AzureCognitiveDeploymentVersionUpgradeOption_NO_AUTO_UPGRADE
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a custom responsible-AI policy selection with dynamic throttling", func() {
				input := validResource()
				input.Spec.RaiPolicyName = "strict-chat"
				input.Spec.DynamicThrottlingEnabled = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_cognitive_deployment", func() {

			ginkgo.It("should reject a missing cognitive account reference", func() {
				input := validResource()
				input.Spec.CognitiveAccountId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing deployment name", func() {
				input := validResource()
				input.Spec.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing model block", func() {
				input := validResource()
				input.Spec.Model = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a model without a format", func() {
				input := validResource()
				input.Spec.Model.Format = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a model without a name", func() {
				input := validResource()
				input.Spec.Model.Name = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a missing sku block", func() {
				input := validResource()
				input.Spec.Sku = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a sku name outside the vocabulary", func() {
				input := validResource()
				input.Spec.Sku.Name = "Regional"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a zero capacity", func() {
				input := validResource()
				input.Spec.Sku.Capacity = proto.Int32(0)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
