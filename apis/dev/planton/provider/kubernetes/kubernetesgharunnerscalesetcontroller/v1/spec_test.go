package kubernetesgharunnerscalesetcontrollerv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesGhaRunnerScaleSetControllerSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGhaRunnerScaleSetControllerSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesGhaRunnerScaleSetControllerSpec Validation Tests", func() {
	var spec *KubernetesGhaRunnerScaleSetControllerSpec

	ginkgo.BeforeEach(func() {
		spec = &KubernetesGhaRunnerScaleSetControllerSpec{
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "arc-system",
				},
			},
			Container: &KubernetesGhaRunnerScaleSetControllerContainer{
				Resources: &kubernetes.ContainerResources{
					Requests: &kubernetes.CpuMemory{
						Cpu:    "100m",
						Memory: "128Mi",
					},
					Limits: &kubernetes.CpuMemory{
						Cpu:    "500m",
						Memory: "512Mi",
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with minimal configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with default container resources", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Container = &KubernetesGhaRunnerScaleSetControllerContainer{}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with flags configured", func() {
			ginkgo.It("should not return a validation error", func() {
				concurrentReconciles := int32(5)
				spec.Flags = &KubernetesGhaRunnerScaleSetControllerFlags{
					LogLevel:                      KubernetesGhaRunnerScaleSetControllerFlags_info,
					LogFormat:                     KubernetesGhaRunnerScaleSetControllerFlags_json,
					RunnerMaxConcurrentReconciles: &concurrentReconciles,
					UpdateStrategy:                KubernetesGhaRunnerScaleSetControllerFlags_eventual,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with watch single namespace", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Flags = &KubernetesGhaRunnerScaleSetControllerFlags{
					WatchSingleNamespace: "runners",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with metrics enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Metrics = &KubernetesGhaRunnerScaleSetControllerMetrics{
					ControllerManagerAddr: ":8080",
					ListenerAddr:          ":8080",
					ListenerEndpoint:      "/metrics",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with image pull secrets", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.ImagePullSecrets = []string{"ghcr-secret", "docker-secret"}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with priority class name", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.PriorityClassName = "system-cluster-critical"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with custom image", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Container.Image = &KubernetesGhaRunnerScaleSetControllerImage{
					Repository: "ghcr.io/custom/controller",
					Tag:        "v1.0.0",
					PullPolicy: "IfNotPresent",
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with create namespace enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.CreateNamespace = true
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with multiple replicas", func() {
			ginkgo.It("should not return a validation error", func() {
				replicaCount := int32(3)
				spec.ReplicaCount = &replicaCount
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("without namespace", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Namespace = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without container", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Container = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
