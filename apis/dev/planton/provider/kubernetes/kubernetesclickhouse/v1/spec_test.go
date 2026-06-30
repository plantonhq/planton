package kubernetesclickhousev1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesClickHouseSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesClickHouseSpec Validation Suite")
}

var _ = ginkgo.Describe("KubernetesClickHouseSpec validations", func() {
	var spec *KubernetesClickHouseSpec

	ginkgo.BeforeEach(func() {
		spec = &KubernetesClickHouseSpec{
			TargetCluster: &kubernetes.KubernetesClusterSelector{
				ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
				ClusterName: "test-cluster",
			},
			Namespace: &foreignkeyv1.StringValueOrRef{
				LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
					Value: "test-namespace",
				},
			},
			ClusterName: "test-cluster",
			Container: &KubernetesClickHouseContainer{
				Replicas:           1,
				PersistenceEnabled: true,
				DiskSize:           "8Gi",
				Resources: &kubernetes.ContainerResources{
					Limits: &kubernetes.CpuMemory{
						Cpu:    "1000m",
						Memory: "2Gi",
					},
					Requests: &kubernetes.CpuMemory{
						Cpu:    "100m",
						Memory: "256Mi",
					},
				},
			},
			Logging: &KubernetesClickHouseLoggingConfig{
				Level: KubernetesClickHouseLoggingConfig_information,
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("spec with persistence enabled", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with persistence disabled", func() {
			ginkgo.It("should not return a validation error when disk_size is empty", func() {
				spec.Container.PersistenceEnabled = false
				spec.Container.DiskSize = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with clustering enabled", func() {
			ginkgo.It("should not return a validation error with valid shard and replica counts", func() {
				spec.Cluster = &KubernetesClickHouseClusterConfig{
					IsEnabled:    true,
					ShardCount:   2,
					ReplicaCount: 2,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("spec with persistence enabled but no disk_size", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Container.PersistenceEnabled = true
				spec.Container.DiskSize = ""
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with invalid disk_size format", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Container.PersistenceEnabled = true
				spec.Container.DiskSize = "invalid"
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with zero replicas", func() {
			ginkgo.It("should return a validation error", func() {
				spec.Container.Replicas = 0
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with clustering enabled but invalid counts", func() {
			ginkgo.It("should return a validation error when shard_count is zero", func() {
				spec.Cluster = &KubernetesClickHouseClusterConfig{
					IsEnabled:    true,
					ShardCount:   0,
					ReplicaCount: 2,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when replica_count is zero", func() {
				spec.Cluster = &KubernetesClickHouseClusterConfig{
					IsEnabled:    true,
					ShardCount:   2,
					ReplicaCount: 0,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("spec with clustering disabled", func() {
			ginkgo.It("should not validate cluster counts when clustering is disabled", func() {
				spec.Cluster = &KubernetesClickHouseClusterConfig{
					IsEnabled:    false,
					ShardCount:   0,
					ReplicaCount: 0,
				}
				err := protovalidate.Validate(spec)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})
})
