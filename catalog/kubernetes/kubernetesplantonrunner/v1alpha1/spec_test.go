package kubernetesplantonrunnerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesPlantonRunnerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesPlantonRunnerSpec Validation Tests")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func stringPtr(v string) *string { return &v }

// minimalValidRunner is the common case: a runner in a namespace with its
// runner token supplied (in real deployments the token arrives as a
// managed-secret reference; validation sees the resolved value).
func minimalValidRunner() *KubernetesPlantonRunner {
	return &KubernetesPlantonRunner{
		ApiVersion: "kubernetes.planton.dev/v1alpha1",
		Kind:       "KubernetesPlantonRunner",
		Metadata: &shared.CloudResourceMetadata{
			Name: "cluster-runner",
		},
		Spec: &KubernetesPlantonRunnerSpec{
			Namespace:       literalRef("planton-runner"),
			CreateNamespace: true,
			Token:           "prt_FAKE_PLACEHOLDER_VALUE",
		},
	}
}

var _ = ginkgo.Describe("KubernetesPlantonRunnerSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("kubernetes_planton_runner", func() {

			ginkgo.It("should not return a validation error for a minimal runner", func() {
				err := protovalidate.Validate(minimalValidRunner())
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit runner name", func() {
				input := minimalValidRunner()
				input.Spec.RunnerName = "prod-cluster-runner"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a self-hosted control plane endpoint", func() {
				input := minimalValidRunner()
				input.Spec.ControlPlaneEndpoint = "planton.example.com:443"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an exact chart version pin", func() {
				input := minimalValidRunner()
				input.Spec.ChartVersion = "0.4.0"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a mirrored chart repository", func() {
				input := minimalValidRunner()
				input.Spec.ChartRepository = stringPtr("oci://asia-south1-docker.pkg.dev/plantonhq/charts")
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept container sizing overrides", func() {
				input := minimalValidRunner()
				input.Spec.Resources = &kubernetes.ContainerResources{
					Requests: &kubernetes.CpuMemory{Cpu: "250m", Memory: "512Mi"},
					Limits:   &kubernetes.CpuMemory{Cpu: "2", Memory: "2Gi"},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the build worker with a Tekton namespace", func() {
				input := minimalValidRunner()
				input.Spec.Build = &KubernetesPlantonRunnerBuild{
					Enabled:         true,
					TektonNamespace: "tekton-pipelines",
				}
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
		ginkgo.Context("kubernetes_planton_runner", func() {

			ginkgo.It("should return an error when the namespace is missing", func() {
				input := minimalValidRunner()
				input.Spec.Namespace = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when the token is missing", func() {
				input := minimalValidRunner()
				input.Spec.Token = ""
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an uppercase runner name", func() {
				input := minimalValidRunner()
				input.Spec.RunnerName = "Prod-Runner"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a chart repository that is not an oci:// path", func() {
				for _, repo := range []string{"ghcr.io/plantonhq/charts", "oci://ghcr.io/plantonhq/charts/"} {
					input := minimalValidRunner()
					input.Spec.ChartRepository = stringPtr(repo)
					gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil(), repo)
				}
			})

			ginkgo.It("should reject a chart version range", func() {
				input := minimalValidRunner()
				input.Spec.ChartVersion = ">=0.4.0"
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an invalid Tekton namespace", func() {
				input := minimalValidRunner()
				input.Spec.Build = &KubernetesPlantonRunnerBuild{
					Enabled:         true,
					TektonNamespace: "Tekton_Pipelines",
				}
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
