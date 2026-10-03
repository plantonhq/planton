package kubernetesgharunnerscalesetcontrollerv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	kubernetes "github.com/plantonhq/planton/catalog/kubernetes"
	"github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/catalogkind"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesGhaRunnerScaleSetController(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGhaRunnerScaleSetController Suite")
}

func int32Ptr(i int32) *int32 { return &i }
func strPtr(s string) *string { return &s }

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func valueFrom(kind catalogkind.CatalogKind, name, fieldPath string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{
				Kind:      kind,
				Name:      name,
				FieldPath: fieldPath,
			},
		},
	}
}

var _ = ginkgo.Describe("KubernetesGhaRunnerScaleSetController Validation Tests", func() {
	var input *KubernetesGhaRunnerScaleSetController

	ginkgo.BeforeEach(func() {
		input = &KubernetesGhaRunnerScaleSetController{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesGhaRunnerScaleSetController",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "arc",
			},
			Spec: &KubernetesGhaRunnerScaleSetControllerSpec{
				Namespace: literal("arc-system"),
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("a minimal spec should be valid (chart defaults)", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "arc-system", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("a full flags block should be valid", func() {
			input.Spec.Flags = &KubernetesGhaRunnerScaleSetControllerFlags{
				LogLevel:                        strPtr("info"),
				LogFormat:                       strPtr("json"),
				WatchSingleNamespace:            "ci-runners",
				RunnerMaxConcurrentReconciles:   int32Ptr(8),
				UpdateStrategy:                  strPtr("eventual"),
				ExcludeLabelPropagationPrefixes: []string{"argocd.argoproj.io/instance"},
				K8SClientRateLimiterQps:         int32Ptr(30),
				K8SClientRateLimiterBurst:       int32Ptr(60),
				RateLimiter:                     strPtr("typed_rate_limiter"),
				HealthProbeBindAddress:          ":8081",
				PriorityClassName:               "system-cluster-critical",
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("metrics, image, resources and scheduling should be valid", func() {
			input.Spec.Replicas = int32Ptr(2)
			input.Spec.Image = &kubernetes.ContainerImage{Repo: "mirror.example.com/gha-rs-controller", Tag: "0.14.2"}
			input.Spec.Resources = &kubernetes.ContainerResources{
				Requests: &kubernetes.CpuMemory{Cpu: "100m", Memory: "128Mi"},
			}
			input.Spec.Metrics = &KubernetesGhaRunnerScaleSetControllerMetrics{
				ControllerManagerAddr: ":8080",
				ListenerAddr:          ":8080",
				ListenerEndpoint:      "/metrics",
			}
			input.Spec.ImagePullSecrets = []string{"mirror-pull"}
			input.Spec.Scheduling = &KubernetesGhaRunnerScaleSetControllerScheduling{
				NodeSelector: map[string]string{"role": "platform"},
				Tolerations: []*kubernetes.WorkloadToleration{
					{Key: "platform", Operator: "Exists", Effect: "NoSchedule"},
				},
			}
			input.Spec.HelmValues = "podLabels:\n  team: platform\n"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a missing namespace should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown log level should fail", func() {
			input.Spec.Flags = &KubernetesGhaRunnerScaleSetControllerFlags{LogLevel: strPtr("trace")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown update strategy should fail", func() {
			input.Spec.Flags = &KubernetesGhaRunnerScaleSetControllerFlags{UpdateStrategy: strPtr("rolling")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unknown rate limiter should fail", func() {
			input.Spec.Flags = &KubernetesGhaRunnerScaleSetControllerFlags{RateLimiter: strPtr("token_bucket")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an uppercase watch namespace should fail", func() {
			input.Spec.Flags = &KubernetesGhaRunnerScaleSetControllerFlags{WatchSingleNamespace: "CI"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a metrics block missing the listener endpoint should fail", func() {
			input.Spec.Metrics = &KubernetesGhaRunnerScaleSetControllerMetrics{
				ControllerManagerAddr: ":8080",
				ListenerAddr:          ":8080",
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("zero replicas should fail", func() {
			input.Spec.Replicas = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
