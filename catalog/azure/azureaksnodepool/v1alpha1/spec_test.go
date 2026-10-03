package azureaksnodepoolv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestAzureAksNodePoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureAksNodePoolSpec Validation Tests")
}

// literal builds a StringValueOrRef carrying a literal value.
func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

// ref builds a StringValueOrRef carrying a value_from reference.
func ref(name string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_ValueFrom{
			ValueFrom: &foreignkeyv1.ValueFromRef{Name: name},
		},
	}
}

const (
	testClusterId = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.ContainerService/managedClusters/my-aks"
	testSubnetId  = "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet/subnets/pool"
)

// validResource returns a minimal valid AzureAksNodePool that individual
// cases then mutate into the shape under test.
func validResource() *AzureAksNodePool {
	return &AzureAksNodePool{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureAksNodePool",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-pool",
		},
		Spec: &AzureAksNodePoolSpec{
			KubernetesClusterId: literal(testClusterId),
			Name:                "general",
			VmSize:              "Standard_D4s_v5",
			NodeCount:           2,
		},
	}
}

var _ = ginkgo.Describe("AzureAksNodePoolSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should not return a validation error for minimal valid fields", func() {
			err := protovalidate.Validate(validResource())
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept the cluster as a reference", func() {
			input := validResource()
			input.Spec.KubernetesClusterId = ref("prod-cluster")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept an autoscaled pool that scales to zero", func() {
			input := validResource()
			input.Spec.NodeCount = 0
			input.Spec.AutoScalingEnabled = true
			input.Spec.MinCount = 0
			input.Spec.MaxCount = 10
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a parked pool with zero nodes and no autoscaling", func() {
			input := validResource()
			input.Spec.NodeCount = 0
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a spot pool with eviction policy and max price", func() {
			input := validResource()
			input.Spec.Priority = AzureAksNodePoolPriority_SPOT
			input.Spec.EvictionPolicy = AzureAksNodePoolEvictionPolicy_EVICTION_DEALLOCATE
			input.Spec.SpotMaxPrice = 0.27113
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a spot pool paying up to the on-demand price", func() {
			input := validResource()
			input.Spec.Priority = AzureAksNodePoolPriority_SPOT
			input.Spec.SpotMaxPrice = -1.0
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a system pool on Linux on-demand nodes", func() {
			input := validResource()
			input.Spec.Mode = AzureAksNodePoolMode_SYSTEM
			input.Spec.NodeCount = 3
			input.Spec.Zones = []string{"1", "2", "3"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a windows pool with a short name and windows profile", func() {
			input := validResource()
			input.Spec.Name = "win1"
			input.Spec.OsType = AzureAksNodePoolOsType_WINDOWS
			input.Spec.OsSku = AzureAksNodePoolOsSku_WINDOWS_2022
			input.Spec.WindowsProfile = &AzureAksNodePoolWindowsProfile{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept taints, labels, and zones", func() {
			input := validResource()
			input.Spec.NodeTaints = []string{"sku=gpu:NoSchedule", "dedicated=batch:PreferNoSchedule"}
			input.Spec.NodeLabels = map[string]string{"workload": "gpu"}
			input.Spec.Zones = []string{"1", "2"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept a dedicated subnet and pod subnet", func() {
			input := validResource()
			input.Spec.VnetSubnetId = literal(testSubnetId)
			input.Spec.PodSubnetId = ref("pods-subnet")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept upgrade settings with max_unavailable", func() {
			input := validResource()
			input.Spec.UpgradeSettings = &AzureAksNodePoolUpgradeSettings{
				MaxUnavailable:            "1",
				DrainTimeoutInMinutes:     15,
				NodeSoakDurationInMinutes: 5,
				UndrainableNodeBehavior:   AzureAksNodePoolUndrainableNodeBehavior_CORDON,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept kubelet config and linux os config with sysctls", func() {
			input := validResource()
			input.Spec.KubeletConfig = &AzureAksNodePoolKubeletConfig{
				CpuManagerPolicy:     AzureAksNodePoolCpuManagerPolicy_CPU_MANAGER_STATIC,
				ImageGcHighThreshold: 85,
				ContainerLogMaxFiles: 5,
			}
			input.Spec.LinuxOsConfig = &AzureAksNodePoolLinuxOsConfig{
				SysctlConfig: &AzureAksNodePoolSysctlConfig{
					VmMaxMapCount:    262144,
					NetCoreSomaxconn: 65535,
				},
				TransparentHugePage: AzureAksNodePoolTransparentHugePage_THP_NEVER,
				SwapFileSizeMb:      1024,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept gpu instance profile with driver disabled", func() {
			input := validResource()
			input.Spec.GpuInstance = AzureAksNodePoolGpuInstance_MIG1G
			input.Spec.GpuDriver = AzureAksNodePoolGpuDriver_NONE
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept node network profile with host ports and ASGs", func() {
			input := validResource()
			input.Spec.NodeNetworkProfile = &AzureAksNodePoolNodeNetworkProfile{
				AllowedHostPorts: []*AzureAksNodePoolAllowedHostPorts{
					{PortStart: 7000, PortEnd: 8000, Protocol: AzureAksNodePoolHostPortProtocol_UDP},
				},
				ApplicationSecurityGroupIds: []string{
					"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg/providers/Microsoft.Network/applicationSecurityGroups/game-nodes",
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept node public ips from a prefix", func() {
			input := validResource()
			input.Spec.NodePublicIpEnabled = true
			input.Spec.NodePublicIpPrefixId = ref("node-prefix")
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept placement, disk, and rotation options", func() {
			input := validResource()
			input.Spec.OsDiskType = AzureAksNodePoolOsDiskType_EPHEMERAL
			input.Spec.OsDiskSizeGb = 128
			input.Spec.KubeletDiskType = AzureAksNodePoolKubeletDiskType_TEMPORARY
			input.Spec.ScaleDownMode = AzureAksNodePoolScaleDownMode_DEALLOCATE
			input.Spec.TemporaryNameForRotation = "generaltmp"
			input.Spec.Tags = map[string]string{"team": "platform"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should return an error when kubernetes_cluster_id is missing", func() {
			input := validResource()
			input.Spec.KubernetesClusterId = nil
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error when name is missing", func() {
			input := validResource()
			input.Spec.Name = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an uppercase pool name", func() {
			input := validResource()
			input.Spec.Name = "General"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a name longer than 12 characters", func() {
			input := validResource()
			input.Spec.Name = "waytoolongname"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error when vm_size is missing", func() {
			input := validResource()
			input.Spec.VmSize = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a node count above 1000", func() {
			input := validResource()
			input.Spec.NodeCount = 1001
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for inverted autoscaling bounds", func() {
			input := validResource()
			input.Spec.AutoScalingEnabled = true
			input.Spec.MinCount = 10
			input.Spec.MaxCount = 5
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for min/max without autoscaling", func() {
			input := validResource()
			input.Spec.MinCount = 1
			input.Spec.MaxCount = 5
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a spot system pool", func() {
			input := validResource()
			input.Spec.Mode = AzureAksNodePoolMode_SYSTEM
			input.Spec.Priority = AzureAksNodePoolPriority_SPOT
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a windows system pool", func() {
			input := validResource()
			input.Spec.Name = "win1"
			input.Spec.Mode = AzureAksNodePoolMode_SYSTEM
			input.Spec.OsType = AzureAksNodePoolOsType_WINDOWS
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an eviction policy on an on-demand pool", func() {
			input := validResource()
			input.Spec.EvictionPolicy = AzureAksNodePoolEvictionPolicy_EVICTION_DELETE
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a spot price on an on-demand pool", func() {
			input := validResource()
			input.Spec.SpotMaxPrice = 0.5
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a negative spot price other than -1", func() {
			input := validResource()
			input.Spec.Priority = AzureAksNodePoolPriority_SPOT
			input.Spec.SpotMaxPrice = -0.5
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a malformed taint", func() {
			input := validResource()
			input.Spec.NodeTaints = []string{"sku=gpu"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an invalid zone", func() {
			input := validResource()
			input.Spec.Zones = []string{"4"}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a windows os_sku on a linux pool", func() {
			input := validResource()
			input.Spec.OsSku = AzureAksNodePoolOsSku_WINDOWS_2022
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a linux os_sku on a windows pool", func() {
			input := validResource()
			input.Spec.Name = "win1"
			input.Spec.OsType = AzureAksNodePoolOsType_WINDOWS
			input.Spec.OsSku = AzureAksNodePoolOsSku_UBUNTU
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a long windows pool name", func() {
			input := validResource()
			input.Spec.Name = "windows"
			input.Spec.OsType = AzureAksNodePoolOsType_WINDOWS
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a windows profile on a linux pool", func() {
			input := validResource()
			input.Spec.WindowsProfile = &AzureAksNodePoolWindowsProfile{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for linux os config on a windows pool", func() {
			input := validResource()
			input.Spec.Name = "win1"
			input.Spec.OsType = AzureAksNodePoolOsType_WINDOWS
			input.Spec.LinuxOsConfig = &AzureAksNodePoolLinuxOsConfig{}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for both surge and unavailable in upgrade settings", func() {
			input := validResource()
			input.Spec.UpgradeSettings = &AzureAksNodePoolUpgradeSettings{
				MaxSurge:       "10%",
				MaxUnavailable: "1",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for surge settings on a spot pool", func() {
			input := validResource()
			input.Spec.Priority = AzureAksNodePoolPriority_SPOT
			input.Spec.UpgradeSettings = &AzureAksNodePoolUpgradeSettings{
				MaxSurge: "10%",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept drain-timeout-only upgrade settings on a spot pool", func() {
			input := validResource()
			input.Spec.Priority = AzureAksNodePoolPriority_SPOT
			input.Spec.UpgradeSettings = &AzureAksNodePoolUpgradeSettings{
				DrainTimeoutInMinutes: 15,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should return an error for an out-of-range soak duration", func() {
			input := validResource()
			input.Spec.UpgradeSettings = &AzureAksNodePoolUpgradeSettings{
				MaxSurge:                  "10%",
				NodeSoakDurationInMinutes: 45,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for a node public ip prefix without node public ips", func() {
			input := validResource()
			input.Spec.NodePublicIpPrefixId = ref("node-prefix")
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an out-of-range sysctl value", func() {
			input := validResource()
			input.Spec.LinuxOsConfig = &AzureAksNodePoolLinuxOsConfig{
				SysctlConfig: &AzureAksNodePoolSysctlConfig{
					NetCoreSomaxconn: 100,
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an out-of-range host port", func() {
			input := validResource()
			input.Spec.NodeNetworkProfile = &AzureAksNodePoolNodeNetworkProfile{
				AllowedHostPorts: []*AzureAksNodePoolAllowedHostPorts{
					{PortStart: 70000, PortEnd: 70001},
				},
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should return an error for an out-of-range image gc threshold", func() {
			input := validResource()
			input.Spec.KubeletConfig = &AzureAksNodePoolKubeletConfig{
				ImageGcHighThreshold: 150,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})
})
