package gcpgkenodepoolv1alpha1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/shared"
)

func TestGcpGkeNodePoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpGkeNodePoolSpec Custom Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

// minimalSpec returns a valid baseline spec every case mutates from.
func minimalSpec() *GcpGkeNodePoolSpec {
	return &GcpGkeNodePoolSpec{
		ClusterName: literal("test-gke-cluster"),
		Location:    literal("us-central1"),
	}
}

func newNodePool(spec *GcpGkeNodePoolSpec) *GcpGkeNodePool {
	return &GcpGkeNodePool{
		ApiVersion: "gcp.planton.dev/v1alpha1",
		Kind:       "GcpGkeNodePool",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-node-pool",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("GcpGkeNodePoolSpec Custom Validation Tests", func() {

	ginkgo.Describe("valid configurations", func() {

		ginkgo.It("accepts a minimal pool (cluster + location only)", func() {
			gomega.Expect(protovalidate.Validate(newNodePool(minimalSpec()))).To(gomega.BeNil())
		})

		ginkgo.It("accepts cluster and location by reference", func() {
			spec := minimalSpec()
			spec.ClusterName = ref("my-cluster")
			spec.Location = ref("my-cluster")
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts a zonal and a multi-digit-region location literal", func() {
			spec := minimalSpec()
			spec.Location = literal("us-central1-a")
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
			spec.Location = literal("europe-west12")
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts an explicit node_pool_name", func() {
			spec := minimalSpec()
			spec.NodePoolName = "general-pool"
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts a fixed node_count", func() {
			spec := minimalSpec()
			spec.NodePoolSize = &GcpGkeNodePoolSpec_NodeCount{NodeCount: 3}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts per-zone autoscaling with scale-to-zero", func() {
			spec := minimalSpec()
			spec.NodePoolSize = &GcpGkeNodePoolSpec_Autoscaling{
				Autoscaling: &GcpGkeNodePoolAutoscaling{
					MinNodes: proto.Uint32(0),
					MaxNodes: proto.Uint32(10),
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts total autoscaling limits", func() {
			spec := minimalSpec()
			spec.NodePoolSize = &GcpGkeNodePoolSpec_Autoscaling{
				Autoscaling: &GcpGkeNodePoolAutoscaling{
					TotalMinNodes:  proto.Uint32(1),
					TotalMaxNodes:  proto.Uint32(50),
					LocationPolicy: "ANY",
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts initial_node_count alongside autoscaling", func() {
			spec := minimalSpec()
			spec.InitialNodeCount = proto.Uint32(1)
			spec.NodePoolSize = &GcpGkeNodePoolSpec_Autoscaling{
				Autoscaling: &GcpGkeNodePoolAutoscaling{MaxNodes: proto.Uint32(5)},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts node_locations within a region", func() {
			spec := minimalSpec()
			spec.NodeLocations = []string{"us-central1-a", "us-central1-b"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts multi-digit-region node_locations", func() {
			spec := minimalSpec()
			spec.NodeLocations = []string{"europe-west12-a"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts management with auto-upgrade off", func() {
			spec := minimalSpec()
			spec.Management = &GcpGkeNodePoolManagement{
				AutoUpgrade: proto.Bool(false),
			}
			spec.Version = "1.31.5-gke.1000"
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts surge upgrade settings", func() {
			spec := minimalSpec()
			spec.UpgradeSettings = &GcpGkeNodePoolUpgradeSettings{
				MaxSurge:       proto.Uint32(2),
				MaxUnavailable: proto.Uint32(1),
				Strategy:       "SURGE",
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts a blue-green upgrade with rollout policy", func() {
			spec := minimalSpec()
			spec.UpgradeSettings = &GcpGkeNodePoolUpgradeSettings{
				Strategy: "BLUE_GREEN",
				BlueGreenSettings: &GcpGkeNodePoolBlueGreenSettings{
					StandardRolloutPolicy: &GcpGkeNodePoolStandardRolloutPolicy{
						BatchPercentage:   proto.Float32(0.25),
						BatchSoakDuration: "600s",
					},
					NodePoolSoakDuration: "3600s",
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts a compact placement policy", func() {
			spec := minimalSpec()
			spec.PlacementPolicy = &GcpGkeNodePoolPlacementPolicy{Type: "COMPACT"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts queued provisioning", func() {
			spec := minimalSpec()
			spec.QueuedProvisioningEnabled = true
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts a dedicated pod range", func() {
			spec := minimalSpec()
			spec.NetworkConfig = &GcpGkeNodePoolNetworkConfig{
				CreatePodRange:   true,
				PodRange:         "pool-pods",
				PodIpv4CidrBlock: "10.96.0.0/14",
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts an existing pod range by name only", func() {
			spec := minimalSpec()
			spec.NetworkConfig = &GcpGkeNodePoolNetworkConfig{PodRange: "shared-pods"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts a full node_config with spot and taints", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				MachineType: proto.String("n2-standard-8"),
				DiskSizeGb:  proto.Uint32(200),
				DiskType:    "pd-ssd",
				Spot:        true,
				Labels:      map[string]string{"workload-class": "batch"},
				Taints: []*GcpGkeNodePoolTaint{
					{Key: "workload-class", Value: "batch", Effect: "NO_SCHEDULE"},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts a service account by reference", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				ServiceAccount: ref("gke-nodes-sa"),
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts a GPU accelerator with driver install and sharing", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				MachineType: proto.String("g2-standard-8"),
				GuestAccelerators: []*GcpGkeNodePoolGuestAccelerator{
					{
						Type:             "nvidia-l4",
						Count:            1,
						GpuDriverVersion: "DEFAULT",
						GpuSharingConfig: &GcpGkeNodePoolGpuSharingConfig{
							GpuSharingStrategy:     "GPU_TIME_SHARING",
							MaxSharedClientsPerGpu: 4,
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts shielded, confidential, and CMEK boot disk settings", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				ShieldedInstanceConfig: &GcpGkeNodePoolShieldedInstanceConfig{
					EnableSecureBoot: true,
				},
				ConfidentialNodes: &GcpGkeNodePoolConfidentialNodes{
					Enabled:                  true,
					ConfidentialInstanceType: "SEV",
				},
				BootDiskKmsKey: ref("nodes-cmek"),
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts local SSD, image streaming, and gVNIC + fast socket", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				EphemeralStorageLocalSsd: &GcpGkeNodePoolEphemeralStorageLocalSsd{
					LocalSsdCount: 2,
				},
				GcfsEnabled:       true,
				GvnicEnabled:      true,
				FastSocketEnabled: true,
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts a specific reservation", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				ReservationAffinity: &GcpGkeNodePoolReservationAffinity{
					ConsumeReservationType: "SPECIFIC_RESERVATION",
					Key:                    "compute.googleapis.com/reservation-name",
					Values:                 []string{"my-reservation"},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts kubelet and linux node tuning", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				KubeletConfig: &GcpGkeNodePoolKubeletConfig{
					CpuManagerPolicy:                   "static",
					PodPidsLimit:                       proto.Int64(4096),
					InsecureKubeletReadonlyPortEnabled: "FALSE",
					ContainerLogMaxSize:                "50Mi",
					ContainerLogMaxFiles:               proto.Int64(5),
					ImageGcLowThresholdPercent:         proto.Int64(60),
					ImageGcHighThresholdPercent:        proto.Int64(80),
				},
				LinuxNodeConfig: &GcpGkeNodePoolLinuxNodeConfig{
					Sysctls:    map[string]string{"net.core.somaxconn": "4096"},
					CgroupMode: "CGROUP_MODE_V2",
					HugepagesConfig: &GcpGkeNodePoolHugepagesConfig{
						HugepageSize_2M: proto.Int64(1024),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts secondary boot disks and max pods per node", func() {
			spec := minimalSpec()
			spec.MaxPodsPerNode = proto.Int32(64)
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				SecondaryBootDisks: []*GcpGkeNodePoolSecondaryBootDisk{
					{DiskImage: "projects/p/global/images/preloaded", Mode: "CONTAINER_IMAGE_CACHE"},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts flex-start with a max run duration", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				FlexStart:      true,
				MaxRunDuration: "86400s",
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("required fields", func() {

		ginkgo.It("rejects a missing cluster_name", func() {
			spec := minimalSpec()
			spec.ClusterName = nil
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a missing location", func() {
			spec := minimalSpec()
			spec.Location = nil
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("naming and location patterns", func() {

		ginkgo.It("rejects an invalid node_pool_name", func() {
			spec := minimalSpec()
			spec.NodePoolName = "Bad_Name"
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a node_pool_name over 40 characters", func() {
			spec := minimalSpec()
			spec.NodePoolName = "a-very-long-node-pool-name-that-exceeds-forty"
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a region in node_locations (zones only)", func() {
			spec := minimalSpec()
			spec.NodeLocations = []string{"us-central1"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects duplicate node_locations", func() {
			spec := minimalSpec()
			spec.NodeLocations = []string{"us-central1-a", "us-central1-a"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("sizing coherence", func() {

		ginkgo.It("rejects initial_node_count on a fixed-size pool", func() {
			spec := minimalSpec()
			spec.InitialNodeCount = proto.Uint32(1)
			spec.NodePoolSize = &GcpGkeNodePoolSpec_NodeCount{NodeCount: 3}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects mixing per-zone and total autoscaling limits", func() {
			spec := minimalSpec()
			spec.NodePoolSize = &GcpGkeNodePoolSpec_Autoscaling{
				Autoscaling: &GcpGkeNodePoolAutoscaling{
					MaxNodes:      proto.Uint32(5),
					TotalMaxNodes: proto.Uint32(20),
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects autoscaling without any maximum", func() {
			spec := minimalSpec()
			spec.NodePoolSize = &GcpGkeNodePoolSpec_Autoscaling{
				Autoscaling: &GcpGkeNodePoolAutoscaling{MinNodes: proto.Uint32(1)},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects min_nodes above max_nodes", func() {
			spec := minimalSpec()
			spec.NodePoolSize = &GcpGkeNodePoolSpec_Autoscaling{
				Autoscaling: &GcpGkeNodePoolAutoscaling{
					MinNodes: proto.Uint32(5),
					MaxNodes: proto.Uint32(2),
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects total_min_nodes above total_max_nodes", func() {
			spec := minimalSpec()
			spec.NodePoolSize = &GcpGkeNodePoolSpec_Autoscaling{
				Autoscaling: &GcpGkeNodePoolAutoscaling{
					TotalMinNodes: proto.Uint32(50),
					TotalMaxNodes: proto.Uint32(10),
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid location_policy", func() {
			spec := minimalSpec()
			spec.NodePoolSize = &GcpGkeNodePoolSpec_Autoscaling{
				Autoscaling: &GcpGkeNodePoolAutoscaling{
					MaxNodes:       proto.Uint32(5),
					LocationPolicy: "SPREAD",
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects max_pods_per_node out of range", func() {
			spec := minimalSpec()
			spec.MaxPodsPerNode = proto.Int32(4)
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
			spec.MaxPodsPerNode = proto.Int32(300)
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("upgrade settings coherence", func() {

		ginkgo.It("rejects blue_green_settings under the SURGE strategy", func() {
			spec := minimalSpec()
			spec.UpgradeSettings = &GcpGkeNodePoolUpgradeSettings{
				Strategy: "SURGE",
				BlueGreenSettings: &GcpGkeNodePoolBlueGreenSettings{
					StandardRolloutPolicy: &GcpGkeNodePoolStandardRolloutPolicy{
						BatchNodeCount: proto.Uint32(1),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects surge dials under the BLUE_GREEN strategy", func() {
			spec := minimalSpec()
			spec.UpgradeSettings = &GcpGkeNodePoolUpgradeSettings{
				Strategy: "BLUE_GREEN",
				MaxSurge: proto.Uint32(2),
				BlueGreenSettings: &GcpGkeNodePoolBlueGreenSettings{
					StandardRolloutPolicy: &GcpGkeNodePoolStandardRolloutPolicy{
						BatchNodeCount: proto.Uint32(1),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid strategy", func() {
			spec := minimalSpec()
			spec.UpgradeSettings = &GcpGkeNodePoolUpgradeSettings{Strategy: "ROLLING"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects blue-green without a rollout policy", func() {
			spec := minimalSpec()
			spec.UpgradeSettings = &GcpGkeNodePoolUpgradeSettings{
				Strategy:          "BLUE_GREEN",
				BlueGreenSettings: &GcpGkeNodePoolBlueGreenSettings{},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a rollout batch sized by both percentage and count", func() {
			spec := minimalSpec()
			spec.UpgradeSettings = &GcpGkeNodePoolUpgradeSettings{
				Strategy: "BLUE_GREEN",
				BlueGreenSettings: &GcpGkeNodePoolBlueGreenSettings{
					StandardRolloutPolicy: &GcpGkeNodePoolStandardRolloutPolicy{
						BatchPercentage: proto.Float32(0.5),
						BatchNodeCount:  proto.Uint32(2),
					},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a malformed soak duration", func() {
			spec := minimalSpec()
			spec.UpgradeSettings = &GcpGkeNodePoolUpgradeSettings{
				Strategy: "BLUE_GREEN",
				BlueGreenSettings: &GcpGkeNodePoolBlueGreenSettings{
					StandardRolloutPolicy: &GcpGkeNodePoolStandardRolloutPolicy{
						BatchNodeCount: proto.Uint32(1),
					},
					NodePoolSoakDuration: "1h",
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("placement and network coherence", func() {

		ginkgo.It("rejects an invalid placement type", func() {
			spec := minimalSpec()
			spec.PlacementPolicy = &GcpGkeNodePoolPlacementPolicy{Type: "SPREAD"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects create_pod_range without a pod_range name", func() {
			spec := minimalSpec()
			spec.NetworkConfig = &GcpGkeNodePoolNetworkConfig{CreatePodRange: true}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid egress bandwidth tier", func() {
			spec := minimalSpec()
			spec.NetworkConfig = &GcpGkeNodePoolNetworkConfig{TotalEgressBandwidthTier: "TIER_2"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("node_config coherence", func() {

		ginkgo.It("rejects spot and preemptible together", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{Spot: true, Preemptible: true}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects fast socket without gVNIC", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{FastSocketEnabled: true}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a boot disk under 10 GB", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{DiskSizeGb: proto.Uint32(5)}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown disk type", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{DiskType: "pd-extreme-fast"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an unknown image type", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{ImageType: proto.String("DEBIAN")}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a taint with an invalid effect", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				Taints: []*GcpGkeNodePoolTaint{
					{Key: "k", Value: "v", Effect: "TAINT"},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a taint missing its key", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				Taints: []*GcpGkeNodePoolTaint{
					{Value: "v", Effect: "NO_SCHEDULE"},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an accelerator without a count", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				GuestAccelerators: []*GcpGkeNodePoolGuestAccelerator{
					{Type: "nvidia-l4"},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid GPU driver version", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				GuestAccelerators: []*GcpGkeNodePoolGuestAccelerator{
					{Type: "nvidia-l4", Count: 1, GpuDriverVersion: "NEWEST"},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects GPU sharing without a strategy", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				GuestAccelerators: []*GcpGkeNodePoolGuestAccelerator{
					{
						Type:  "nvidia-l4",
						Count: 1,
						GpuSharingConfig: &GcpGkeNodePoolGpuSharingConfig{
							MaxSharedClientsPerGpu: 4,
						},
					},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a specific reservation without key and values", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				ReservationAffinity: &GcpGkeNodePoolReservationAffinity{
					ConsumeReservationType: "SPECIFIC_RESERVATION",
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid confidential instance type", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				ConfidentialNodes: &GcpGkeNodePoolConfidentialNodes{
					Enabled:                  true,
					ConfidentialInstanceType: "SGX",
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a malformed max_run_duration", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{MaxRunDuration: "24h"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid kubelet cpu_manager_policy", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				KubeletConfig: &GcpGkeNodePoolKubeletConfig{CpuManagerPolicy: "exclusive"},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a pod PID limit below the floor", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				KubeletConfig: &GcpGkeNodePoolKubeletConfig{PodPidsLimit: proto.Int64(100)},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid cgroup mode", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				LinuxNodeConfig: &GcpGkeNodePoolLinuxNodeConfig{CgroupMode: "V2"},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid logging variant", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{LoggingVariant: "VERBOSE"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid secondary boot disk mode", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{
				SecondaryBootDisks: []*GcpGkeNodePoolSecondaryBootDisk{
					{DiskImage: "img", Mode: "DATA_CACHE"},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid workload metadata mode", func() {
			spec := minimalSpec()
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{WorkloadMetadataMode: "METADATA_SERVER"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("pool lifecycle additions", func() {

		ginkgo.It("rejects node_pool_name combined with name_prefix", func() {
			spec := minimalSpec()
			spec.NodePoolName = "general-pool"
			spec.NamePrefix = "general-"
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("accepts a name_prefix alone", func() {
			spec := minimalSpec()
			spec.NamePrefix = "general-"
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts every deletion_policy value and rejects others", func() {
			spec := minimalSpec()
			for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
				spec.DeletionPolicy = v
				gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
			}
			spec.DeletionPolicy = "KEEP"
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a malformed node drain grace duration", func() {
			spec := minimalSpec()
			spec.NodeDrainConfig = &GcpGkeNodePoolNodeDrainConfig{GraceTerminationDuration: "5m"}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("networking additions", func() {

		ginkgo.It("rejects an additional node network without a subnetwork", func() {
			spec := minimalSpec()
			spec.NetworkConfig = &GcpGkeNodePoolNetworkConfig{
				AdditionalNodeNetworks: []*GcpGkeNodePoolAdditionalNodeNetwork{
					{Network: literal("projects/p/global/networks/n")},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an additional pod network without a secondary range", func() {
			spec := minimalSpec()
			spec.NetworkConfig = &GcpGkeNodePoolNetworkConfig{
				AdditionalPodNetworks: []*GcpGkeNodePoolAdditionalPodNetwork{
					{Subnetwork: literal("projects/p/regions/r/subnetworks/s")},
				},
			}
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("node_config additions", func() {

		nodeCfg := func(mutate func(*GcpGkeNodePoolNodeConfig)) *GcpGkeNodePoolSpec {
			spec := minimalSpec()
			nc := &GcpGkeNodePoolNodeConfig{}
			mutate(nc)
			spec.NodeConfig = nc
			return spec
		}

		ginkgo.It("accepts ANY_RESERVATION_THEN_FAIL reservation consumption", func() {
			spec := nodeCfg(func(nc *GcpGkeNodePoolNodeConfig) {
				nc.ReservationAffinity = &GcpGkeNodePoolReservationAffinity{
					ConsumeReservationType: "ANY_RESERVATION_THEN_FAIL",
				}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid local SSD encryption mode", func() {
			spec := nodeCfg(func(nc *GcpGkeNodePoolNodeConfig) { nc.LocalSsdEncryptionMode = "CMEK" })
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid sandbox type", func() {
			spec := nodeCfg(func(nc *GcpGkeNodePoolNodeConfig) { nc.SandboxType = "KATA" })
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid windows OS version", func() {
			spec := nodeCfg(func(nc *GcpGkeNodePoolNodeConfig) { nc.WindowsOsVersion = "OS_VERSION_LTSC2016" })
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid host maintenance interval", func() {
			spec := nodeCfg(func(nc *GcpGkeNodePoolNodeConfig) { nc.HostMaintenanceInterval = "NEVER" })
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid architecture taint behavior", func() {
			spec := nodeCfg(func(nc *GcpGkeNodePoolNodeConfig) { nc.ArchitectureTaintBehavior = "X86" })
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects three threads per core", func() {
			spec := nodeCfg(func(nc *GcpGkeNodePoolNodeConfig) {
				nc.AdvancedMachineFeatures = &GcpGkeNodePoolAdvancedMachineFeatures{ThreadsPerCore: 3}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a boot disk under 10 GB", func() {
			five := int64(5)
			spec := nodeCfg(func(nc *GcpGkeNodePoolNodeConfig) {
				nc.BootDisk = &GcpGkeNodePoolBootDisk{SizeGb: &five}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a sole-tenant affinity with an invalid operator", func() {
			spec := nodeCfg(func(nc *GcpGkeNodePoolNodeConfig) {
				nc.SoleTenantConfig = &GcpGkeNodePoolSoleTenantConfig{
					NodeAffinities: []*GcpGkeNodePoolSoleTenantAffinity{
						{Key: "compute.googleapis.com/node-group-name", Operator: "EQUALS", Values: []string{"g"}},
					},
				}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a registry CA domain without a secret URI", func() {
			spec := nodeCfg(func(nc *GcpGkeNodePoolNodeConfig) {
				nc.ContainerdConfig = &GcpGkeNodePoolContainerdConfig{
					PrivateRegistryAccess: &GcpGkeNodePoolPrivateRegistryAccess{
						Enabled: true,
						CertificateAuthorityDomains: []*GcpGkeNodePoolRegistryCaDomain{
							{Fqdns: []string{"registry.internal:5000"}},
						},
					},
				}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("kubelet and linux tuning additions", func() {

		kubelet := func(mutate func(*GcpGkeNodePoolKubeletConfig)) *GcpGkeNodePoolSpec {
			spec := minimalSpec()
			kc := &GcpGkeNodePoolKubeletConfig{}
			mutate(kc)
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{KubeletConfig: kc}
			return spec
		}
		linux := func(mutate func(*GcpGkeNodePoolLinuxNodeConfig)) *GcpGkeNodePoolSpec {
			spec := minimalSpec()
			lc := &GcpGkeNodePoolLinuxNodeConfig{}
			mutate(lc)
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{LinuxNodeConfig: lc}
			return spec
		}

		ginkgo.It("accepts a full eviction tuning set", func() {
			spec := kubelet(func(kc *GcpGkeNodePoolKubeletConfig) {
				kc.EvictionSoft = &GcpGkeNodePoolEvictionSignals{MemoryAvailable: "200Mi", PidAvailable: "10%"}
				kc.EvictionSoftGracePeriod = &GcpGkeNodePoolEvictionGracePeriods{MemoryAvailable: "90s"}
				kc.EvictionMinimumReclaim = &GcpGkeNodePoolEvictionMinimumReclaim{MemoryAvailable: "10%"}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("rejects an absolute quantity in eviction minimum reclaim", func() {
			// GKE accepts only percentages for minimum reclaim — the live API
			// rejects quantities with 'invalid percentage "100Mi"'.
			spec := kubelet(func(kc *GcpGkeNodePoolKubeletConfig) {
				kc.EvictionMinimumReclaim = &GcpGkeNodePoolEvictionMinimumReclaim{MemoryAvailable: "100Mi"}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
			spec = kubelet(func(kc *GcpGkeNodePoolKubeletConfig) {
				kc.EvictionMinimumReclaim = &GcpGkeNodePoolEvictionMinimumReclaim{NodefsAvailable: "500Mi"}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a malformed image GC age", func() {
			spec := kubelet(func(kc *GcpGkeNodePoolKubeletConfig) { kc.ImageMinimumGcAge = "2m" })
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a malformed crash-loop backoff cap", func() {
			spec := kubelet(func(kc *GcpGkeNodePoolKubeletConfig) {
				kc.CrashLoopBackOff = &GcpGkeNodePoolCrashLoopBackOff{MaxContainerRestartPeriod: "5m"}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects a lowercase memory manager policy", func() {
			spec := kubelet(func(kc *GcpGkeNodePoolKubeletConfig) {
				kc.MemoryManager = &GcpGkeNodePoolMemoryManager{Policy: "static"}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid topology manager policy and scope", func() {
			spec := kubelet(func(kc *GcpGkeNodePoolKubeletConfig) {
				kc.TopologyManager = &GcpGkeNodePoolTopologyManager{Policy: "numa-strict"}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
			spec = kubelet(func(kc *GcpGkeNodePoolKubeletConfig) {
				kc.TopologyManager = &GcpGkeNodePoolTopologyManager{Scope: "node"}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid transparent hugepage mode", func() {
			spec := linux(func(lc *GcpGkeNodePoolLinuxNodeConfig) { lc.TransparentHugepageEnabled = "SOMETIMES" })
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects an invalid kernel module loading policy", func() {
			spec := linux(func(lc *GcpGkeNodePoolLinuxNodeConfig) { lc.NodeKernelModuleLoadingPolicy = "ALLOW_ALL" })
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects swap sized by two profiles at once", func() {
			gib := int64(16)
			spec := linux(func(lc *GcpGkeNodePoolLinuxNodeConfig) {
				lc.SwapConfig = &GcpGkeNodePoolSwapConfig{
					BootDiskProfile:          &GcpGkeNodePoolSwapSizing{SwapSizeGib: &gib},
					DedicatedLocalSsdProfile: &GcpGkeNodePoolSwapDedicatedSsd{},
				}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})

		ginkgo.It("rejects swap sizing by both GiB and percent", func() {
			gib := int64(16)
			pct := int32(50)
			spec := linux(func(lc *GcpGkeNodePoolLinuxNodeConfig) {
				lc.SwapConfig = &GcpGkeNodePoolSwapConfig{
					BootDiskProfile: &GcpGkeNodePoolSwapSizing{SwapSizeGib: &gib, SwapSizePercent: &pct},
				}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).ToNot(gomega.BeNil())
		})
	})

	ginkgo.Describe("graceful shutdown, init scripts, and the end-of-support hold", func() {

		kubelet := func(mutate func(*GcpGkeNodePoolKubeletConfig)) *GcpGkeNodePoolSpec {
			spec := minimalSpec()
			kc := &GcpGkeNodePoolKubeletConfig{}
			mutate(kc)
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{KubeletConfig: kc}
			return spec
		}
		linux := func(mutate func(*GcpGkeNodePoolLinuxNodeConfig)) *GcpGkeNodePoolSpec {
			spec := minimalSpec()
			lc := &GcpGkeNodePoolLinuxNodeConfig{}
			mutate(lc)
			spec.NodeConfig = &GcpGkeNodePoolNodeConfig{LinuxNodeConfig: lc}
			return spec
		}

		ginkgo.It("accepts a shutdown grace period with a smaller critical-pod share", func() {
			total, critical := int64(120), int64(30)
			spec := kubelet(func(kc *GcpGkeNodePoolKubeletConfig) {
				kc.ShutdownGracePeriodSeconds = &total
				kc.ShutdownGracePeriodCriticalPodsSeconds = &critical
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("rejects a critical-pod share above the total and a total out of range", func() {
			total, critical := int64(60), int64(90)
			spec := kubelet(func(kc *GcpGkeNodePoolKubeletConfig) {
				kc.ShutdownGracePeriodSeconds = &total
				kc.ShutdownGracePeriodCriticalPodsSeconds = &critical
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).NotTo(gomega.BeNil())
			small := int64(5)
			spec = kubelet(func(kc *GcpGkeNodePoolKubeletConfig) { kc.ShutdownGracePeriodSeconds = &small })
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("accepts an init script from Cloud Storage pinned to a generation", func() {
			gen := int64(1700000000000000)
			spec := linux(func(lc *GcpGkeNodePoolLinuxNodeConfig) {
				lc.CustomNodeInit = &GcpGkeNodePoolCustomNodeInit{GcsUri: "gs://my-bucket/node-init.sh", GcsGeneration: &gen}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("accepts an init script from Secret Manager", func() {
			spec := linux(func(lc *GcpGkeNodePoolLinuxNodeConfig) {
				lc.CustomNodeInit = &GcpGkeNodePoolCustomNodeInit{SecretManagerSecretUri: "projects/p/secrets/node-init/versions/latest"}
			})
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})

		ginkgo.It("rejects an init script with both sources, no source, or a generation without a GCS URI", func() {
			gen := int64(7)
			bad := []*GcpGkeNodePoolCustomNodeInit{
				{GcsUri: "gs://b/x.sh", SecretManagerSecretUri: "projects/p/secrets/s/versions/1"},
				{},
				{SecretManagerSecretUri: "projects/p/secrets/s/versions/1", GcsGeneration: &gen},
				{GcsUri: "s3://b/x.sh"},
			}
			for i, cfg := range bad {
				spec := linux(func(lc *GcpGkeNodePoolLinuxNodeConfig) { lc.CustomNodeInit = cfg })
				gomega.Expect(protovalidate.Validate(newNodePool(spec))).NotTo(gomega.BeNil(), "case %d", i)
			}
		})

		ginkgo.It("accepts the end-of-support upgrade hold", func() {
			spec := minimalSpec()
			spec.ExcludeUpgradesUntilEndOfSupport = true
			gomega.Expect(protovalidate.Validate(newNodePool(spec))).To(gomega.BeNil())
		})
	})
})
