package kubernetesargocdv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesArgocd(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesArgocd Suite")
}

var _ = ginkgo.Describe("KubernetesArgocd Custom Validation Tests", func() {
	var input *KubernetesArgocd

	ginkgo.BeforeEach(func() {
		input = &KubernetesArgocd{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesArgocd",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-argocd",
			},
			Spec: &KubernetesArgocdSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				Container: &KubernetesArgocdContainer{},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("argocd_kubernetes", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
