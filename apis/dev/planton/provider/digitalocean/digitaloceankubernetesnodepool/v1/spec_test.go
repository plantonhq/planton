package digitaloceankubernetesnodepoolv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
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
					ApiVersion: "digital-ocean.planton.dev/v1",
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
					ApiVersion: "digital-ocean.planton.dev/v1",
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
						NodeCount: 3,
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
					ApiVersion: "digital-ocean.planton.dev/v1",
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
					ApiVersion: "digital-ocean.planton.dev/v1",
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
					ApiVersion: "digital-ocean.planton.dev/v1",
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
						NodeCount: 5,
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
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing required fields", func() {

			ginkgo.It("should return a validation error when node_pool_name is missing", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1",
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
					ApiVersion: "digital-ocean.planton.dev/v1",
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
					ApiVersion: "digital-ocean.planton.dev/v1",
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

			ginkgo.It("should return a validation error when node_count is zero", func() {
				input := &DigitalOceanKubernetesNodePool{
					ApiVersion: "digital-ocean.planton.dev/v1",
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
					ApiVersion: "digital-ocean.planton.dev/v1",
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
					ApiVersion: "digital-ocean.planton.dev/v1",
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
		})
	})
})
