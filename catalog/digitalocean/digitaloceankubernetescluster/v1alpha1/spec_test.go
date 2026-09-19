package digitaloceankubernetesclusterv1alpha1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/catalog/digitalocean"
	"github.com/plantonhq/planton/shared"
)

func TestDigitalOceanKubernetesClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanKubernetesClusterSpec Custom Validation Tests")
}

func boolPtr(b bool) *bool { return &b }

func float64Ptr(f float64) *float64 { return &f }

// validMinimalSpec returns a fresh spec carrying only the required fields, so
// each test mutates its own copy without cross-test bleed.
func validMinimalSpec() *DigitalOceanKubernetesClusterSpec {
	return &DigitalOceanKubernetesClusterSpec{
		ClusterName:       "test-cluster",
		Region:            digitalocean.DigitalOceanRegion_nyc3,
		KubernetesVersion: "1.35",
		Vpc: &foreignkeyv1.StringValueOrRef{
			LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "b5648f9e-a28a-4760-bb87-b2fad07ae295"},
		},
		DefaultNodePool: &DigitalOceanKubernetesClusterDefaultNodePool{
			Size:      "s-2vcpu-4gb",
			NodeCount: 3,
		},
	}
}

func wrap(spec *DigitalOceanKubernetesClusterSpec) *DigitalOceanKubernetesCluster {
	return &DigitalOceanKubernetesCluster{
		ApiVersion: "digital-ocean.planton.dev/v1alpha1",
		Kind:       "DigitalOceanKubernetesCluster",
		Metadata: &shared.CloudResourceMetadata{
			Name: "test-k8s-cluster",
		},
		Spec: spec,
	}
}

