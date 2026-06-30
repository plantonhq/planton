package civokubernetesnodepoolv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestCivoKubernetesNodePoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CivoKubernetesNodePoolSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CivoKubernetesNodePoolSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("minimal valid node pool", func() {
			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-nodepool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "test-cluster",
							},
						},
						Size:      "g4s.kube.medium",
						NodeCount: 3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("node pool with various node counts", func() {
			ginkgo.It("should accept single node pool", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "single-node-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "single-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "dev-cluster",
							},
						},
						Size:      "g4s.kube.small",
						NodeCount: 1,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept large node pool", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "large-node-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "large-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-cluster",
							},
						},
						Size:      "g4s.kube.xlarge",
						NodeCount: 10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("node pool with different sizes", func() {
			ginkgo.It("should accept small node size", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "small-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "small-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "cluster-123",
							},
						},
						Size:      "g4s.kube.small",
						NodeCount: 2,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept medium node size", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "medium-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "medium-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "cluster-456",
							},
						},
						Size:      "g4s.kube.medium",
						NodeCount: 3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept large node size", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "large-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "large-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "cluster-789",
							},
						},
						Size:      "g4s.kube.large",
						NodeCount: 5,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("node pool with autoscaling", func() {
			ginkgo.It("should accept node pool with autoscaling enabled", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "autoscaling-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "autoscaling-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-cluster",
							},
						},
						Size:      "g4s.kube.medium",
						NodeCount: 3,
						AutoScale: true,
						MinNodes:  2,
						MaxNodes:  10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept different autoscaling ranges", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "wide-range-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "wide-range-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "scaling-cluster",
							},
						},
						Size:      "g4s.kube.large",
						NodeCount: 5,
						AutoScale: true,
						MinNodes:  1,
						MaxNodes:  20,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept autoscaling with node_count between min and max", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "balanced-autoscale-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "balanced-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "cluster-xyz",
							},
						},
						Size:      "g4s.kube.medium",
						NodeCount: 5,
						AutoScale: true,
						MinNodes:  3,
						MaxNodes:  8,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("node pool with tags", func() {
			ginkgo.It("should accept node pool with tags", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "tagged-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "tagged-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "cluster-abc",
							},
						},
						Size:      "g4s.kube.medium",
						NodeCount: 3,
						Tags:      []string{"environment:production", "team:platform", "managed-by:planton"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("real-world configurations", func() {
			ginkgo.It("should accept general-purpose worker pool", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "general-workers",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "general-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "prod-k8s",
							},
						},
						Size:      "g4s.kube.medium",
						NodeCount: 5,
						AutoScale: true,
						MinNodes:  3,
						MaxNodes:  10,
						Tags:      []string{"workload:general", "tier:workers"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept compute-intensive pool", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "compute-intensive-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "compute-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "data-cluster",
							},
						},
						Size:      "g4s.kube.xlarge",
						NodeCount: 3,
						Tags:      []string{"workload:compute-intensive", "team:data-science"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept batch processing pool", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "batch-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "batch-workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "batch-cluster",
							},
						},
						Size:      "g4s.kube.large",
						NodeCount: 2,
						AutoScale: true,
						MinNodes:  0,
						MaxNodes:  20,
						Tags:      []string{"workload:batch", "autoscale:aggressive"},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("missing required fields", func() {
			ginkgo.It("should return a validation error for missing node_pool_name", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-name-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "", // Missing required field
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "cluster-123",
							},
						},
						Size:      "g4s.kube.medium",
						NodeCount: 3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing cluster", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-cluster-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "workers",
						Cluster:      nil, // Missing required field
						Size:         "g4s.kube.medium",
						NodeCount:    3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing size", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-size-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "cluster-123",
							},
						},
						Size:      "", // Missing required field
						NodeCount: 3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing metadata", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata:   nil, // Missing required metadata
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "cluster-123",
							},
						},
						Size:      "g4s.kube.medium",
						NodeCount: 3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing spec", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-spec-pool",
					},
					Spec: nil, // Missing required spec
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid node count", func() {
			ginkgo.It("should return a validation error for zero node count", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "zero-nodes-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "cluster-123",
							},
						},
						Size:      "g4s.kube.medium",
						NodeCount: 0, // Invalid: must be > 0
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid API version and kind", func() {
			ginkgo.It("should return a validation error for wrong api_version", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "wrong.api.version/v1", // Wrong value
					Kind:       "CivoKubernetesNodePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "wrong-api-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "cluster-123",
							},
						},
						Size:      "g4s.kube.medium",
						NodeCount: 3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for wrong kind", func() {
				input := &CivoKubernetesNodePool{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "WrongKind", // Wrong kind
					Metadata: &shared.CloudResourceMetadata{
						Name: "wrong-kind-pool",
					},
					Spec: &CivoKubernetesNodePoolSpec{
						NodePoolName: "workers",
						Cluster: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "cluster-123",
							},
						},
						Size:      "g4s.kube.medium",
						NodeCount: 3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
