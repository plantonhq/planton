package civokubernetesclusterv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	civoprovider "github.com/plantonhq/planton/apis/dev/planton/provider/civo"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestCivoKubernetesClusterSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CivoKubernetesClusterSpec Custom Validation Tests")
}

var _ = ginkgo.Describe("CivoKubernetesClusterSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("minimal valid cluster", func() {
			ginkgo.It("should not return a validation error for minimal valid fields", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "test-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("cluster with optional features", func() {
			ginkgo.It("should accept cluster with high availability enabled", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "ha-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "ha-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-456",
							},
						},
						HighlyAvailable: true,
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.large",
							NodeCount: 5,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept cluster with auto-upgrade enabled", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "auto-upgrade-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "auto-upgrade-k8s",
						Region:            civoprovider.CivoRegion_nyc1,
						KubernetesVersion: "1.28.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-789",
							},
						},
						AutoUpgrade: true,
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept cluster with surge upgrade disabled", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-surge-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "no-surge-k8s",
						Region:            civoprovider.CivoRegion_fra1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-abc",
							},
						},
						DisableSurgeUpgrade: true,
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept cluster with tags", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "tagged-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "tagged-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-def",
							},
						},
						Tags: []string{"production", "team-platform", "cost-center-eng"},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("different regions", func() {
			ginkgo.It("should accept LON1 region", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "london-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "london-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-lon",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.small",
							NodeCount: 1,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept NYC1 region", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "newyork-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "newyork-k8s",
						Region:            civoprovider.CivoRegion_nyc1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-nyc",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept FRA1 region", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "frankfurt-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "frankfurt-k8s",
						Region:            civoprovider.CivoRegion_fra1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-fra",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.large",
							NodeCount: 5,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("node pool variations", func() {
			ginkgo.It("should accept single node cluster", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "single-node-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "single-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-single",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.small",
							NodeCount: 1,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept large production cluster", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "large-prod-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "prod-k8s",
						Region:            civoprovider.CivoRegion_fra1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-prod",
							},
						},
						HighlyAvailable: true,
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.xlarge",
							NodeCount: 10,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept various node sizes", func() {
				nodeSizes := []string{
					"g4s.kube.small",
					"g4s.kube.medium",
					"g4s.kube.large",
					"g4s.kube.xlarge",
				}

				for _, size := range nodeSizes {
					input := &CivoKubernetesCluster{
						ApiVersion: "civo.planton.dev/v1",
						Kind:       "CivoKubernetesCluster",
						Metadata: &shared.CloudResourceMetadata{
							Name: "test-" + size,
						},
						Spec: &CivoKubernetesClusterSpec{
							ClusterName:       "test-" + size,
							Region:            civoprovider.CivoRegion_lon1,
							KubernetesVersion: "1.29.0+k3s1",
							Network: &foreignkeyv1.StringValueOrRef{
								LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
									Value: "network-test",
								},
							},
							DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
								Size:      size,
								NodeCount: 3,
							},
						},
					}
					err := protovalidate.Validate(input)
					gomega.Expect(err).To(gomega.BeNil())
				}
			})
		})

		ginkgo.Context("real-world configurations", func() {
			ginkgo.It("should accept dev cluster configuration", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "dev-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "dev-k8s",
						Region:            civoprovider.CivoRegion_nyc1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "vpc-dev",
							},
						},
						AutoUpgrade: true,
						Tags:        []string{"environment:dev", "managed-by:planton"},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.small",
							NodeCount: 1,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept staging cluster configuration", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "staging-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "staging-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "vpc-staging",
							},
						},
						AutoUpgrade: false,
						Tags:        []string{"environment:staging", "team:platform"},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should accept production cluster configuration", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "production-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "prod-k8s",
						Region:            civoprovider.CivoRegion_fra1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "vpc-production",
							},
						},
						HighlyAvailable:     true,
						AutoUpgrade:         false,
						DisableSurgeUpgrade: false,
						Tags:                []string{"environment:production", "critical:true", "backup:enabled"},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.large",
							NodeCount: 5,
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
			ginkgo.It("should return a validation error for missing cluster_name", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-name-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "", // Missing required field
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing kubernetes_version", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-version-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "test-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "", // Missing required field
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing network", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-network-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "test-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network:           nil, // Missing required field
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing default_node_pool", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-nodepool-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "test-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
						DefaultNodePool: nil, // Missing required field
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing metadata", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata:   nil, // Missing required metadata
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "test-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing spec", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-spec-cluster",
					},
					Spec: nil, // Missing required spec
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid node pool configuration", func() {
			ginkgo.It("should return a validation error for zero node count", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "zero-nodes-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "zero-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 0, // Invalid: must be > 0
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for missing node size", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-size-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "test-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "", // Missing required field
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid API version and kind", func() {
			ginkgo.It("should return a validation error for wrong api_version", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "wrong.api.version/v1", // Wrong value
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "wrong-api-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "test-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error for wrong kind", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "WrongKind", // Wrong kind
					Metadata: &shared.CloudResourceMetadata{
						Name: "wrong-kind-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "test-k8s",
						Region:            civoprovider.CivoRegion_lon1,
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("invalid region", func() {
			ginkgo.It("should return a validation error for unspecified region", func() {
				input := &CivoKubernetesCluster{
					ApiVersion: "civo.planton.dev/v1",
					Kind:       "CivoKubernetesCluster",
					Metadata: &shared.CloudResourceMetadata{
						Name: "unspecified-region-cluster",
					},
					Spec: &CivoKubernetesClusterSpec{
						ClusterName:       "test-k8s",
						Region:            civoprovider.CivoRegion_civo_region_unspecified, // Invalid
						KubernetesVersion: "1.29.0+k3s1",
						Network: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "network-123",
							},
						},
						DefaultNodePool: &CivoKubernetesClusterDefaultNodePool{
							Size:      "g4s.kube.medium",
							NodeCount: 3,
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
