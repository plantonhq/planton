package kubernetesrookcephoperatorv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/provider/kubernetes"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
)

func TestKubernetesRookCephOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesRookCephOperator Suite")
}

var _ = ginkgo.Describe("KubernetesRookCephOperator Validation Tests", func() {
	var input *KubernetesRookCephOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesRookCephOperator{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesRookCephOperator",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-rook-ceph-operator",
			},
			Spec: &KubernetesRookCephOperatorSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "rook-ceph",
					},
				},
				Container: &KubernetesRookCephOperatorSpecContainer{
					Resources: &kubernetes.ContainerResources{
						Limits: &kubernetes.CpuMemory{
							Cpu:    "500m",
							Memory: "512Mi",
						},
						Requests: &kubernetes.CpuMemory{
							Cpu:    "200m",
							Memory: "128Mi",
						},
					},
				},
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("with all required fields", func() {
			ginkgo.It("should not return a validation error", func() {
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with minimal configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				minimalInput := &KubernetesRookCephOperator{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesRookCephOperator",
					Metadata: &shared.CloudResourceMetadata{
						Name: "minimal-rook-ceph-operator",
					},
					Spec: &KubernetesRookCephOperatorSpec{
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "rook-ceph",
							},
						},
						Container: &KubernetesRookCephOperatorSpecContainer{},
					},
				}
				err := protovalidate.Validate(minimalInput)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with custom resource limits", func() {
			ginkgo.It("should not return a validation error", func() {
				customInput := &KubernetesRookCephOperator{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesRookCephOperator",
					Metadata: &shared.CloudResourceMetadata{
						Name: "custom-resources-operator",
					},
					Spec: &KubernetesRookCephOperatorSpec{
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "rook-ceph",
							},
						},
						Container: &KubernetesRookCephOperatorSpecContainer{
							Resources: &kubernetes.ContainerResources{
								Limits: &kubernetes.CpuMemory{
									Cpu:    "1000m",
									Memory: "1Gi",
								},
								Requests: &kubernetes.CpuMemory{
									Cpu:    "250m",
									Memory: "256Mi",
								},
							},
						},
					},
				}
				err := protovalidate.Validate(customInput)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with CSI configuration", func() {
			ginkgo.It("should not return a validation error", func() {
				enableRbd := true
				enableCephfs := true
				disableCsi := false
				enableHostNetwork := true
				provisionerReplicas := int32(3)
				enableCsiAddons := true
				enableNfs := false

				csiInput := &KubernetesRookCephOperator{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesRookCephOperator",
					Metadata: &shared.CloudResourceMetadata{
						Name: "csi-config-operator",
					},
					Spec: &KubernetesRookCephOperatorSpec{
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "rook-ceph",
							},
						},
						Container: &KubernetesRookCephOperatorSpecContainer{},
						Csi: &KubernetesRookCephOperatorCsiSpec{
							EnableRbdDriver:      &enableRbd,
							EnableCephfsDriver:   &enableCephfs,
							DisableCsiDriver:     &disableCsi,
							EnableCsiHostNetwork: &enableHostNetwork,
							ProvisionerReplicas:  &provisionerReplicas,
							EnableCsiAddons:      &enableCsiAddons,
							EnableNfsDriver:      &enableNfs,
						},
					},
				}
				err := protovalidate.Validate(csiInput)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with kubernetes cluster selector", func() {
			ginkgo.It("should not return a validation error", func() {
				selectorInput := &KubernetesRookCephOperator{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesRookCephOperator",
					Metadata: &shared.CloudResourceMetadata{
						Name: "selector-operator",
					},
					Spec: &KubernetesRookCephOperatorSpec{
						TargetCluster: &kubernetes.KubernetesClusterSelector{
							ClusterKind: cloudresourcekind.CloudResourceKind_AwsEksCluster,
							ClusterName: "prod-eks-cluster",
						},
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "rook-ceph",
							},
						},
						Container: &KubernetesRookCephOperatorSpecContainer{},
					},
				}
				err := protovalidate.Validate(selectorInput)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {
		ginkgo.Context("with incorrect api_version", func() {
			ginkgo.It("should return a validation error", func() {
				input.ApiVersion = "wrong-api-version"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with incorrect kind", func() {
			ginkgo.It("should return a validation error", func() {
				input.Kind = "WrongKind"
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing metadata", func() {
			ginkgo.It("should return a validation error", func() {
				input.Metadata = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing spec", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing container in spec", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Container = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("with missing namespace in spec", func() {
			ginkgo.It("should return a validation error", func() {
				input.Spec.Namespace = nil
				err := protovalidate.Validate(input)
				gomega.Expect(err).ToNot(gomega.BeNil())
			})
		})
	})
})
