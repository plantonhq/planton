package kubernetestektonoperatorv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
)

func TestKubernetesTektonOperatorSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesTektonOperatorSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesTektonOperatorSpec Validation Tests", func() {
	var spec *KubernetesTektonOperatorSpec

	ginkgo.BeforeEach(func() {
		// Note: Tekton Operator uses fixed namespaces managed by the operator:
		// - 'tekton-operator' for the operator
		// - 'tekton-pipelines' for components (Pipelines, Triggers, Dashboard)
		// Therefore, no namespace field is included in the spec.
		spec = &KubernetesTektonOperatorSpec{
			Container: &KubernetesTektonOperatorSpecContainer{
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
			Components: &KubernetesTektonOperatorComponents{
				Pipelines: true,
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with pipelines enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with triggers enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Components = &KubernetesTektonOperatorComponents{
					Triggers: true,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with dashboard enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Components = &KubernetesTektonOperatorComponents{
					Dashboard: true,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with all components enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Components = &KubernetesTektonOperatorComponents{
					Pipelines: true,
					Triggers:  true,
					Dashboard: true,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with default container resources", func() {
			ginkgo.It("should not return a validation error", func() {
				spec.Container = &KubernetesTektonOperatorSpecContainer{}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("without container", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Container = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("without components", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Components = nil
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("with no components enabled", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Components = &KubernetesTektonOperatorComponents{
					Pipelines: false,
					Triggers:  false,
					Dashboard: false,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
