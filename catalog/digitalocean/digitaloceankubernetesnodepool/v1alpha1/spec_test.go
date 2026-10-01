package digitaloceankubernetesnodepoolv1alpha1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/shared"
)

func TestDigitalOceanKubernetesNodePoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "DigitalOceanKubernetesNodePoolSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("DigitalOceanKubernetesNodePoolSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("digitalocean_kubernetes_node_pool", func() {

			ginkgo.It("should not return a validation error for minimal valid fields (fixed node count)", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "app-workers",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "app-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						NodeCount: 3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with autoscaling enabled", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "autoscale-workers",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "autoscale-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-4vcpu-8gb",
						AutoScale: true,
						MinNodes:  3,
						MaxNodes:  10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with labels", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "labeled-workers",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "labeled-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						NodeCount: 3,
						Labels: map[string]string{
							"workload": "web",
							"env":      "production",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with taints", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "tainted-workers",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "tainted-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "g-4vcpu-16gb",
						NodeCount: 2,
						Taints: []*DigitalOceanKubernetesNodePoolTaint{
							{
								Key:    "nvidia.com/gpu",
								Value:  "true",
								Effect: "NoSchedule",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with all optional fields", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-featured-workers",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "full-featured-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-4vcpu-8gb",
						AutoScale: true,
						MinNodes:  3,
						MaxNodes:  10,
						Labels: map[string]string{
							"workload": "application",
							"tier":     "backend",
						},
						Taints: []*DigitalOceanKubernetesNodePoolTaint{
							{
								Key:    "dedicated",
								Value:  "backend",
								Effect: "NoSchedule",
							},
						},
						Tags: []string{
							"env:production",
							"team:platform",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a valueless taint", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "valueless-taint-workers",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "valueless-taint-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						NodeCount: 1,
						Taints: []*DigitalOceanKubernetesNodePoolTaint{
							{
								Key:    "dedicated",
								Effect: "PreferNoSchedule",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for a valid gpu_partition_mode", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "gpu-workers",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "gpu-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:             "gpu-mi300x1-192gb",
						NodeCount:        1,
						GpuPartitionMode: "AMD_PARTITION_MODE_SPX_NPS1",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error when autoscale bounds are equal", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "pinned-autoscale-workers",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "pinned-autoscale-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						AutoScale: true,
						MinNodes:  2,
						MaxNodes:  2,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing required fields", func() {

			ginkgo.It("should return a validation error when node_pool_name is missing", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						NodeCount: 3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cluster is missing", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "test-pool",
						Size:         "s-2vcpu-4gb",
						NodeCount:    3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when size is missing", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "test-pool",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						NodeCount: 3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when node_count is set together with auto_scale", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "test-pool",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						NodeCount: 2,
						AutoScale: true,
						MinNodes:  1,
						MaxNodes:  3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when node_count is zero on a fixed pool", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "test-pool",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						NodeCount: 0,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when taint is missing required key", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "test-pool",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						NodeCount: 3,
						Taints: []*DigitalOceanKubernetesNodePoolTaint{
							{
								Value:  "true",
								Effect: "NoSchedule",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when taint is missing required effect", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "test-pool",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						NodeCount: 3,
						Taints: []*DigitalOceanKubernetesNodePoolTaint{
							{
								Key:   "nvidia.com/gpu",
								Value: "true",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for an invalid taint effect", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "test-pool",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						NodeCount: 3,
						Taints: []*DigitalOceanKubernetesNodePoolTaint{
							{
								Key:    "dedicated",
								Effect: "noschedule",
							},
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("autoscale bounds", func() {

			ginkgo.It("should return a validation error when auto_scale is on without min_nodes", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "test-pool",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						AutoScale: true,
						MaxNodes:  5,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when max_nodes is below min_nodes", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "test-pool",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						AutoScale: true,
						MinNodes:  5,
						MaxNodes:  2,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("value-restricted fields", func() {

			ginkgo.It("should return a validation error for an invalid gpu_partition_mode", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "test-pool",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:             "gpu-mi300x1-192gb",
						NodeCount:        1,
						GpuPartitionMode: "SPX",
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for a tag with illegal characters", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1alpha1",
					Kind:       "DigitalOceanKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &DigitalOceanKubernetesNodePoolSpec{
						NodePoolName: "test-pool",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "test-cluster-id"},
						},
						Size:      "s-2vcpu-4gb",
						NodeCount: 3,
						Tags:      []string{"has space"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
