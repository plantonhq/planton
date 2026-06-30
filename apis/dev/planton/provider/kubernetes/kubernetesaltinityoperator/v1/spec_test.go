package kubernetesaltinityoperatorv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesAltinityOperatorSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesAltinityOperatorSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesAltinityOperatorSpec validations", func() {
	var spec *KubernetesAltinityOperatorSpec

	ginkgo.BeforeEach(func() {
		spec = &KubernetesAltinityOperatorSpec{
			TargetCluster: &kubernetes.KubernetesClusterSelector{
				ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
				ClusterName: "my-k8s-cluster",
			},
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "kubernetes-altinity-operator",
				},
			},
			Container: &KubernetesAltinityOperatorSpecContainer{
				Resources: &kubernetes.ContainerResources{
					Limits: &kubernetes.CpuMemory{
						Cpu:    "1000m",
						Memory: "1Gi",
					},
					Requests: &kubernetes.CpuMemory{
						Cpu:    "100m",
						Memory: "256Mi",
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

		ginkgo.Context("spec with valid namespace pattern", func() {
			ginkgo.It("should not return a validation error for lowercase with hyphens", func() {
				spec.Namespace = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-operator-namespace",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for single word", func() {
				spec.Namespace = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "operator",
					},
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for namespace with numbers", func() {
				spec.Namespace = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "operator-v2",
					},
				}
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
				spec.Container.Resources.Requests.Cpu = "50m"
				spec.Container.Resources.Requests.Memory = "128Mi"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with namespace creation flag", func() {
			ginkgo.It("should validate successfully when create_namespace is true", func() {
				spec.CreateNamespace = true
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should validate successfully when create_namespace is false", func() {
				spec.CreateNamespace = false
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should validate successfully when create_namespace is not set (default)", func() {
				// create_namespace defaults to false when not explicitly set
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("spec with missing required container field", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Container = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
