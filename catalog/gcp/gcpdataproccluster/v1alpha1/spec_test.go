package gcpdataprocclusterv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
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
			ApiVersion: "gcp.planton.dev/v1alpha1",
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

	// Helper for a minimal valid virtual (Dataproc-on-GKE) arm.
	virtualArm := func() *GcpDataprocClusterVirtualClusterConfig {
		return &GcpDataprocClusterVirtualClusterConfig{
			KubernetesClusterConfig: &GcpDataprocClusterKubernetesClusterConfig{
				GkeClusterConfig: &GcpDataprocClusterGkeClusterConfig{
					GkeClusterTarget: svr("projects/my-gcp-project/locations/us-central1/clusters/my-gke"),
				},
				KubernetesSoftwareConfig: &GcpDataprocClusterKubernetesSoftwareConfig{
					ComponentVersion: map[string]string{"SPARK": "3.5-dataproc-17"},
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

	ginkgo.It("should accept a spec without project_id (ambient project)", func() {
		msg := minimal()
		msg.Spec.ProjectId = nil
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a multi-digit region", func() {
		msg := minimal()
		msg.Spec.Region = "europe-west12"
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

	ginkgo.It("should accept user labels on the GCE arm", func() {
		msg := minimal()
		msg.Spec.Labels = map[string]string{"team": "data-eng", "cost-center": "1234"}
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{}
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

	ginkgo.It("should accept secondary workers with instance_flexibility_policy", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SecondaryWorkerConfig: &GcpDataprocClusterSecondaryWorkerConfig{
				NumInstances:   20,
				Preemptibility: "SPOT",
				InstanceFlexibilityPolicy: &GcpDataprocClusterInstanceFlexibilityPolicy{
					InstanceSelectionList: []*GcpDataprocClusterInstanceSelection{
						{MachineTypes: []string{"n2-standard-8", "n2d-standard-8"}, Rank: 0},
						{MachineTypes: []string{"e2-standard-8"}, Rank: 1},
					},
					ProvisioningModelMix: &GcpDataprocClusterProvisioningModelMix{
						StandardCapacityBase:             2,
						StandardCapacityPercentAboveBase: 25,
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept all valid cluster_tier values", func() {
		for _, tier := range []string{"CLUSTER_TIER_STANDARD", "CLUSTER_TIER_PREMIUM"} {
			msg := minimal()
			msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{ClusterTier: tier}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "cluster_tier: %s", tier)
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

	ginkgo.It("should accept gce_config with shielded and confidential instance config", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			GceConfig: &GcpDataprocClusterGceConfig{
				ShieldedInstanceConfig: &GcpDataprocClusterShieldedInstanceConfig{
					EnableSecureBoot:          true,
					EnableVtpm:                true,
					EnableIntegrityMonitoring: true,
				},
				ConfidentialInstanceConfig: &GcpDataprocClusterConfidentialInstanceConfig{
					EnableConfidentialCompute: true,
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept reservation_affinity ANY_RESERVATION without key/values", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			GceConfig: &GcpDataprocClusterGceConfig{
				ReservationAffinity: &GcpDataprocClusterReservationAffinity{
					ConsumeReservationType: "ANY_RESERVATION",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept SPECIFIC_RESERVATION with key and values", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			GceConfig: &GcpDataprocClusterGceConfig{
				ReservationAffinity: &GcpDataprocClusterReservationAffinity{
					ConsumeReservationType: "SPECIFIC_RESERVATION",
					Key:                    "compute.googleapis.com/reservation-name",
					Values:                 []string{"my-reservation"},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept node_group_affinity with a node group URI", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			GceConfig: &GcpDataprocClusterGceConfig{
				NodeGroupAffinity: &GcpDataprocClusterNodeGroupAffinity{
					NodeGroupUri: "projects/my-project/zones/us-central1-a/nodeGroups/sole-tenant-group",
				},
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

	ginkgo.It("should accept all valid local_ssd_interface values", func() {
		for _, iface := range []string{"scsi", "nvme"} {
			msg := minimal()
			msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
				WorkerConfig: &GcpDataprocClusterWorkerConfig{
					DiskConfig: &GcpDataprocClusterDiskConfig{
						NumLocalSsds:      2,
						LocalSsdInterface: iface,
					},
				},
			}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "local_ssd_interface: %s", iface)
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

	ginkgo.It("should accept security_config with kerberos_config only", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SecurityConfig: &GcpDataprocClusterSecurityConfig{
				KerberosConfig: &GcpDataprocClusterKerberosConfig{
					EnableKerberos:           true,
					RootPrincipalPasswordUri: "gs://my-secure-bucket/kerberos/root-password.encrypted",
					KmsKeyUri:                svr("projects/p/locations/l/keyRings/kr/cryptoKeys/k"),
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept security_config with identity_config only", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SecurityConfig: &GcpDataprocClusterSecurityConfig{
				IdentityConfig: &GcpDataprocClusterIdentityConfig{
					UserServiceAccountMapping: map[string]string{
						"bob@example.com": "bob-sa@my-project.iam.gserviceaccount.com",
					},
				},
			},
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
			AutoscalingPolicyUri: svr("projects/my-project/locations/us-central1/autoscalingPolicies/my-policy"),
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept spec with metastore_config", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MetastoreConfig: &GcpDataprocClusterMetastoreConfig{
				DataprocMetastoreService: svr("projects/my-project/locations/us-central1/services/shared-hive-metastore"),
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept dataproc_metric_config with valid metric sources", func() {
		for _, source := range []string{"MONITORING_AGENT_DEFAULTS", "HDFS", "SPARK", "YARN", "SPARK_HISTORY_SERVER", "HIVESERVER2"} {
			msg := minimal()
			msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
				DataprocMetricConfig: &GcpDataprocClusterMetricConfig{
					Metrics: []*GcpDataprocClusterMetric{
						{MetricSource: source},
					},
				},
			}
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "metric_source: %s", source)
		}
	})

	ginkgo.It("should accept dataproc_metric_config with metric_overrides", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			DataprocMetricConfig: &GcpDataprocClusterMetricConfig{
				Metrics: []*GcpDataprocClusterMetric{
					{
						MetricSource:    "YARN",
						MetricOverrides: []string{"yarn:ResourceManager:QueueMetrics:AppsCompleted"},
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept an auxiliary node group with the DRIVER role", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			AuxiliaryNodeGroups: []*GcpDataprocClusterAuxiliaryNodeGroup{
				{
					Roles:       []string{"DRIVER"},
					NodeGroupId: "spark-drivers",
					NodeGroupConfig: &GcpDataprocClusterAuxiliaryNodeGroupConfig{
						NumInstances: 2,
						MachineType:  "n2-highmem-8",
						DiskConfig: &GcpDataprocClusterDiskConfig{
							BootDiskSizeGb: 200,
							BootDiskType:   "pd-ssd",
						},
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a fully-featured GCE-arm spec", func() {
		msg := minimal()
		msg.Spec.GracefulDecommissionTimeout = "300s"
		msg.Spec.Labels = map[string]string{"team": "data-eng"}
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			StagingBucket: svr("my-staging-bucket"),
			TempBucket:    svr("my-temp-bucket"),
			ClusterTier:   "CLUSTER_TIER_STANDARD",
			GceConfig: &GcpDataprocClusterGceConfig{
				Subnetwork:     svr("projects/my-project/regions/us-central1/subnetworks/dataproc"),
				ServiceAccount: svr("dataproc-sa@my-project.iam.gserviceaccount.com"),
				InternalIpOnly: true,
				Tags:           []string{"dataproc", "spark"},
				Metadata:       map[string]string{"enable-oslogin": "true"},
				ShieldedInstanceConfig: &GcpDataprocClusterShieldedInstanceConfig{
					EnableSecureBoot:          true,
					EnableVtpm:                true,
					EnableIntegrityMonitoring: true,
				},
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
					BootDiskSizeGb:    500,
					BootDiskType:      "pd-ssd",
					NumLocalSsds:      2,
					LocalSsdInterface: "nvme",
				},
				Accelerators: []*GcpDataprocClusterAccelerator{
					{AcceleratorType: "nvidia-tesla-t4", AcceleratorCount: 1},
				},
			},
			SecondaryWorkerConfig: &GcpDataprocClusterSecondaryWorkerConfig{
				NumInstances:   10,
				Preemptibility: "SPOT",
				InstanceFlexibilityPolicy: &GcpDataprocClusterInstanceFlexibilityPolicy{
					InstanceSelectionList: []*GcpDataprocClusterInstanceSelection{
						{MachineTypes: []string{"n2-standard-8", "n2d-standard-8"}, Rank: 0},
					},
					ProvisioningModelMix: &GcpDataprocClusterProvisioningModelMix{
						StandardCapacityBase:             2,
						StandardCapacityPercentAboveBase: 20,
					},
				},
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
			AutoscalingPolicyUri: svr("projects/my-project/locations/us-central1/autoscalingPolicies/etl-policy"),
			EncryptionKmsKeyName: svr("projects/p/locations/l/keyRings/kr/cryptoKeys/k"),
			EndpointConfig: &GcpDataprocClusterEndpointConfig{
				EnableHttpPortAccess: true,
			},
			LifecycleConfig: &GcpDataprocClusterLifecycleConfig{
				IdleDeleteTtl: "1800s",
			},
			MetastoreConfig: &GcpDataprocClusterMetastoreConfig{
				DataprocMetastoreService: svr("projects/my-project/locations/us-central1/services/shared-hive-metastore"),
			},
			DataprocMetricConfig: &GcpDataprocClusterMetricConfig{
				Metrics: []*GcpDataprocClusterMetric{
					{MetricSource: "SPARK"},
					{MetricSource: "YARN"},
				},
			},
			AuxiliaryNodeGroups: []*GcpDataprocClusterAuxiliaryNodeGroup{
				{
					Roles:       []string{"DRIVER"},
					NodeGroupId: "spark-drivers",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a minimal virtual (Dataproc-on-GKE) spec", func() {
		msg := minimal()
		msg.Spec.VirtualClusterConfig = virtualArm()
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept a virtual spec with node pool targets and auxiliary services", func() {
		msg := minimal()
		arm := virtualArm()
		arm.StagingBucket = svr("my-dataproc-staging")
		arm.KubernetesClusterConfig.KubernetesNamespace = svr("dataproc-workloads")
		arm.KubernetesClusterConfig.GkeClusterConfig.NodePoolTarget = []*GcpDataprocClusterNodePoolTarget{
			{
				NodePool: svr("projects/my-gcp-project/locations/us-central1/clusters/my-gke/nodePools/dataproc-pool"),
				Roles:    []string{"DEFAULT"},
				NodePoolConfig: &GcpDataprocClusterNodePoolConfig{
					Locations: []string{"us-central1-a"},
					Autoscaling: &GcpDataprocClusterNodePoolAutoscaling{
						MinNodeCount: 0,
						MaxNodeCount: 10,
					},
					MachineType: "n2-standard-8",
					Spot:        true,
				},
			},
			{
				NodePool: svr("projects/my-gcp-project/locations/us-central1/clusters/my-gke/nodePools/controller-pool"),
				Roles:    []string{"CONTROLLER", "SPARK_DRIVER"},
			},
		}
		arm.AuxiliaryServicesConfig = &GcpDataprocClusterAuxiliaryServicesConfig{
			MetastoreConfig: &GcpDataprocClusterMetastoreConfig{
				DataprocMetastoreService: svr("projects/my-gcp-project/locations/us-central1/services/shared-hive-metastore"),
			},
			SparkHistoryServerConfig: &GcpDataprocClusterSparkHistoryServerConfig{
				DataprocCluster: svr("projects/my-gcp-project/regions/us-central1/clusters/history-server"),
			},
		}
		msg.Spec.VirtualClusterConfig = arm
		err := validator.Validate(msg)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	ginkgo.It("should accept all valid node pool roles", func() {
		for _, role := range []string{"DEFAULT", "CONTROLLER", "SPARK_DRIVER", "SPARK_EXECUTOR"} {
			msg := minimal()
			arm := virtualArm()
			arm.KubernetesClusterConfig.GkeClusterConfig.NodePoolTarget = []*GcpDataprocClusterNodePoolTarget{
				{
					NodePool: svr("projects/p/locations/l/clusters/c/nodePools/np"),
					Roles:    []string{role},
				},
			}
			msg.Spec.VirtualClusterConfig = arm
			err := validator.Validate(msg)
			gomega.Expect(err).ToNot(gomega.HaveOccurred(), "role: %s", role)
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject missing region", func() {
		msg := minimal()
		msg.Spec.Region = ""
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a zone where a region is expected", func() {
		msg := minimal()
		msg.Spec.Region = "us-central1-a"
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

	ginkgo.It("should reject both deployment arms set together", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{}
		msg.Spec.VirtualClusterConfig = virtualArm()
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("mutually exclusive"))
	})

	ginkgo.It("should reject user labels on the virtual arm", func() {
		msg := minimal()
		msg.Spec.VirtualClusterConfig = virtualArm()
		msg.Spec.Labels = map[string]string{"team": "data-eng"}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("labels"))
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

	ginkgo.It("should reject invalid cluster_tier value", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			ClusterTier: "CLUSTER_TIER_GOLD",
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("cluster_tier"))
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

	ginkgo.It("should reject invalid local_ssd_interface", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			WorkerConfig: &GcpDataprocClusterWorkerConfig{
				DiskConfig: &GcpDataprocClusterDiskConfig{
					NumLocalSsds:      1,
					LocalSsdInterface: "sata",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("local_ssd_interface"))
	})

	ginkgo.It("should reject network and subnetwork set together", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			GceConfig: &GcpDataprocClusterGceConfig{
				Network:    svr("projects/my-project/global/networks/default"),
				Subnetwork: svr("projects/my-project/regions/us-central1/subnetworks/default"),
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("network"))
	})

	ginkgo.It("should reject SPECIFIC_RESERVATION without key and values", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			GceConfig: &GcpDataprocClusterGceConfig{
				ReservationAffinity: &GcpDataprocClusterReservationAffinity{
					ConsumeReservationType: "SPECIFIC_RESERVATION",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("SPECIFIC_RESERVATION"))
	})

	ginkgo.It("should reject invalid consume_reservation_type", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			GceConfig: &GcpDataprocClusterGceConfig{
				ReservationAffinity: &GcpDataprocClusterReservationAffinity{
					ConsumeReservationType: "SOME_RESERVATION",
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject node_group_affinity without node_group_uri", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			GceConfig: &GcpDataprocClusterGceConfig{
				NodeGroupAffinity: &GcpDataprocClusterNodeGroupAffinity{},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject instance_selection with empty machine_types", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SecondaryWorkerConfig: &GcpDataprocClusterSecondaryWorkerConfig{
				InstanceFlexibilityPolicy: &GcpDataprocClusterInstanceFlexibilityPolicy{
					InstanceSelectionList: []*GcpDataprocClusterInstanceSelection{
						{MachineTypes: []string{}, Rank: 0},
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject provisioning_model_mix percent above 100", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SecondaryWorkerConfig: &GcpDataprocClusterSecondaryWorkerConfig{
				InstanceFlexibilityPolicy: &GcpDataprocClusterInstanceFlexibilityPolicy{
					ProvisioningModelMix: &GcpDataprocClusterProvisioningModelMix{
						StandardCapacityPercentAboveBase: 120,
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject security_config with both kerberos and identity config", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SecurityConfig: &GcpDataprocClusterSecurityConfig{
				KerberosConfig: &GcpDataprocClusterKerberosConfig{
					RootPrincipalPasswordUri: "gs://my-secure-bucket/kerberos/root-password.encrypted",
					KmsKeyUri:                svr("projects/p/locations/l/keyRings/kr/cryptoKeys/k"),
				},
				IdentityConfig: &GcpDataprocClusterIdentityConfig{
					UserServiceAccountMapping: map[string]string{
						"bob@example.com": "bob-sa@my-project.iam.gserviceaccount.com",
					},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("exactly one"))
	})

	ginkgo.It("should reject security_config with neither mechanism set", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SecurityConfig: &GcpDataprocClusterSecurityConfig{},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject kerberos_config without root_principal_password_uri", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SecurityConfig: &GcpDataprocClusterSecurityConfig{
				KerberosConfig: &GcpDataprocClusterKerberosConfig{
					KmsKeyUri: svr("projects/p/locations/l/keyRings/kr/cryptoKeys/k"),
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject identity_config with an empty mapping", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SecurityConfig: &GcpDataprocClusterSecurityConfig{
				IdentityConfig: &GcpDataprocClusterIdentityConfig{
					UserServiceAccountMapping: map[string]string{},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject metastore_config without a service", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MetastoreConfig: &GcpDataprocClusterMetastoreConfig{},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid metric_source", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			DataprocMetricConfig: &GcpDataprocClusterMetricConfig{
				Metrics: []*GcpDataprocClusterMetric{
					{MetricSource: "FLINK"},
				},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("metric_source"))
	})

	ginkgo.It("should reject dataproc_metric_config with no metrics", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			DataprocMetricConfig: &GcpDataprocClusterMetricConfig{
				Metrics: []*GcpDataprocClusterMetric{},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an auxiliary node group role other than DRIVER", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			AuxiliaryNodeGroups: []*GcpDataprocClusterAuxiliaryNodeGroup{
				{Roles: []string{"WORKER"}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an auxiliary node group with no roles", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			AuxiliaryNodeGroups: []*GcpDataprocClusterAuxiliaryNodeGroup{
				{Roles: []string{}},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject node_group_id shorter than 3 characters", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			AuxiliaryNodeGroups: []*GcpDataprocClusterAuxiliaryNodeGroup{
				{Roles: []string{"DRIVER"}, NodeGroupId: "ab"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("node_group_id"))
	})

	ginkgo.It("should reject node_group_id longer than 33 characters", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			AuxiliaryNodeGroups: []*GcpDataprocClusterAuxiliaryNodeGroup{
				{Roles: []string{"DRIVER"}, NodeGroupId: "abcdefghijklmnopqrstuvwxyz-0123456"},
			},
		}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
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

	ginkgo.It("should reject a virtual arm missing kubernetes_cluster_config", func() {
		msg := minimal()
		msg.Spec.VirtualClusterConfig = &GcpDataprocClusterVirtualClusterConfig{}
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a virtual arm missing gke_cluster_config", func() {
		msg := minimal()
		arm := virtualArm()
		arm.KubernetesClusterConfig.GkeClusterConfig = nil
		msg.Spec.VirtualClusterConfig = arm
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a virtual arm missing kubernetes_software_config", func() {
		msg := minimal()
		arm := virtualArm()
		arm.KubernetesClusterConfig.KubernetesSoftwareConfig = nil
		msg.Spec.VirtualClusterConfig = arm
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a virtual arm missing gke_cluster_target", func() {
		msg := minimal()
		arm := virtualArm()
		arm.KubernetesClusterConfig.GkeClusterConfig.GkeClusterTarget = nil
		msg.Spec.VirtualClusterConfig = arm
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an empty component_version map", func() {
		msg := minimal()
		arm := virtualArm()
		arm.KubernetesClusterConfig.KubernetesSoftwareConfig.ComponentVersion = map[string]string{}
		msg.Spec.VirtualClusterConfig = arm
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a node pool target without node_pool", func() {
		msg := minimal()
		arm := virtualArm()
		arm.KubernetesClusterConfig.GkeClusterConfig.NodePoolTarget = []*GcpDataprocClusterNodePoolTarget{
			{Roles: []string{"DEFAULT"}},
		}
		msg.Spec.VirtualClusterConfig = arm
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a node pool target with no roles", func() {
		msg := minimal()
		arm := virtualArm()
		arm.KubernetesClusterConfig.GkeClusterConfig.NodePoolTarget = []*GcpDataprocClusterNodePoolTarget{
			{NodePool: svr("projects/p/locations/l/clusters/c/nodePools/np")},
		}
		msg.Spec.VirtualClusterConfig = arm
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid node pool role", func() {
		msg := minimal()
		arm := virtualArm()
		arm.KubernetesClusterConfig.GkeClusterConfig.NodePoolTarget = []*GcpDataprocClusterNodePoolTarget{
			{
				NodePool: svr("projects/p/locations/l/clusters/c/nodePools/np"),
				Roles:    []string{"WORKER"},
			},
		}
		msg.Spec.VirtualClusterConfig = arm
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a node pool config with preemptible and spot together", func() {
		msg := minimal()
		arm := virtualArm()
		arm.KubernetesClusterConfig.GkeClusterConfig.NodePoolTarget = []*GcpDataprocClusterNodePoolTarget{
			{
				NodePool: svr("projects/p/locations/l/clusters/c/nodePools/np"),
				Roles:    []string{"DEFAULT"},
				NodePoolConfig: &GcpDataprocClusterNodePoolConfig{
					Locations:   []string{"us-central1-a"},
					Preemptible: true,
					Spot:        true,
				},
			},
		}
		msg.Spec.VirtualClusterConfig = arm
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("preemptible"))
	})

	ginkgo.It("should reject a node pool config with empty locations", func() {
		msg := minimal()
		arm := virtualArm()
		arm.KubernetesClusterConfig.GkeClusterConfig.NodePoolTarget = []*GcpDataprocClusterNodePoolTarget{
			{
				NodePool: svr("projects/p/locations/l/clusters/c/nodePools/np"),
				Roles:    []string{"DEFAULT"},
				NodePoolConfig: &GcpDataprocClusterNodePoolConfig{
					Locations: []string{},
				},
			},
		}
		msg.Spec.VirtualClusterConfig = arm
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject node pool autoscaling with max below min", func() {
		msg := minimal()
		arm := virtualArm()
		arm.KubernetesClusterConfig.GkeClusterConfig.NodePoolTarget = []*GcpDataprocClusterNodePoolTarget{
			{
				NodePool: svr("projects/p/locations/l/clusters/c/nodePools/np"),
				Roles:    []string{"DEFAULT"},
				NodePoolConfig: &GcpDataprocClusterNodePoolConfig{
					Locations: []string{"us-central1-a"},
					Autoscaling: &GcpDataprocClusterNodePoolAutoscaling{
						MinNodeCount: 5,
						MaxNodeCount: 2,
					},
				},
			},
		}
		msg.Spec.VirtualClusterConfig = arm
		err := validator.Validate(msg)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(err.Error()).To(gomega.ContainSubstring("max_node_count"))
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

	// ──────────────── Structural type, engine, teardown ────────────────

	ginkgo.It("should accept every valid cluster_type", func() {
		for _, t := range []string{"", "STANDARD", "SINGLE_NODE", "ZERO_SCALE"} {
			msg := minimal()
			msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{ClusterType: t}
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an unknown cluster_type", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{ClusterType: "TINY"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should accept both engines", func() {
		for _, e := range []string{"", "DEFAULT", "LIGHTNING"} {
			msg := minimal()
			msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{Engine: e}
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an unknown engine", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{Engine: "TURBO"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should accept every valid deletion_policy", func() {
		for _, p := range []string{"", "DELETE", "PREVENT", "ABANDON"} {
			msg := minimal()
			msg.Spec.DeletionPolicy = p
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		msg := minimal()
		msg.Spec.DeletionPolicy = "RETAIN"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	// ──────────────── Stop-lifecycle and disk performance ────────────────

	ginkgo.It("should accept stop-lifecycle settings", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			LifecycleConfig: &GcpDataprocClusterLifecycleConfig{
				IdleStopTtl:  "1800s",
				AutoStopTime: "2026-12-31T00:00:00Z",
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed idle_stop_ttl", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			LifecycleConfig: &GcpDataprocClusterLifecycleConfig{IdleStopTtl: "30m"},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should accept provisioned boot-disk performance", func() {
		iops := int64(3000)
		throughput := int64(290)
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MasterConfig: &GcpDataprocClusterMasterConfig{
				DiskConfig: &GcpDataprocClusterDiskConfig{
					BootDiskProvisionedIops:       &iops,
					BootDiskProvisionedThroughput: &throughput,
				},
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	// ──────────────── Instance flexibility on master and workers ────────────────

	ginkgo.It("should accept flexibility selections on master and primary workers", func() {
		flexibility := &GcpDataprocClusterInstanceFlexibilityPolicy{
			InstanceSelectionList: []*GcpDataprocClusterInstanceSelection{
				{MachineTypes: []string{"n2-standard-8", "n2d-standard-8"}, Rank: 1},
			},
		}
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MasterConfig: &GcpDataprocClusterMasterConfig{InstanceFlexibilityPolicy: flexibility},
			WorkerConfig: &GcpDataprocClusterWorkerConfig{InstanceFlexibilityPolicy: flexibility},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a provisioning mix on master flexibility", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MasterConfig: &GcpDataprocClusterMasterConfig{
				InstanceFlexibilityPolicy: &GcpDataprocClusterInstanceFlexibilityPolicy{
					ProvisioningModelMix: &GcpDataprocClusterProvisioningModelMix{StandardCapacityBase: 2},
				},
			},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a provisioning mix on primary-worker flexibility", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			WorkerConfig: &GcpDataprocClusterWorkerConfig{
				InstanceFlexibilityPolicy: &GcpDataprocClusterInstanceFlexibilityPolicy{
					ProvisioningModelMix: &GcpDataprocClusterProvisioningModelMix{StandardCapacityBase: 2},
				},
			},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should still accept a provisioning mix on secondary workers", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			SecondaryWorkerConfig: &GcpDataprocClusterSecondaryWorkerConfig{
				InstanceFlexibilityPolicy: &GcpDataprocClusterInstanceFlexibilityPolicy{
					ProvisioningModelMix: &GcpDataprocClusterProvisioningModelMix{
						StandardCapacityBase:             2,
						StandardCapacityPercentAboveBase: 50,
					},
				},
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	// The API silently drops a machineTypeUri paired with a flexibility
	// policy; because machine_type is create-only, the pairing re-plans as
	// whole-cluster replacement on every apply (live-verified 2026-08-12).
	ginkgo.It("should reject machine_type paired with flexibility on the master", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			MasterConfig: &GcpDataprocClusterMasterConfig{
				MachineType: "n2-standard-4",
				InstanceFlexibilityPolicy: &GcpDataprocClusterInstanceFlexibilityPolicy{
					InstanceSelectionList: []*GcpDataprocClusterInstanceSelection{
						{MachineTypes: []string{"n2-standard-4"}, Rank: 1},
					},
				},
			},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject machine_type paired with flexibility on primary workers", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			WorkerConfig: &GcpDataprocClusterWorkerConfig{
				MachineType: "e2-standard-2",
				InstanceFlexibilityPolicy: &GcpDataprocClusterInstanceFlexibilityPolicy{
					InstanceSelectionList: []*GcpDataprocClusterInstanceSelection{
						{MachineTypes: []string{"e2-standard-2", "n2-standard-2"}, Rank: 1},
					},
				},
			},
		}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	// ──────────────── Resource manager tags ────────────────

	ginkgo.It("should accept resource manager tags on the GCE config", func() {
		msg := minimal()
		msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
			GceConfig: &GcpDataprocClusterGceConfig{
				ResourceManagerTags: map[string]string{
					"tagKeys/281475039756228": "tagValues/281476599162946",
				},
			},
		}
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.Context("attached disks, per-selection disk shapes, and confidential type", func() {
		ginkgo.It("accepts attached disks on the worker role and a per-selection disk shape", func() {
			msg := minimal()
			iops := int64(3000)
			msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
				WorkerConfig: &GcpDataprocClusterWorkerConfig{
					NumInstances: 2,
					DiskConfig: &GcpDataprocClusterDiskConfig{
						BootDiskType: "pd-balanced",
						AttachedDisks: []*GcpDataprocClusterAttachedDisk{
							{DiskSizeGb: 500, DiskType: "pd-ssd"},
							{DiskSizeGb: 1000, DiskType: "hyperdisk-balanced", ProvisionedIops: &iops},
						},
					},
					InstanceFlexibilityPolicy: &GcpDataprocClusterInstanceFlexibilityPolicy{
						InstanceSelectionList: []*GcpDataprocClusterInstanceSelection{{
							MachineTypes: []string{"n2-standard-8"},
							DiskConfig:   &GcpDataprocClusterDiskConfig{BootDiskSizeGb: 200, NumLocalSsds: 2, LocalSsdInterface: "nvme"},
						}},
					},
				},
			}
			gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
		})

		ginkgo.It("rejects an attached disk below 10 GB or with an unknown type", func() {
			msg := minimal()
			msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
				MasterConfig: &GcpDataprocClusterMasterConfig{
					DiskConfig: &GcpDataprocClusterDiskConfig{AttachedDisks: []*GcpDataprocClusterAttachedDisk{{DiskSizeGb: 5}}},
				},
			}
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
			msg.Spec.ClusterConfig.MasterConfig.DiskConfig.AttachedDisks[0] = &GcpDataprocClusterAttachedDisk{DiskType: "local-ssd"}
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		})

		ginkgo.It("rejects attached disks on an auxiliary node group", func() {
			msg := minimal()
			msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
				AuxiliaryNodeGroups: []*GcpDataprocClusterAuxiliaryNodeGroup{{
					Roles: []string{"DRIVER"},
					NodeGroupConfig: &GcpDataprocClusterAuxiliaryNodeGroupConfig{
						DiskConfig: &GcpDataprocClusterDiskConfig{AttachedDisks: []*GcpDataprocClusterAttachedDisk{{DiskSizeGb: 100}}},
					},
				}},
			}
			err := validator.Validate(msg)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("auxiliary"))
		})

		ginkgo.It("accepts each confidential instance type and rejects an unknown one", func() {
			for _, t := range []string{"SEV", "SEV_SNP", "TDX"} {
				msg := minimal()
				msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
					GceConfig: &GcpDataprocClusterGceConfig{
						ConfidentialInstanceConfig: &GcpDataprocClusterConfidentialInstanceConfig{ConfidentialInstanceType: t},
					},
				}
				gomega.Expect(validator.Validate(msg)).To(gomega.Succeed(), "type %q", t)
			}
			msg := minimal()
			msg.Spec.ClusterConfig = &GcpDataprocClusterConfig{
				GceConfig: &GcpDataprocClusterGceConfig{
					ConfidentialInstanceConfig: &GcpDataprocClusterConfidentialInstanceConfig{ConfidentialInstanceType: "SGX"},
				},
			}
			gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		})
	})
})
