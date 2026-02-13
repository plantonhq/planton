package scalewaykapsulepoolv1

import (
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/openmcf/apis/org/openmcf/shared/foreignkey/v1"

	"buf.build/go/protovalidate"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared"
)

func TestScalewayKapsulePoolSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "ScalewayKapsulePoolSpec Validation Tests")
}

var _ = ginkgo.Describe("ScalewayKapsulePoolSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("scaleway_kapsule_pool", func() {

			ginkgo.It("should not return a validation error for minimal valid fields (fixed size)", func() {
				input := &ScalewayKapsulePool{
					ApiVersion: "scaleway.openmcf.org/v1",
					Kind:       "ScalewayKapsulePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "app-workers",
					},
					Spec: &ScalewayKapsulePoolSpec{
						Region: "fr-par",
						ClusterId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "fr-par/test-cluster-id"},
						},
						NodeType: "GP1-XS",
						Size:     3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with autoscaling enabled", func() {
				input := &ScalewayKapsulePool{
					ApiVersion: "scaleway.openmcf.org/v1",
					Kind:       "ScalewayKapsulePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "autoscale-workers",
					},
					Spec: &ScalewayKapsulePoolSpec{
						Region: "fr-par",
						ClusterId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "fr-par/test-cluster-id"},
						},
						NodeType:  "PRO2-S",
						Size:      3,
						AutoScale: true,
						MinSize:   2,
						MaxSize:   10,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with kubernetes labels", func() {
				input := &ScalewayKapsulePool{
					ApiVersion: "scaleway.openmcf.org/v1",
					Kind:       "ScalewayKapsulePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "labeled-workers",
					},
					Spec: &ScalewayKapsulePoolSpec{
						Region: "fr-par",
						ClusterId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "fr-par/test-cluster-id"},
						},
						NodeType: "GP1-XS",
						Size:     3,
						KubernetesLabels: map[string]string{
							"workload": "web",
							"env":      "production",
						},
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error with taints", func() {
				input := &ScalewayKapsulePool{
					ApiVersion: "scaleway.openmcf.org/v1",
					Kind:       "ScalewayKapsulePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "gpu-workers",
					},
					Spec: &ScalewayKapsulePoolSpec{
						Region: "fr-par",
						ClusterId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "fr-par/test-cluster-id"},
						},
						NodeType: "GPU-3070-S",
						Size:     2,
						Taints: []*ScalewayKapsulePoolTaint{
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
				input := &ScalewayKapsulePool{
					ApiVersion: "scaleway.openmcf.org/v1",
					Kind:       "ScalewayKapsulePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "full-featured-pool",
						Org:  "mycompany",
						Env:  "production",
					},
					Spec: &ScalewayKapsulePoolSpec{
						Region: "fr-par",
						ClusterId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "fr-par/test-cluster-id"},
						},
						NodeType:           "PRO2-M",
						Size:               5,
						AutoScale:          true,
						MinSize:            3,
						MaxSize:            10,
						Autohealing:        true,
						ContainerRuntime:   "containerd",
						RootVolumeType:     "l_ssd",
						RootVolumeSizeInGb: 100,
						PublicIpDisabled:    true,
						Zone:               "fr-par-1",
						PlacementGroupId:   "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
						KubernetesLabels: map[string]string{
							"workload": "application",
							"tier":     "backend",
						},
						Taints: []*ScalewayKapsulePoolTaint{
							{
								Key:    "dedicated",
								Value:  "backend",
								Effect: "NoSchedule",
							},
						},
						UpgradePolicy: &ScalewayKapsulePoolUpgradePolicy{
							MaxSurge:       1,
							MaxUnavailable: 0,
						},
						KubeletArgs: map[string]string{
							"maxPods": "150",
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

			ginkgo.It("should return a validation error when region is missing", func() {
				input := &ScalewayKapsulePool{
					ApiVersion: "scaleway.openmcf.org/v1",
					Kind:       "ScalewayKapsulePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &ScalewayKapsulePoolSpec{
						ClusterId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "fr-par/test-cluster-id"},
						},
						NodeType: "GP1-XS",
						Size:     3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when cluster_id is missing", func() {
				input := &ScalewayKapsulePool{
					ApiVersion: "scaleway.openmcf.org/v1",
					Kind:       "ScalewayKapsulePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &ScalewayKapsulePoolSpec{
						Region:   "fr-par",
						NodeType: "GP1-XS",
						Size:     3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when node_type is missing", func() {
				input := &ScalewayKapsulePool{
					ApiVersion: "scaleway.openmcf.org/v1",
					Kind:       "ScalewayKapsulePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &ScalewayKapsulePoolSpec{
						Region: "fr-par",
						ClusterId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "fr-par/test-cluster-id"},
						},
						Size: 3,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when size is zero", func() {
				input := &ScalewayKapsulePool{
					ApiVersion: "scaleway.openmcf.org/v1",
					Kind:       "ScalewayKapsulePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &ScalewayKapsulePoolSpec{
						Region: "fr-par",
						ClusterId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "fr-par/test-cluster-id"},
						},
						NodeType: "GP1-XS",
						Size:     0,
					},
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return a validation error when taint is missing required key", func() {
				input := &ScalewayKapsulePool{
					ApiVersion: "scaleway.openmcf.org/v1",
					Kind:       "ScalewayKapsulePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &ScalewayKapsulePoolSpec{
						Region: "fr-par",
						ClusterId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "fr-par/test-cluster-id"},
						},
						NodeType: "GP1-XS",
						Size:     3,
						Taints: []*ScalewayKapsulePoolTaint{
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
				input := &ScalewayKapsulePool{
					ApiVersion: "scaleway.openmcf.org/v1",
					Kind:       "ScalewayKapsulePool",
					Metadata: &shared.CloudResourceMetadata{
						Name: "test-pool",
					},
					Spec: &ScalewayKapsulePoolSpec{
						Region: "fr-par",
						ClusterId: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: "fr-par/test-cluster-id"},
						},
						NodeType: "GP1-XS",
						Size:     3,
						Taints: []*ScalewayKapsulePoolTaint{
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
