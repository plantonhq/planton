package gcpdataprocclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpDataprocClusterSpec Suite")
}

var _ = ginkgo.Describe("GcpDataprocClusterSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpDataprocCluster.
	minimal := func() *GcpDataprocCluster {
		return &GcpDataprocCluster{
			ApiVersion: "gcp.planton.dev/v1",
			Kind:       "GcpDataprocCluster",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-dataproc",
			},
			Spec: &GcpDataprocClusterSpec{
				ProjectId: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "my-gcp-project",
					},
				},
				Region:      "us-central1",
				ClusterName: "my-spark-cluster",
			},
		}
	}

	// Helper for StringValueOrRef with a literal value.
	svr := func(v string) *foreignkeyv1.StringValueOrRef {
		return &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec", func() {
		msg := minimal()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept cluster_name at minimum length (2 chars)", func() {
		msg := minimal()
		msg.Spec.ClusterName = "ab"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept cluster_name with hyphens and numbers", func() {
		msg := minimal()
		msg.Spec.ClusterName = "my-spark-cluster-2026"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept cluster_name at max length (55 chars)", func() {
		msg := minimal()
		msg.Spec.ClusterName = "a" + string(make([]byte, 53)) + "z"
		// Build a valid 55-char name: a + 53 lowercase chars + z
		name := "a"
		for i := 0; i < 53; i++ {
			name += "b"
		}
		name += "z"
		msg.Spec.ClusterName = name
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with graceful_decommission_timeout", func() {
		msg := minimal()
		msg.Spec.GracefulDecommissionTimeout = "3600s"
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with cluster_config and master_config", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MasterConfig: &GcpDataprocClusterMasterConfig{
				NumInstances: 1,
				MachineType:  "n2-standard-4",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with worker_config", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			WorkerConfig: &GcpDataprocClusterWorkerConfig{
				NumInstances:    4,
				MachineType:     "n2-standard-8",
				MinNumInstances: 2,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with secondary_worker_config SPOT", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SecondaryWorkerConfig: &GcpDataprocClusterSecondaryWorkerConfig{
				NumInstances:   10,
				Preemptibility: "SPOT",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept all valid preemptibility values", func() {
		for _, p := range []string{"PREEMPTIBLE", "SPOT", "NON_PREEMPTIBLE"} {
			msg := minimal()
			msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
				SecondaryWorkerConfig: &GcpDataprocClusterSecondaryWorkerConfig{
					NumInstances:   5,
					Preemptibility: p,
				},
			}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "preemptibility: %s", p)
		}
	})

	ginkgo.It("should accept spec with software_config", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SoftwareConfig: &GcpDataprocClusterSoftwareConfig{
				ImageVersion:       "2.2-debian12",
				OptionalComponents: []string{"JUPYTER", "DOCKER"},
				Properties: map[string]string{
					"spark:spark.executor.memory": "4g",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with initialization_actions", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			InitializationActions: []*GcpDataprocClusterInitAction{
				{Script: "gs://my-bucket/init.sh", TimeoutSec: 600},
				{Script: "gs://my-bucket/setup.sh"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with gce_config using network", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			GceConfig: &GcpDataprocClusterGceConfig{
				Network:        svr("projects/my-project/global/networks/default"),
				InternalIpOnly: true,
				Tags:           []string{"dataproc", "spark"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with gce_config using subnetwork", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			GceConfig: &GcpDataprocClusterGceConfig{
				Subnetwork:     svr("projects/my-project/regions/us-central1/subnetworks/default"),
				ServiceAccount: svr("dataproc-sa@my-project.iam.gserviceaccount.com"),
				Zone:           "us-central1-a",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with disk_config on master", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MasterConfig: &GcpDataprocClusterMasterConfig{
				DiskConfig: &GcpDataprocClusterDiskConfig{
					BootDiskSizeGb: 200,
					BootDiskType:   "pd-ssd",
					NumLocalSsds:   2,
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept all valid boot_disk_type values", func() {
		for _, t := range []string{"pd-standard", "pd-ssd", "pd-balanced"} {
			msg := minimal()
			msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
				MasterConfig: &GcpDataprocClusterMasterConfig{
					DiskConfig: &GcpDataprocClusterDiskConfig{
						BootDiskType: t,
					},
				},
			}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "disk type: %s", t)
		}
	})

	ginkgo.It("should accept spec with accelerators on master", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MasterConfig: &GcpDataprocClusterMasterConfig{
				Accelerators: []*GcpDataprocClusterAccelerator{
					{AcceleratorType: "nvidia-tesla-t4", AcceleratorCount: 2},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with CMEK encryption", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			EncryptionKmsKeyName: svr("projects/p/locations/l/keyRings/kr/cryptoKeys/k"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with endpoint_config", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			EndpointConfig: &GcpDataprocClusterEndpointConfig{
				EnableHttpPortAccess: true,
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with lifecycle_config idle_delete_ttl", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			LifecycleConfig: &GcpDataprocClusterLifecycleConfig{
				IdleDeleteTtl: "1800s",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with lifecycle_config auto_delete_time", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			LifecycleConfig: &GcpDataprocClusterLifecycleConfig{
				AutoDeleteTime: "2026-03-01T00:00:00Z",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with staging and temp buckets", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			StagingBucket: svr("my-staging-bucket"),
			TempBucket:    svr("my-temp-bucket"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with autoscaling_policy_uri", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			AutoscalingPolicyUri: "projects/my-project/locations/us-central1/autoscalingPolicies/my-policy",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully-featured spec", func() {
		msg := minimal()
		msg.Spec.GracefulDecommissionTimeout = "300s"
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			StagingBucket: svr("my-staging-bucket"),
			TempBucket:    svr("my-temp-bucket"),
			GceConfig: &GcpDataprocClusterGceConfig{
				Subnetwork:     svr("projects/my-project/regions/us-central1/subnetworks/dataproc"),
				ServiceAccount: svr("dataproc-sa@my-project.iam.gserviceaccount.com"),
				InternalIpOnly: true,
				Tags:           []string{"dataproc", "spark"},
				Metadata:       map[string]string{"enable-oslogin": "true"},
			},
			MasterConfig: &GcpDataprocClusterMasterConfig{
				NumInstances: 3,
				MachineType:  "n2-standard-8",
				DiskConfig: &GcpDataprocClusterDiskConfig{
					BootDiskSizeGb: 200,
					BootDiskType:   "pd-ssd",
				},
			},
			WorkerConfig: &GcpDataprocClusterWorkerConfig{
				NumInstances:    5,
				MachineType:     "n2-standard-8",
				MinNumInstances: 2,
				DiskConfig: &GcpDataprocClusterDiskConfig{
					BootDiskSizeGb: 500,
					BootDiskType:   "pd-ssd",
					NumLocalSsds:   2,
				},
				Accelerators: []*GcpDataprocClusterAccelerator{
					{AcceleratorType: "nvidia-tesla-t4", AcceleratorCount: 1},
				},
			},
			SecondaryWorkerConfig: &GcpDataprocClusterSecondaryWorkerConfig{
				NumInstances:   10,
				Preemptibility: "SPOT",
			},
			SoftwareConfig: &GcpDataprocClusterSoftwareConfig{
				ImageVersion:       "2.2-debian12",
				OptionalComponents: []string{"JUPYTER", "DOCKER"},
				Properties: map[string]string{
					"spark:spark.executor.memory":              "8g",
					"hdfs:dfs.replication":                     "2",
					"yarn:yarn.nodemanager.resource.memory-mb": "16384",
				},
			},
			InitializationActions: []*GcpDataprocClusterInitAction{
				{Script: "gs://my-bucket/init.sh", TimeoutSec: 600},
			},
			EncryptionKmsKeyName: svr("projects/p/locations/l/keyRings/kr/cryptoKeys/k"),
			EndpointConfig: &GcpDataprocClusterEndpointConfig{
				EnableHttpPortAccess: true,
			},
			LifecycleConfig: &GcpDataprocClusterLifecycleConfig{
				IdleDeleteTtl: "1800s",
			},
		}
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

	ginkgo.It("should reject missing region", func() {
		msg := minimal()
		msg.Spec.Region = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject missing cluster_name", func() {
		msg := minimal()
		msg.Spec.ClusterName = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cluster_name starting with digit", func() {
		msg := minimal()
		msg.Spec.ClusterName = "1-bad-name"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cluster_name with uppercase", func() {
		msg := minimal()
		msg.Spec.ClusterName = "MyBadCluster"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cluster_name ending with hyphen", func() {
		msg := minimal()
		msg.Spec.ClusterName = "bad-name-"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cluster_name with single char", func() {
		msg := minimal()
		msg.Spec.ClusterName = "a"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject cluster_name with underscores", func() {
		msg := minimal()
		msg.Spec.ClusterName = "bad_name_here"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject invalid preemptibility value", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SecondaryWorkerConfig: &GcpDataprocClusterSecondaryWorkerConfig{
				NumInstances:   5,
				Preemptibility: "INVALID",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("preemptibility"))
	})

	ginkgo.It("should reject invalid boot_disk_type", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MasterConfig: &GcpDataprocClusterMasterConfig{
				DiskConfig: &GcpDataprocClusterDiskConfig{
					BootDiskType: "pd-extreme",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("boot_disk_type"))
	})

	ginkgo.It("should reject boot_disk_size_gb below minimum", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MasterConfig: &GcpDataprocClusterMasterConfig{
				DiskConfig: &GcpDataprocClusterDiskConfig{
					BootDiskSizeGb: 5,
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("boot_disk_size_gb"))
	})

	ginkgo.It("should reject invalid graceful_decommission_timeout format", func() {
		msg := minimal()
		msg.Spec.GracefulDecommissionTimeout = "1h"
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("graceful_decommission_timeout"))
	})

	ginkgo.It("should reject invalid idle_delete_ttl format", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			LifecycleConfig: &GcpDataprocClusterLifecycleConfig{
				IdleDeleteTtl: "30m",
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("idle_delete_ttl"))
	})

	ginkgo.It("should reject accelerator without type", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MasterConfig: &GcpDataprocClusterMasterConfig{
				Accelerators: []*GcpDataprocClusterAccelerator{
					{AcceleratorCount: 2},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject accelerator with zero count", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MasterConfig: &GcpDataprocClusterMasterConfig{
				Accelerators: []*GcpDataprocClusterAccelerator{
					{AcceleratorType: "nvidia-tesla-t4", AcceleratorCount: 0},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject init_action without script", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			InitializationActions: []*GcpDataprocClusterInitAction{
				{TimeoutSec: 300},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject wrong api_version", func() {
		msg := minimal()
		msg.ApiVersion = "wrong/v1"
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
