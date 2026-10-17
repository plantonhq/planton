package kubernetesclickhousev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestKubernetesClickHouseSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesClickHouseSpec Validation Suite")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

var _ = ginkgo.Describe("KubernetesClickHouseSpec validations", func() {
	var spec *KubernetesClickHouseSpec

	ginkgo.BeforeEach(func() {
		spec = &KubernetesClickHouseSpec{
			Namespace: literal("analytics"),
			Version:   "25.3",
			DiskSize:  "50Gi",
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.It("accepts a minimal single-node spec", func() {
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a replicated topology with coordination unset (auto managed keeper)", func() {
			spec.Shards = int32Ptr(2)
			spec.Replicas = int32Ptr(3)
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts an explicit managed keeper with a valid quorum size", func() {
			spec.Replicas = int32Ptr(2)
			spec.Coordination = &KubernetesClickHouseCoordination{
				Type: KubernetesClickHouseCoordination_managed_keeper,
				Keeper: &KubernetesClickHouseManagedKeeper{
					Replicas: int32Ptr(3),
					DiskSize: stringPtr("20Gi"),
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts external zookeeper coordination with nodes", func() {
			spec.Replicas = int32Ptr(2)
			spec.Coordination = &KubernetesClickHouseCoordination{
				Type: KubernetesClickHouseCoordination_external_zookeeper,
				External: &KubernetesClickHouseExternalCoordination{
					Nodes: []*KubernetesClickHouseCoordinationNode{
						{Host: "zk-0.zk-headless.zoo.svc.cluster.local"},
					},
					Root: "/clickhouse/analytics",
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts coordination type none on a single-replica multi-shard topology", func() {
			spec.Shards = int32Ptr(4)
			spec.Coordination = &KubernetesClickHouseCoordination{
				Type: KubernetesClickHouseCoordination_none,
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts users with passwords, profiles and quotas", func() {
			spec.Users = []*KubernetesClickHouseUser{
				{
					Name:     "analyst",
					Password: literal("s3cret-pass"),
					Profile:  "readonly",
					Quota:    "default",
					Networks: []string{"10.0.0.0/8"},
					Grants:   []string{"GRANT SELECT ON analytics.*"},
				},
			}
			spec.Profiles = []*KubernetesClickHouseNamedSettings{
				{Name: "readonly", Settings: map[string]string{"readonly": "1"}},
			}
			spec.Quotas = []*KubernetesClickHouseNamedSettings{
				{Name: "default", Settings: map[string]string{"interval/duration": "3600"}},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})

		ginkgo.It("accepts a user without a password (the module generates it)", func() {
			spec.Users = []*KubernetesClickHouseUser{
				{Name: "ingest", Networks: []string{"0.0.0.0/0", "::/0"}},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.It("rejects a missing version", func() {
			spec.Version = ""
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a malformed disk size", func() {
			spec.DiskSize = "100 gigabytes"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a malformed log disk size", func() {
			spec.LogDiskSize = "ten-gigs"
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a cluster name over the operator's 15-character cap", func() {
			spec.ClusterName = stringPtr("a-very-long-cluster-name")
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a cluster name with invalid characters", func() {
			spec.ClusterName = stringPtr("Main_Cluster")
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects coordination type none when replicas > 1", func() {
			spec.Replicas = int32Ptr(3)
			spec.Coordination = &KubernetesClickHouseCoordination{
				Type: KubernetesClickHouseCoordination_none,
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects external coordination without nodes", func() {
			spec.Coordination = &KubernetesClickHouseCoordination{
				Type: KubernetesClickHouseCoordination_external_keeper,
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an even managed keeper ensemble size", func() {
			spec.Coordination = &KubernetesClickHouseCoordination{
				Type:   KubernetesClickHouseCoordination_managed_keeper,
				Keeper: &KubernetesClickHouseManagedKeeper{Replicas: int32Ptr(2)},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a coordination node port out of range", func() {
			spec.Coordination = &KubernetesClickHouseCoordination{
				Type: KubernetesClickHouseCoordination_external_keeper,
				External: &KubernetesClickHouseExternalCoordination{
					Nodes: []*KubernetesClickHouseCoordinationNode{
						{Host: "keeper.example.com", Port: int32Ptr(70000)},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a user name with invalid characters", func() {
			spec.Users = []*KubernetesClickHouseUser{
				{Name: "bad user!", Password: literal("pw")},
			}
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects zero shards", func() {
			spec.Shards = int32Ptr(0)
			gomega.Expect(protovalidate.Validate(spec)).ToNot(gomega.BeNil())
		})
	})
})

func int32Ptr(v int32) *int32 { return &v }

func stringPtr(v string) *string { return &v }
