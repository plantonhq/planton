package kubernetessolrv1

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

func TestKubernetesSolr(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesSolr Suite")
}

var _ = ginkgo.Describe("KubernetesSolr Custom Validation Tests", func() {
	var input *KubernetesSolr

	ginkgo.BeforeEach(func() {
		input = &KubernetesSolr{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesSolr",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-solr",
			},
			Spec: &KubernetesSolrSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				SolrContainer: &KubernetesSolrSolrContainer{
					Replicas: 1,
					DiskSize: "10Gi",
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
					Image: &kubernetes.ContainerImage{
						Repo: "solr",
						Tag:  "8.7.0",
					},
				},
				ZookeeperContainer: &KubernetesSolrZookeeperContainer{
					Replicas: 1,
					DiskSize: "10Gi",
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
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("solr_kubernetes", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
