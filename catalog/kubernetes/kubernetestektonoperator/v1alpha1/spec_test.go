package kubernetestektonoperatorv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
)

func TestKubernetesTektonOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesTektonOperator Suite")
}

var _ = ginkgo.Describe("KubernetesTektonOperator Validation Tests", func() {
	var input *KubernetesTektonOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesTektonOperator{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesTektonOperator",
			Metadata: &shared.CloudResourceMetadata{
				Name: "tekton-operator",
			},
			Spec: &KubernetesTektonOperatorSpec{},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("an empty spec should be valid (the manifest defaults)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("image overrides and resources should be valid", func() {
			input.Spec.OperatorImage = &kubernetes.ContainerImage{
				Repo: "mirror.example.com/tektoncd/operator",
				Tag:  "v0.80.0",
			}
			input.Spec.WebhookImage = &kubernetes.ContainerImage{
				Repo: "mirror.example.com/tektoncd/operator-webhook",
				Tag:  "v0.80.0",
			}
			input.Spec.OperatorResources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "128Mi"},
				Limits:   &kubernetes.CpuMemory{Cpu: "500m", Memory: "512Mi"},
			}
			input.Spec.WebhookResources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "50m", Memory: "64Mi"},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("scheduling and pull secrets should be valid", func() {
			input.Spec.NodeSelector = map[string]string{"role": "platform"}
			input.Spec.Tolerations = []*kubernetes.WorkloadToleration{
				{Key: "platform", Operator: "Exists", Effect: "NoSchedule"},
			}
			input.Spec.ImagePullSecrets = []string{"mirror-pull"}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a registry root with a path should be valid", func() {
			input.Spec.ImageRegistry = "asia-south1-docker.pkg.dev/example-project/ghcr"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing spec should fail", func() {
			input.Spec = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an image_registry with a scheme should fail", func() {
			input.Spec.ImageRegistry = "https://mirror.example.com/ghcr"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an image_registry with a trailing slash should fail", func() {
			input.Spec.ImageRegistry = "mirror.example.com/ghcr/"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a wrong kind constant should fail", func() {
			input.Kind = "TektonOperator"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
