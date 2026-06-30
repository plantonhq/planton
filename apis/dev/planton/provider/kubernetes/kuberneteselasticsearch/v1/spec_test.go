package kuberneteselasticsearchv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestKubernetesElasticsearch(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesElasticsearch Suite")
}

var _ = ginkgo.Describe("KubernetesElasticsearch Custom Validation Tests", func() {
	var input *KubernetesElasticsearch

	ginkgo.BeforeEach(func() {
		input = &KubernetesElasticsearch{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesElasticsearch",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-es",
			},
			Spec: &KubernetesElasticsearchSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				Elasticsearch: &KubernetesElasticsearchElasticsearchSpec{
					Container: &KubernetesElasticsearchElasticsearchContainer{
						Replicas:           1,
						PersistenceEnabled: true,
						DiskSize:           "10Gi", // valid format
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
					Ingress: &KubernetesElasticsearchIngress{
						Enabled:  true,
						Hostname: "elasticsearch.example.com",
					},
				},
				Kibana: &KubernetesElasticsearchKibanaSpec{
					Enabled: true,
					Container: &KubernetesElasticsearchKibanaContainer{
						Replicas: 1,
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
					Ingress: &KubernetesElasticsearchIngress{
						Enabled:  true,
						Hostname: "kibana.example.com",
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("elasticsearch_kubernetes", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