var _ = ginkgo.Describe("DigitalOceanKubernetesClusterSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("digitalocean_kubernetes_cluster", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				err := protovalidate.Validate(wrap(validMinimalSpec()))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept the full surface with every optional field set", func() {
				spec := validMinimalSpec()
				spec.HighlyAvailable = true
				spec.AutoUpgrade = true
				spec.RegistryIntegration = true
				spec.SurgeUpgrade = boolPtr(false)
				spec.Tags = []string{"env:prod", "team_platform"}
				spec.MaintenancePolicy = &DigitalOceanKubernetesClusterMaintenancePolicy{
					Day:       "sunday",
					StartTime: "02:00",
				}
				spec.ControlPlaneFirewall = &DigitalOceanKubernetesClusterControlPlaneFirewall{
					Enabled:          boolPtr(true),
					AllowedAddresses: []string{"203.0.113.5", "198.51.100.0/24"},
				}
				spec.ClusterSubnet = "10.100.0.0/16"
				spec.ServiceSubnet = "10.200.0.0/16"
				spec.WorkerSubnetUuid = "7e3f2d5c-1111-2222-3333-444455556666"
				spec.IsolatedWorkers = true
				spec.DestroyAllAssociatedResources = true
				spec.KubeconfigExpireSeconds = 3600
				spec.ClusterAutoscalerConfiguration = &DigitalOceanKubernetesClusterAutoscalerConfiguration{
					ScaleDownUtilizationThreshold: float64Ptr(0.5),
					ScaleDownUnneededTime:         "1m30s",
					Expanders:                     []string{"least-waste", "random"},
				}
				spec.Sso = &DigitalOceanKubernetesClusterSso{
					Enabled:   boolPtr(true),
					Required:  true,
					IssuerUrl: "https://issuer.example.com",
					ClientId:  "client-abc",
				}
				spec.RoutingAgent = &DigitalOceanKubernetesClusterFeatureToggle{Enabled: boolPtr(true)}
				spec.P2POciRegistryPlugin = &DigitalOceanKubernetesClusterFeatureToggle{Enabled: boolPtr(true)}
				spec.AmdGpuDevicePlugin = &DigitalOceanKubernetesClusterFeatureToggle{Enabled: boolPtr(true)}
				spec.AmdGpuDeviceMetricsExporterPlugin = &DigitalOceanKubernetesClusterFeatureToggle{Enabled: boolPtr(true)}
				spec.NvidiaGpuDevicePlugin = &DigitalOceanKubernetesClusterFeatureToggle{Enabled: boolPtr(true)}
				spec.RdmaSharedDevicePlugin = &DigitalOceanKubernetesClusterFeatureToggle{Enabled: boolPtr(false)}
				spec.CorednsAutoscaler = &DigitalOceanKubernetesClusterFeatureToggle{Enabled: boolPtr(true)}
				spec.DefaultNodePool = &DigitalOceanKubernetesClusterDefaultNodePool{
					Size:      "s-2vcpu-4gb",
					AutoScale: true,
					MinNodes:  1,
					MaxNodes:  5,
					Labels:    map[string]string{"workload": "general"},
					Taints: []*DigitalOceanKubernetesClusterNodePoolTaint{
						{Key: "dedicated", Value: "gpu", Effect: "NoSchedule"},
					},
					Tags:             []string{"pool:default"},
					GpuPartitionMode: "AMD_PARTITION_MODE_SPX_NPS1",
				}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept an addon toggle asserting OFF (enabled: false)", func() {
				spec := validMinimalSpec()
				spec.RoutingAgent = &DigitalOceanKubernetesClusterFeatureToggle{Enabled: boolPtr(false)}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a valueless taint (Kubernetes allows them)", func() {
				spec := validMinimalSpec()
				spec.DefaultNodePool.Taints = []*DigitalOceanKubernetesClusterNodePoolTaint{
					{Key: "dedicated", Effect: "NoSchedule"},
				}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept a disabled control-plane firewall with staged addresses", func() {
				spec := validMinimalSpec()
				spec.ControlPlaneFirewall = &DigitalOceanKubernetesClusterControlPlaneFirewall{
					Enabled:          boolPtr(false),
					AllowedAddresses: []string{"203.0.113.5/32"},
				}
				err := protovalidate.Validate(wrap(spec))
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("required fields", func() {

			ginkgo.It("should return an error when cluster_name is missing", func() {
				spec := validMinimalSpec()
				spec.ClusterName = ""
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when region is missing", func() {
				spec := validMinimalSpec()
				spec.Region = digitalocean.DigitalOceanRegion_digital_ocean_region_unspecified
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when kubernetes_version is missing", func() {
				spec := validMinimalSpec()
				spec.KubernetesVersion = ""
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when vpc is missing", func() {
				spec := validMinimalSpec()
				spec.Vpc = nil
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should return an error when default_node_pool is missing", func() {
				spec := validMinimalSpec()
				spec.DefaultNodePool = nil
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("tag validation", func() {

			ginkgo.It("should reject a cluster tag with characters outside the provider's set", func() {
				spec := validMinimalSpec()
				spec.Tags = []string{"has space"}
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a node pool tag with characters outside the provider's set", func() {
				spec := validMinimalSpec()
				spec.DefaultNodePool.Tags = []string{"bad/tag"}
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("default node pool", func() {

			ginkgo.It("should return an error when node_count is zero on a fixed pool", func() {
				spec := validMinimalSpec()
				spec.DefaultNodePool.NodeCount = 0
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject node_count together with auto_scale", func() {
				spec := validMinimalSpec()
				spec.DefaultNodePool.AutoScale = true
				spec.DefaultNodePool.MinNodes = 1
				spec.DefaultNodePool.MaxNodes = 3
				spec.DefaultNodePool.NodeCount = 2
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should accept auto_scale with bounds and no node_count", func() {
				spec := validMinimalSpec()
				spec.DefaultNodePool.AutoScale = true
				spec.DefaultNodePool.MinNodes = 1
				spec.DefaultNodePool.MaxNodes = 3
				spec.DefaultNodePool.NodeCount = 0
				gomega.Expect(protovalidate.Validate(wrap(spec))).To(gomega.BeNil())
			})

			ginkgo.It("should reject auto_scale without min_nodes", func() {
				spec := validMinimalSpec()
				spec.DefaultNodePool.AutoScale = true
				spec.DefaultNodePool.MaxNodes = 5
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject auto_scale with max_nodes below min_nodes", func() {
				spec := validMinimalSpec()
				spec.DefaultNodePool.AutoScale = true
				spec.DefaultNodePool.MinNodes = 5
				spec.DefaultNodePool.MaxNodes = 2
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an unknown gpu_partition_mode", func() {
				spec := validMinimalSpec()
				spec.DefaultNodePool.GpuPartitionMode = "SPX_NPS1"
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a taint without a key", func() {
				spec := validMinimalSpec()
				spec.DefaultNodePool.Taints = []*DigitalOceanKubernetesClusterNodePoolTaint{
					{Value: "gpu", Effect: "NoSchedule"},
				}
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a taint with an unknown effect", func() {
				spec := validMinimalSpec()
				spec.DefaultNodePool.Taints = []*DigitalOceanKubernetesClusterNodePoolTaint{
					{Key: "dedicated", Effect: "noschedule"},
				}
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("maintenance policy", func() {

			ginkgo.It("should reject an unknown day", func() {
				spec := validMinimalSpec()
				spec.MaintenancePolicy = &DigitalOceanKubernetesClusterMaintenancePolicy{
					Day:       "someday",
					StartTime: "02:00",
				}
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a start_time that is not 24-hour HH:MM", func() {
				spec := validMinimalSpec()
				spec.MaintenancePolicy = &DigitalOceanKubernetesClusterMaintenancePolicy{
					Day:       "sunday",
					StartTime: "25:00",
				}
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("control plane firewall", func() {

			ginkgo.It("should reject a firewall without the enabled flag", func() {
				spec := validMinimalSpec()
				spec.ControlPlaneFirewall = &DigitalOceanKubernetesClusterControlPlaneFirewall{
					AllowedAddresses: []string{"203.0.113.5/32"},
				}
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject an allowed address that is neither IP nor CIDR", func() {
				spec := validMinimalSpec()
				spec.ControlPlaneFirewall = &DigitalOceanKubernetesClusterControlPlaneFirewall{
					Enabled:          boolPtr(true),
					AllowedAddresses: []string{"office-network"},
				}
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("subnets", func() {

			ginkgo.It("should reject a cluster_subnet that is not CIDR notation", func() {
				spec := validMinimalSpec()
				spec.ClusterSubnet = "10.100.0.0"
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject a service_subnet that is not CIDR notation", func() {
				spec := validMinimalSpec()
				spec.ServiceSubnet = "not-a-cidr"
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("sso", func() {

			ginkgo.It("should reject sso without the enabled flag", func() {
				spec := validMinimalSpec()
				spec.Sso = &DigitalOceanKubernetesClusterSso{
					IssuerUrl: "https://issuer.example.com",
				}
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("cluster autoscaler configuration", func() {

			ginkgo.It("should reject a utilization threshold above 1", func() {
				spec := validMinimalSpec()
				spec.ClusterAutoscalerConfiguration = &DigitalOceanKubernetesClusterAutoscalerConfiguration{
					ScaleDownUtilizationThreshold: float64Ptr(1.5),
				}
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})
		})

		ginkgo.Context("addon mutual exclusions", func() {

			ginkgo.It("should reject AMD device plugin combined with AMD DRA driver", func() {
				spec := validMinimalSpec()
				spec.AmdGpuDevicePlugin = &DigitalOceanKubernetesClusterFeatureToggle{Enabled: boolPtr(true)}
				spec.AmdGpuDraDriver = &DigitalOceanKubernetesClusterFeatureToggle{Enabled: boolPtr(true)}
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})

			ginkgo.It("should reject NVIDIA device plugin combined with NVIDIA DRA driver", func() {
				spec := validMinimalSpec()
				spec.NvidiaGpuDevicePlugin = &DigitalOceanKubernetesClusterFeatureToggle{Enabled: boolPtr(true)}
				spec.NvidiaGpuDraDriver = &DigitalOceanKubernetesClusterFeatureToggle{Enabled: boolPtr(true)}
				gomega.Expect(protovalidate.Validate(wrap(spec))).NotTo(gomega.BeNil())
			})
		})
	})
})
