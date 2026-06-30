package kuberneteshelmreleasev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	cloudresourcekind "github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesHelmRelease(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesHelmRelease Suite")
}

var _ = ginkgo.Describe("KubernetesHelmRelease Custom Validation Tests", func() {
	var input *KubernetesHelmRelease

	ginkgo.BeforeEach(func() {
		input = &KubernetesHelmRelease{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesHelmRelease",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-helmrelease",
			},
			Spec: &KubernetesHelmReleaseSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				Repo:    "https://charts.helm.sh/stable",
				Name:    "nginx-ingress",
				Version: "1.41.3",
				Values: map[string]string{
					"someKey": "someValue",
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("kuberneteshelmrelease", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
