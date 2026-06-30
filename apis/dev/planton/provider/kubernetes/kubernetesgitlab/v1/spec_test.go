package kubernetesgitlabv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesGitlab(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesGitlab Suite")
}

var _ = ginkgo.Describe("KubernetesGitlab Custom Validation Tests", func() {
	var input *KubernetesGitlab

	ginkgo.BeforeEach(func() {
		input = &KubernetesGitlab{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesGitlab",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-gitlab",
			},
			Spec: &KubernetesGitlabSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: 607, // GKE cluster kind
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				CreateNamespace: true,
				Container:       &KubernetesGitlabSpecContainer{},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("gitlab_kubernetes with create_namespace true", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("gitlab_kubernetes with create_namespace false", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.CreateNamespace = false
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Namespace Management", func() {
		ginkgo.Context("when create_namespace is true", func() {
			ginkgo.It("should validate successfully", func() {
				input.Spec.CreateNamespace = true
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when create_namespace is false", func() {
			ginkgo.It("should validate successfully with existing namespace", func() {
				input.Spec.CreateNamespace = false
				input.Spec.Namespace = &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "existing-namespace",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When namespace is missing", func() {
		ginkgo.Context("gitlab_kubernetes without namespace", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})
	})
})
