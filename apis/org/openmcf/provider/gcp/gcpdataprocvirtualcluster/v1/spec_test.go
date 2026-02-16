package gcpdataprocvirtualclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDataprocVirtualClusterSpec Suite")
}

var _ = ginkgo.Describe("GcpDataprocVirtualClusterSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper for StringValueOrRef with a literal value.
	svr := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	// Helper to build a minimal valid GcpDataprocVirtualCluster.
	minimal := func() *GcpDataprocVirtualCluster {
		return &GcpDataprocVirtualCluster{
			ApiVersion: "gcp.openmcf.org/v1",
			Kind:       "GcpDataprocVirtualCluster",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-virtual-cluster",
			},
			Spec: &GcpDataprocVirtualClusterSpec{
				ProjectId:        svr("my-gcp-project"),
				Region:           "us-central1",
				GkeClusterTarget: svr("projects/my-gcp-project/locations/us-central1/clusters/my-gke-cluster"),
				SoftwareConfig: &GcpDataprocVirtualClusterSoftwareConfig{
					ComponentVersion: map[string]string{
						"SPARK": "3.5-dataproc-17",
					},
				},
				NodePoolTargets: []*GcpDataprocVirtualClusterNodePoolTarget{
					{
						NodePool: svr("default-pool"),
						Roles:    []string{"DEFAULT"},
					},
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept cluster_name with valid format", func() {
		msg := minimal()
		msg.Spec.ClusterName = "my-spark-on-gke-2026"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept empty cluster_name (uses metadata.name)", func() {
		msg := minimal()
		msg.Spec.ClusterName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept kubernetes_namespace as StringValueOrRef", func() {
		msg := minimal()
		msg.Spec.KubernetesNamespace = svr("spark-workloads")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept staging_bucket as StringValueOrRef", func() {
		msg := minimal()
		msg.Spec.StagingBucket = svr("my-staging-bucket")
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept multiple node pool targets with different roles", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets = []*GcpDataprocVirtualClusterNodePoolTarget{
			{
				NodePool: svr("default-pool"),
				Roles:    []string{"DEFAULT"},
			},
			{
				NodePool: svr("controller-pool"),
				Roles:    []string{"CONTROLLER"},
			},
			{
				NodePool: svr("driver-pool"),
				Roles:    []string{"SPARK_DRIVER"},
			},
			{
				NodePool: svr("executor-pool"),
				Roles:    []string{"SPARK_EXECUTOR"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept node pool target with node_pool_config", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets[0].NodePoolConfig = &GcpDataprocVirtualClusterNodePoolConfig{
			Locations:   []string{"us-central1-a", "us-central1-b"},
			MachineType: "n1-standard-4",
			Autoscaling: &GcpDataprocVirtualClusterNodePoolAutoscaling{
				MinNodeCount: 1,
				MaxNodeCount: 10,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept node pool config with spot VMs", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets[0].NodePoolConfig = &GcpDataprocVirtualClusterNodePoolConfig{
			Locations: []string{"us-central1-a"},
			Spot:      true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept node pool config with preemptible VMs", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets[0].NodePoolConfig = &GcpDataprocVirtualClusterNodePoolConfig{
			Locations:   []string{"us-central1-a"},
			Preemptible: true,
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept software config with additional properties", func() {
		msg := minimal()
		msg.Spec.SoftwareConfig.Properties = map[string]string{
			"spark:spark.kubernetes.container.image": "custom-image:latest",
			"spark:spark.executor.instances":         "4",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept auxiliary_services_config with metastore", func() {
		msg := minimal()
		msg.Spec.AuxiliaryServicesConfig = &GcpDataprocVirtualClusterAuxiliaryServicesConfig{
			MetastoreService: "projects/my-project/locations/us-central1/services/my-metastore",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept auxiliary_services_config with spark history server", func() {
		msg := minimal()
		msg.Spec.AuxiliaryServicesConfig = &GcpDataprocVirtualClusterAuxiliaryServicesConfig{
			SparkHistoryServerCluster: "projects/my-project/regions/us-central1/clusters/my-history-server",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a full-featured spec", func() {
		msg := minimal()
		msg.Spec.ClusterName = "production-spark-on-gke"
		msg.Spec.KubernetesNamespace = svr("spark-production")
		msg.Spec.StagingBucket = svr("spark-staging-bucket")
		msg.Spec.SoftwareConfig.Properties = map[string]string{
			"spark:spark.executor.memory": "4g",
		}
		msg.Spec.NodePoolTargets = []*GcpDataprocVirtualClusterNodePoolTarget{
			{
				NodePool: svr("default-pool"),
				Roles:    []string{"DEFAULT"},
				NodePoolConfig: &GcpDataprocVirtualClusterNodePoolConfig{
					Locations:   []string{"us-central1-a", "us-central1-b"},
					MachineType: "n1-standard-8",
					Autoscaling: &GcpDataprocVirtualClusterNodePoolAutoscaling{
						MinNodeCount: 2,
						MaxNodeCount: 20,
					},
				},
			},
			{
				NodePool: svr("executor-pool"),
				Roles:    []string{"SPARK_EXECUTOR"},
				NodePoolConfig: &GcpDataprocVirtualClusterNodePoolConfig{
					Locations:   []string{"us-central1-a", "us-central1-b"},
					MachineType: "n1-highmem-16",
					Spot:        true,
					Autoscaling: &GcpDataprocVirtualClusterNodePoolAutoscaling{
						MinNodeCount: 0,
						MaxNodeCount: 50,
					},
				},
			},
		}
		msg.Spec.AuxiliaryServicesConfig = &GcpDataprocVirtualClusterAuxiliaryServicesConfig{
			MetastoreService:          "projects/my-project/locations/us-central1/services/hive-metastore",
			SparkHistoryServerCluster: "projects/my-project/regions/us-central1/clusters/spark-history",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept autoscaling with min_node_count = 0 (scale-to-zero)", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets[0].NodePoolConfig = &GcpDataprocVirtualClusterNodePoolConfig{
			Locations: []string{"us-central1-a"},
			Autoscaling: &GcpDataprocVirtualClusterNodePoolAutoscaling{
				MinNodeCount: 0,
				MaxNodeCount: 5,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept autoscaling with min == max (fixed size)", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets[0].NodePoolConfig = &GcpDataprocVirtualClusterNodePoolConfig{
			Locations: []string{"us-central1-a"},
			Autoscaling: &GcpDataprocVirtualClusterNodePoolAutoscaling{
				MinNodeCount: 3,
				MaxNodeCount: 3,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept node pool with multiple roles", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets[0].Roles = []string{"DEFAULT", "CONTROLLER"}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject missing project_id", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject empty region", func() {
		msg := minimal()
		msg.Spec.Region = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing gke_cluster_target", func() {
		msg := minimal()
		msg.Spec.GkeClusterTarget = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing software_config", func() {
		msg := minimal()
		msg.Spec.SoftwareConfig = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject empty node_pool_targets", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject node pool target with empty roles", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets[0].Roles = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid role value", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets[0].Roles = []string{"INVALID_ROLE"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject node pool target with missing node_pool", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets[0].NodePool = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cluster_name starting with a digit", func() {
		msg := minimal()
		msg.Spec.ClusterName = "1-invalid-name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cluster_name with uppercase letters", func() {
		msg := minimal()
		msg.Spec.ClusterName = "My-Cluster"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cluster_name ending with a hyphen", func() {
		msg := minimal()
		msg.Spec.ClusterName = "my-cluster-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject autoscaling with max_node_count < min_node_count", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets[0].NodePoolConfig = &GcpDataprocVirtualClusterNodePoolConfig{
			Locations: []string{"us-central1-a"},
			Autoscaling: &GcpDataprocVirtualClusterNodePoolAutoscaling{
				MinNodeCount: 10,
				MaxNodeCount: 5,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject autoscaling with negative min_node_count", func() {
		msg := minimal()
		msg.Spec.NodePoolTargets[0].NodePoolConfig = &GcpDataprocVirtualClusterNodePoolConfig{
			Locations: []string{"us-central1-a"},
			Autoscaling: &GcpDataprocVirtualClusterNodePoolAutoscaling{
				MinNodeCount: -1,
				MaxNodeCount: 5,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong.version/v1"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong kind", func() {
		msg := minimal()
		msg.Kind = "WrongKind"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing metadata", func() {
		msg := minimal()
		msg.Metadata = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing spec", func() {
		msg := minimal()
		msg.Spec = nil
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
