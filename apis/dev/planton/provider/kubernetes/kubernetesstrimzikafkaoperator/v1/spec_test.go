package kubernetesstrimzikafkaoperatorv1

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

func TestKubernetesStrimziKafkaOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesStrimziKafkaOperator Suite")
}

var _ = ginkgo.Describe("KubernetesStrimziKafkaOperator Validation Tests", func() {
	var input *KubernetesStrimziKafkaOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesStrimziKafkaOperator{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesStrimziKafkaOperator",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-kafka-operator",
			},
			Spec: &KubernetesStrimziKafkaOperatorSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "strimzi-kafka-operator",
					},
				},
				CreateNamespace: true,
				Container: &KubernetesStrimziKafkaOperatorSpecContainer{
					Resources: &kubernetes.ContainerResources{
						Limits: &kubernetes.CpuMemory{
							Cpu:    "1000m",
							Memory: "1Gi",
						},
						Requests: &kubernetes.CpuMemory{
							Cpu:    "50m",
							Memory: "100Mi",
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
				minimalInput := &KubernetesStrimziKafkaOperator{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesStrimziKafkaOperator",
					Metadata: &shared.CloudResourceMetadata{
						Name: "minimal-kafka-operator",
					},
					Spec: &KubernetesStrimziKafkaOperatorSpec{
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "strimzi-kafka-operator",
							},
						},
						CreateNamespace: false,
						Container:       &KubernetesStrimziKafkaOperatorSpecContainer{},
					},
				}
				err := protovalidate.Validate(minimalInput)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with custom resource limits", func() {
			ginkgo.It("should not return a validation error", func() {
				customInput := &KubernetesStrimziKafkaOperator{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesStrimziKafkaOperator",
					Metadata: &shared.CloudResourceMetadata{
						Name: "custom-resources-operator",
					},
					Spec: &KubernetesStrimziKafkaOperatorSpec{
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "strimzi-kafka-operator",
							},
						},
						CreateNamespace: true,
						Container: &KubernetesStrimziKafkaOperatorSpecContainer{
							Resources: &kubernetes.ContainerResources{
								Limits: &kubernetes.CpuMemory{
									Cpu:    "2000m",
									Memory: "2Gi",
								},
								Requests: &kubernetes.CpuMemory{
									Cpu:    "100m",
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

		ginkgo.Context("with kubernetes cluster selector", func() {
			ginkgo.It("should not return a validation error", func() {
				selectorInput := &KubernetesStrimziKafkaOperator{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesStrimziKafkaOperator",
					Metadata: &shared.CloudResourceMetadata{
						Name: "selector-operator",
					},
					Spec: &KubernetesStrimziKafkaOperatorSpec{
						TargetCluster: &kubernetes.KubernetesClusterSelector{
							ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
							ClusterName: "selector-cluster",
						},
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "strimzi-kafka-operator",
							},
						},
						CreateNamespace: true,
						Container:       &KubernetesStrimziKafkaOperatorSpecContainer{},
					},
				}
				err := protovalidate.Validate(selectorInput)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with create_namespace set to true", func() {
			ginkgo.It("should not return a validation error", func() {
				createNsInput := &KubernetesStrimziKafkaOperator{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesStrimziKafkaOperator",
					Metadata: &shared.CloudResourceMetadata{
						Name: "create-ns-operator",
					},
					Spec: &KubernetesStrimziKafkaOperatorSpec{
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "strimzi-kafka-operator",
							},
						},
						CreateNamespace: true,
						Container:       &KubernetesStrimziKafkaOperatorSpecContainer{},
					},
				}
				err := protovalidate.Validate(createNsInput)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with create_namespace set to false", func() {
			ginkgo.It("should not return a validation error", func() {
				noCreateNsInput := &KubernetesStrimziKafkaOperator{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesStrimziKafkaOperator",
					Metadata: &shared.CloudResourceMetadata{
						Name: "no-create-ns-operator",
					},
					Spec: &KubernetesStrimziKafkaOperatorSpec{
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "strimzi-kafka-operator",
							},
						},
						CreateNamespace: false,
						Container:       &KubernetesStrimziKafkaOperatorSpecContainer{},
					},
				}
				err := protovalidate.Validate(noCreateNsInput)
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
	})
})
