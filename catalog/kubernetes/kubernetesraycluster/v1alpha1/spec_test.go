package kubernetesrayclusterv1alpha1

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

func TestKubernetesRayCluster(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesRayCluster Suite")
}

func int32Ptr(i int32) *int32    { return &i }
func boolPtr(b bool) *bool       { return &b }
func stringPtr(s string) *string { return &s }

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

func smallResources() *kubernetes.ContainerResources {
	return &kubernetes.ContainerResources{
		Limits:   &kubernetes.CpuMemory{Cpu: "1000m", Memory: "2Gi"},
		Requests: &kubernetes.CpuMemory{Cpu: "500m", Memory: "1Gi"},
	}
}

var _ = ginkgo.Describe("KubernetesRayCluster Validation Tests", func() {
	var input *KubernetesRayCluster

	ginkgo.BeforeEach(func() {
		input = &KubernetesRayCluster{
			ApiVersion: "kubernetes.planton.dev/v1alpha1",
			Kind:       "KubernetesRayCluster",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "ml-ray",
			},
			Spec: &KubernetesRayClusterSpec{
				Namespace:  literal("ml-platform"),
				RayVersion: "2.52.0",
				Head: &KubernetesRayClusterHead{
					Resources: smallResources(),
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("minimal spec (head only, no worker groups) should not return a validation error", func() {
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("namespace as a reference should be valid", func() {
			input.Spec.Namespace = valueFrom(catalogkind.CatalogKind_KubernetesNamespace, "ml-platform", "spec.name")
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("worker groups with ordered autoscaling bounds should be valid", func() {
			input.Spec.WorkerGroups = []*KubernetesRayClusterWorkerGroup{
				{
					Name:        "cpu",
					Replicas:    int32Ptr(2),
					MinReplicas: int32Ptr(0),
					MaxReplicas: int32Ptr(10),
					Resources:   smallResources(),
				},
			}
			input.Spec.Autoscaling = &KubernetesRayClusterAutoscaling{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("GCS fault tolerance with a literal endpoint and a password secret should be valid", func() {
			input.Spec.GcsFaultTolerance = &KubernetesRayClusterGcsFaultTolerance{
				Enabled:      true,
				RedisAddress: literal("state-valkey.ml-platform.svc.cluster.local:6379"),
				RedisPasswordSecret: &KubernetesRayClusterSecretSelector{
					Name: literal("state-valkey-auth"),
					Key:  "default",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("GCS fault tolerance composing a KubernetesValkey by reference should be valid", func() {
			input.Spec.GcsFaultTolerance = &KubernetesRayClusterGcsFaultTolerance{
				Enabled:      true,
				RedisAddress: valueFrom(catalogkind.CatalogKind_KubernetesValkey, "state-valkey", "status.outputs.kube_endpoint"),
				RedisPasswordSecret: &KubernetesRayClusterSecretSelector{
					Name: valueFrom(catalogkind.CatalogKind_KubernetesValkey, "state-valkey", "status.outputs.password_secret.name"),
					Key:  "default",
				},
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("each auth mode should be valid", func() {
			for _, m := range []string{"token", "disabled"} {
				input.Spec.Auth = &KubernetesRayClusterAuth{Mode: stringPtr(m)}
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			}
		})

		ginkgo.It("full surface should be valid", func() {
			input.Spec.CreateNamespace = true
			input.Spec.Image = "rayproject/ray:2.52.0-gpu"
			input.Spec.Head.ScheduleTasksOnHead = boolPtr(false)
			input.Spec.Head.RayStartParams = map[string]string{"object-store-memory": "1000000000"}
			input.Spec.WorkerGroups = []*KubernetesRayClusterWorkerGroup{
				{Name: "cpu", Replicas: int32Ptr(2), Resources: smallResources()},
				{
					Name: "gpu", Replicas: int32Ptr(1), Resources: &kubernetes.ContainerResources{
						Limits:   &kubernetes.CpuMemory{Cpu: "4000m", Memory: "16Gi"},
						Requests: &kubernetes.CpuMemory{Cpu: "2000m", Memory: "8Gi"},
					},
					Scheduling: &KubernetesRayClusterScheduling{
						NodeSelector: map[string]string{"accelerator": "nvidia"},
					},
				},
			}
			input.Spec.Autoscaling = &KubernetesRayClusterAutoscaling{
				Enabled:            true,
				IdleTimeoutSeconds: int32Ptr(120),
				UpscalingMode:      "Aggressive",
			}
			input.Spec.Suspend = false
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("a namespace-less spec should fail", func() {
			input.Spec.Namespace = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a spec without ray_version should fail", func() {
			input.Spec.RayVersion = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a spec without a head should fail", func() {
			input.Spec.Head = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a head without resources should fail", func() {
			input.Spec.Head = &KubernetesRayClusterHead{}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("duplicate worker group names should fail", func() {
			input.Spec.WorkerGroups = []*KubernetesRayClusterWorkerGroup{
				{Name: "cpu", Resources: smallResources()},
				{Name: "cpu", Resources: smallResources()},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("a worker group without resources should fail", func() {
			input.Spec.WorkerGroups = []*KubernetesRayClusterWorkerGroup{{Name: "cpu"}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an uppercase worker group name should fail", func() {
			input.Spec.WorkerGroups = []*KubernetesRayClusterWorkerGroup{{Name: "CPU", Resources: smallResources()}}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("disordered autoscaling bounds should fail", func() {
			input.Spec.WorkerGroups = []*KubernetesRayClusterWorkerGroup{
				{Name: "cpu", Replicas: int32Ptr(5), MaxReplicas: int32Ptr(2), Resources: smallResources()},
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("GCS fault tolerance without an endpoint should fail", func() {
			input.Spec.GcsFaultTolerance = &KubernetesRayClusterGcsFaultTolerance{Enabled: true}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unsupported auth mode should fail", func() {
			input.Spec.Auth = &KubernetesRayClusterAuth{Mode: stringPtr("mtls")}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("an unsupported upscaling mode should fail", func() {
			input.Spec.Autoscaling = &KubernetesRayClusterAutoscaling{Enabled: true, UpscalingMode: "Turbo"}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
