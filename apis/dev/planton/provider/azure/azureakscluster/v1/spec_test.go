package azureaksclusterv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}

func TestAzureAksClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureAksClusterSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AzureAksClusterSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_aks_cluster with minimal production configuration", func() {

			ginkgo.It("should not return a validation error for minimal valid production fields", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						VnetSubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						KubernetesVersion: "1.30",
						ControlPlaneSku:   AzureAksClusterControlPlaneSku_STANDARD,
						NetworkPlugin:     AzureAksClusterNetworkPlugin_AZURE_CNI,
						NetworkPluginMode: AzureAksClusterNetworkPluginMode_OVERLAY,
						SystemNodePool: &AzureAksClusterSystemNodePool{
							VmSize: "Standard_D4s_v5",
							Autoscaling: &AzureAksClusterAutoscalingConfig{
								MinCount: 3,
								MaxCount: 5,
							},
							AvailabilityZones: []string{"1", "2", "3"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with user node pools", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						VnetSubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						SystemNodePool: &AzureAksClusterSystemNodePool{
							VmSize: "Standard_D4s_v5",
							Autoscaling: &AzureAksClusterAutoscalingConfig{
								MinCount: 3,
								MaxCount: 5,
							},
							AvailabilityZones: []string{"1", "2", "3"},
						},
						UserNodePools: []*AzureAksClusterUserNodePool{
							{
								Name:   "general",
								VmSize: "Standard_D8s_v5",
								Autoscaling: &AzureAksClusterAutoscalingConfig{
									MinCount: 2,
									MaxCount: 10,
								},
								AvailabilityZones: []string{"1", "2", "3"},
								SpotEnabled:       false,
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with add-ons configuration", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						VnetSubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						SystemNodePool: &AzureAksClusterSystemNodePool{
							VmSize: "Standard_D4s_v5",
							Autoscaling: &AzureAksClusterAutoscalingConfig{
								MinCount: 3,
								MaxCount: 5,
							},
							AvailabilityZones: []string{"1", "2", "3"},
						},
						Addons: &AzureAksClusterAddonsConfig{
							EnableContainerInsights: true,
							EnableKeyVaultCsiDriver: true,
							EnableAzurePolicy:       true,
							EnableWorkloadIdentity:  true,
							LogAnalyticsWorkspaceId: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.OperationalInsights/workspaces/logs",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with private cluster and authorized IP ranges", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						VnetSubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						SystemNodePool: &AzureAksClusterSystemNodePool{
							VmSize: "Standard_D4s_v5",
							Autoscaling: &AzureAksClusterAutoscalingConfig{
								MinCount: 3,
								MaxCount: 5,
							},
							AvailabilityZones: []string{"1", "2", "3"},
						},
						PrivateClusterEnabled: false,
						AuthorizedIpRanges:    []string{"203.0.113.0/24", "198.51.100.0/24"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with advanced networking", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						VnetSubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						SystemNodePool: &AzureAksClusterSystemNodePool{
							VmSize: "Standard_D4s_v5",
							Autoscaling: &AzureAksClusterAutoscalingConfig{
								MinCount: 3,
								MaxCount: 5,
							},
							AvailabilityZones: []string{"1", "2", "3"},
						},
						AdvancedNetworking: &AzureAksClusterAdvancedNetworking{
							PodCidr:          "10.244.0.0/16",
							ServiceCidr:      "10.0.0.0/16",
							DnsServiceIp:     "10.0.0.10",
							CustomDnsServers: []string{"8.8.8.8", "8.8.4.4"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_aks_cluster", func() {

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						ResourceGroup: stringRef("test-rg"),
						VnetSubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						SystemNodePool: &AzureAksClusterSystemNodePool{
							VmSize: "Standard_D4s_v5",
							Autoscaling: &AzureAksClusterAutoscalingConfig{
								MinCount: 3,
								MaxCount: 5,
							},
							AvailabilityZones: []string{"1", "2", "3"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vnet_subnet_id is missing", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						SystemNodePool: &AzureAksClusterSystemNodePool{
							VmSize: "Standard_D4s_v5",
							Autoscaling: &AzureAksClusterAutoscalingConfig{
								MinCount: 3,
								MaxCount: 5,
							},
							AvailabilityZones: []string{"1", "2", "3"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when system_node_pool is missing", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						VnetSubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when system node pool vm_size is empty", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						VnetSubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						SystemNodePool: &AzureAksClusterSystemNodePool{
							VmSize: "",
							Autoscaling: &AzureAksClusterAutoscalingConfig{
								MinCount: 3,
								MaxCount: 5,
							},
							AvailabilityZones: []string{"1", "2", "3"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when system node pool autoscaling is missing", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						VnetSubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						SystemNodePool: &AzureAksClusterSystemNodePool{
							VmSize:            "Standard_D4s_v5",
							AvailabilityZones: []string{"1", "2", "3"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when autoscaling min_count is 0", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						VnetSubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						SystemNodePool: &AzureAksClusterSystemNodePool{
							VmSize: "Standard_D4s_v5",
							Autoscaling: &AzureAksClusterAutoscalingConfig{
								MinCount: 0,
								MaxCount: 5,
							},
							AvailabilityZones: []string{"1", "2", "3"},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when user node pool name is invalid", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						VnetSubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						SystemNodePool: &AzureAksClusterSystemNodePool{
							VmSize: "Standard_D4s_v5",
							Autoscaling: &AzureAksClusterAutoscalingConfig{
								MinCount: 3,
								MaxCount: 5,
							},
							AvailabilityZones: []string{"1", "2", "3"},
						},
						UserNodePools: []*AzureAksClusterUserNodePool{
							{
								Name:   "Invalid-Name-With-Caps",
								VmSize: "Standard_D8s_v5",
								Autoscaling: &AzureAksClusterAutoscalingConfig{
									MinCount: 2,
									MaxCount: 10,
								},
								AvailabilityZones: []string{"1", "2", "3"},
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when authorized_ip_ranges has invalid CIDR", func() {
				input := &AzureAksCluster{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-aks-cluster",
					},
					Spec: &AzureAksClusterSpec{
						Region:        "eastus",
						ResourceGroup: stringRef("test-rg"),
						VnetSubnetId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Network/virtualNetworks/vnet/subnets/subnet",
							},
						},
						SystemNodePool: &AzureAksClusterSystemNodePool{
							VmSize: "Standard_D4s_v5",
							Autoscaling: &AzureAksClusterAutoscalingConfig{
								MinCount: 3,
								MaxCount: 5,
							},
							AvailabilityZones: []string{"1", "2", "3"},
						},
						AuthorizedIpRanges: []string{"invalid-cidr"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
