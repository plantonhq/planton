package azureplantonrunnerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzurePlantonRunnerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzurePlantonRunnerSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func float64Ptr(v float64) *float64 { return &v }

func stringPtr(v string) *string { return &v }

// minimalValidRunner is the common case: a runner in a referenced
// environment with its runner token supplied (in real deployments the
// token arrives as a managed-secret reference; validation sees the
// resolved value).
func minimalValidRunner() *AzurePlantonRunner {
	return &AzurePlantonRunner{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzurePlantonRunner",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "vnet-runner",
		},
		Spec: &AzurePlantonRunnerSpec{
			ResourceGroup:             literalRef("my-resource-group"),
			ContainerAppEnvironmentId: literalRef("/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-resource-group/providers/Microsoft.App/managedEnvironments/my-env"),
			Token:                     "prt_FAKE_PLACEHOLDER_VALUE",
		},
	}
}

var _ = ginkgo.Describe("AzurePlantonRunnerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_planton_runner", func() {

			ginkgo.It("should not return a validation error for a minimal runner", func() {
				err := protovalidate.Validate(minimalValidRunner())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the default sizing pair (0.5 cpu / 1Gi)", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = float64Ptr(0.5)
				input.Spec.Memory = stringPtr("1Gi")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the smallest sizing pair (0.25 cpu / 0.5Gi)", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = float64Ptr(0.25)
				input.Spec.Memory = stringPtr("0.5Gi")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the largest sizing pair (2 cpu / 4Gi)", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = float64Ptr(2.0)
				input.Spec.Memory = stringPtr("4Gi")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a self-hosted control plane endpoint", func() {
				input := minimalValidRunner()
				input.Spec.ControlPlaneEndpoint = "planton.example.com:443"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a pinned runner version and a mirrored repository", func() {
				input := minimalValidRunner()
				input.Spec.RunnerVersion = stringPtr("v0.3.5")
				input.Spec.ImageRepository = stringPtr("mirror.example.com/planton/runner")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_planton_runner", func() {

			ginkgo.It("should return an error when the resource group is missing", func() {
				input := minimalValidRunner()
				input.Spec.ResourceGroup = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when the environment is missing", func() {
				input := minimalValidRunner()
				input.Spec.ContainerAppEnvironmentId = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when the token is missing", func() {
				input := minimalValidRunner()
				input.Spec.Token = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a cpu value outside the Consumption sizes", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = float64Ptr(0.3)
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a mismatched cpu/memory pairing", func() {
				input := minimalValidRunner()
				input.Spec.Cpu = float64Ptr(0.5)
				input.Spec.Memory = stringPtr("2Gi")
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a control plane endpoint carrying a scheme prefix", func() {
				input := minimalValidRunner()
				input.Spec.ControlPlaneEndpoint = "https://planton.example.com:443"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a control plane endpoint without a port", func() {
				input := minimalValidRunner()
				input.Spec.ControlPlaneEndpoint = "planton.example.com"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
