package azureaksnodepoolv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestAzureAksNodePoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureAksNodePoolSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("AzureAksNodePoolSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("azure_aks_node_pool with minimal configuration", func() {

			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &AzureAksNodePool{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-node-pool",
					},
					Spec: &AzureAksNodePoolSpec{
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-aks-cluster"},
						},
						ResourceGroup:     stringRef("test-rg"),
						VmSize:            "Standard_D4s_v3",
						InitialNodeCount:  2,
						AvailabilityZones: []string{"1", "2"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with autoscaling enabled", func() {
				input := &AzureAksNodePool{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "autoscaling-pool",
					},
					Spec: &AzureAksNodePoolSpec{
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-aks-cluster"},
						},
						ResourceGroup:    stringRef("test-rg"),
						VmSize:           "Standard_D8s_v3",
						InitialNodeCount: 2,
						Autoscaling: &AzureAksNodePoolAutoscaling{
							MinNodes: 2,
							MaxNodes: 10,
						},
						AvailabilityZones: []string{"1", "2", "3"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for System mode pool", func() {
				input := &AzureAksNodePool{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "system-pool",
					},
					Spec: &AzureAksNodePoolSpec{
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-aks-cluster"},
						},
						ResourceGroup:     stringRef("test-rg"),
						VmSize:            "Standard_D4s_v3",
						InitialNodeCount:  3,
						AvailabilityZones: []string{"1", "2", "3"},
						Mode:              AzureAksNodePoolMode_SYSTEM.Enum(),
						OsType:            AzureAksNodePoolOsType_LINUX.Enum(),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for User mode pool with Spot enabled", func() {
				input := &AzureAksNodePool{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "spot-pool",
					},
					Spec: &AzureAksNodePoolSpec{
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-aks-cluster"},
						},
						ResourceGroup:    stringRef("test-rg"),
						VmSize:           "Standard_D8s_v3",
						InitialNodeCount: 1, // Must start with at least 1 node, can scale to 0 via autoscaling
						Autoscaling: &AzureAksNodePoolAutoscaling{
							MinNodes: 0, // Allow scale to zero
							MaxNodes: 20,
						},
						AvailabilityZones: []string{"1", "2", "3"},
						Mode:              AzureAksNodePoolMode_USER.Enum(),
						SpotEnabled:       true,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for Windows pool", func() {
				input := &AzureAksNodePool{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "windows-pool",
					},
					Spec: &AzureAksNodePoolSpec{
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-aks-cluster"},
						},
						ResourceGroup:     stringRef("test-rg"),
						VmSize:            "Standard_D4s_v3",
						InitialNodeCount:  2,
						AvailabilityZones: []string{"1", "2"},
						Mode:              AzureAksNodePoolMode_USER.Enum(),
						OsType:            AzureAksNodePoolOsType_WINDOWS.Enum(),
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error without availability zones", func() {
				input := &AzureAksNodePool{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-zones-pool",
					},
					Spec: &AzureAksNodePoolSpec{
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-aks-cluster"},
						},
						ResourceGroup:     stringRef("test-rg"),
						VmSize:            "Standard_D4s_v3",
						InitialNodeCount:  2,
						AvailabilityZones: []string{}, // Empty zones is valid (uses regional defaults)
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("azure_aks_node_pool", func() {

			ginkgo.It("should return a validation error when cluster_name is missing", func() {
				input := &AzureAksNodePool{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-node-pool",
					},
					Spec: &AzureAksNodePoolSpec{
						ResourceGroup:     stringRef("test-rg"),
						VmSize:            "Standard_D4s_v3",
						InitialNodeCount:  2,
						AvailabilityZones: []string{"1", "2"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when vm_size is missing", func() {
				input := &AzureAksNodePool{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-node-pool",
					},
					Spec: &AzureAksNodePoolSpec{
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-aks-cluster"},
						},
						ResourceGroup:     stringRef("test-rg"),
						InitialNodeCount:  2,
						AvailabilityZones: []string{"1", "2"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when initial_node_count is zero", func() {
				input := &AzureAksNodePool{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-node-pool",
					},
					Spec: &AzureAksNodePoolSpec{
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-aks-cluster"},
						},
						ResourceGroup:     stringRef("test-rg"),
						VmSize:            "Standard_D4s_v3",
						InitialNodeCount:  0,
						AvailabilityZones: []string{"1", "2"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when initial_node_count is negative", func() {
				input := &AzureAksNodePool{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-node-pool",
					},
					Spec: &AzureAksNodePoolSpec{
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-aks-cluster"},
						},
						ResourceGroup:     stringRef("test-rg"),
						VmSize:            "Standard_D4s_v3",
						InitialNodeCount:  -1,
						AvailabilityZones: []string{"1", "2"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when only one availability zone is specified", func() {
				input := &AzureAksNodePool{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-node-pool",
					},
					Spec: &AzureAksNodePoolSpec{
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-aks-cluster"},
						},
						ResourceGroup:     stringRef("test-rg"),
						VmSize:            "Standard_D4s_v3",
						InitialNodeCount:  2,
						AvailabilityZones: []string{"1"}, // Only 1 zone - violates min_items=2
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when availability zone is invalid", func() {
				input := &AzureAksNodePool{
					ApiVersion: "azure.planton.dev/v1",
					Kind:       "AzureAksNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-node-pool",
					},
					Spec: &AzureAksNodePoolSpec{
						ClusterName: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-aks-cluster"},
						},
						ResourceGroup:     stringRef("test-rg"),
						VmSize:            "Standard_D4s_v3",
						InitialNodeCount:  2,
						AvailabilityZones: []string{"1", "4"}, // "4" is not valid (only 1, 2, 3 allowed)
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})

func stringRef(s string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: s}}
}
