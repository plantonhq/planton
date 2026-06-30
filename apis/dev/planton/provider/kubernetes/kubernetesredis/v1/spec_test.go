package kubernetesredisv1

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

func TestKubernetesRedis(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesRedis Suite")
}

var _ = ginkgo.Describe("KubernetesRedis Custom Validation Tests", func() {
	var input *KubernetesRedis

	ginkgo.BeforeEach(func() {
		input = &KubernetesRedis{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesRedis",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-redis",
			},
			Spec: &KubernetesRedisSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "test-namespace",
					},
				},
				CreateNamespace: true,
				Container: &KubernetesRedisContainer{
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
				Ingress: &KubernetesRedisIngress{
					Enabled:  true,
					Hostname: "redis.example.com",
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("redis_kubernetes with create_namespace true", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("redis_kubernetes with create_namespace false", func() {
			ginkgo.It("should not return a validation error", func() {
				inputWithoutNamespaceCreation := &KubernetesRedis{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesRedis",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-redis-existing-ns",
					},
					Spec: &KubernetesRedisSpec{
						TargetCluster: &kubernetes.KubernetesClusterSelector{
							ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
							ClusterName: "test-cluster",
						},
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "existing-namespace",
							},
						},
						CreateNamespace: false,
						Container: &KubernetesRedisContainer{
							Replicas:           1,
							PersistenceEnabled: true,
							DiskSize:           "10Gi",
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
				err := protovalidate.Validate(inputWithoutNamespaceCreation)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
