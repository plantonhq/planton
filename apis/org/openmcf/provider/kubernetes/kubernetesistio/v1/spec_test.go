package kubernetesistiov1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/provider/kubernetes"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestKubernetesIstioSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesIstioSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesIstioSpec validations", func() {
	var spec *KubernetesIstioSpec

	ginkgo.BeforeEach(func() {
		spec = &KubernetesIstioSpec{
			TargetCluster: &kubernetes.KubernetesClusterSelector{
				ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
				ClusterName: "my-k8s-cluster",
			},
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "istio-system",
				},
			},
			Container: &KubernetesIstioSpecContainer{
				Resources: &kubernetes.ContainerResources{
					Limits: &kubernetes.CpuMemory{
						Cpu:    "1000m",
						Memory: "1Gi",
					},
					Requests: &kubernetes.CpuMemory{
						Cpu:    "50m",
						Memory: "100Mi",
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("spec with all fields set", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with custom resource allocations", func() {
			ginkgo.It("should not return a validation error for higher limits", func() {
				spec.Container.Resources.Limits.Cpu = "2000m"
				spec.Container.Resources.Limits.Memory = "4Gi"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for lower requests", func() {
				spec.Container.Resources.Requests.Cpu = "25m"
				spec.Container.Resources.Requests.Memory = "64Mi"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for minimal resources", func() {
				spec.Container.Resources.Requests.Cpu = "10m"
				spec.Container.Resources.Requests.Memory = "16Mi"
				spec.Container.Resources.Limits.Cpu = "100m"
				spec.Container.Resources.Limits.Memory = "128Mi"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for production-grade resources", func() {
				spec.Container.Resources.Requests.Cpu = "500m"
				spec.Container.Resources.Requests.Memory = "512Mi"
				spec.Container.Resources.Limits.Cpu = "4000m"
				spec.Container.Resources.Limits.Memory = "8Gi"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with default container resources", func() {
			ginkgo.It("should not return a validation error when using proto defaults", func() {
				// Even with nil resources, the proto default_container_resources should apply
				spec.Container.Resources = &kubernetes.ContainerResources{
					Limits: &kubernetes.CpuMemory{
						Cpu:    "1000m",
						Memory: "1Gi",
					},
					Requests: &kubernetes.CpuMemory{
						Cpu:    "50m",
						Memory: "100Mi",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with version field", func() {
			ginkgo.It("should accept an omitted version (default applies at manifest load)", func() {
				spec.Version = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an explicit full-patch version (release target)", func() {
				v := "1.26.8"
				spec.Version = &v
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept pinning an older version (sequential-upgrade safety)", func() {
				v := "1.22.3"
				spec.Version = &v
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("spec with malformed version", func() {
			ginkgo.It("should reject a bare minor version (full patch required)", func() {
				v := "1.26"
				spec.Version = &v
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a non-semver version string", func() {
				v := "latest"
				spec.Version = &v
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with missing required container field", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Container = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with invalid target cluster configuration", func() {
			ginkgo.It("should accept spec without target_cluster (optional field)", func() {
				spec.TargetCluster = nil
				err := protovalidate.Validate(spec)
				// target_cluster is optional, so this should still pass
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with create_namespace flag", func() {
			ginkgo.It("should not return error when create_namespace is true", func() {
				spec.CreateNamespace = true
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return error when create_namespace is false", func() {
				spec.CreateNamespace = false
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should handle create_namespace with all other fields set", func() {
				spec.CreateNamespace = true
				spec.Container.Resources.Requests.Cpu = "500m"
				spec.Container.Resources.Requests.Memory = "512Mi"
				spec.Container.Resources.Limits.Cpu = "2000m"
				spec.Container.Resources.Limits.Memory = "4Gi"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
