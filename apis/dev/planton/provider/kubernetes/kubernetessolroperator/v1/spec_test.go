package kubernetessolroperatorv1

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

func TestKubernetesSolrOperator(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "KubernetesSolrOperator Suite")
}

var _ = ginkgo.Describe("KubernetesSolrOperator Validation Tests", func() {
	var input *KubernetesSolrOperator

	ginkgo.BeforeEach(func() {
		input = &KubernetesSolrOperator{
			ApiVersion: "kubernetes.planton.dev/v1",
			Kind:       "KubernetesSolrOperator",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-solr-operator",
			},
			Spec: &KubernetesSolrOperatorSpec{
				TargetCluster: &kubernetes.KubernetesClusterSelector{
					ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
					ClusterName: "test-cluster",
				},
				Namespace: &foreignkeyv1.StringValueOrRef{
					LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
						Value: "solr-operator-system",
					},
				},
				Container: &KubernetesSolrOperatorSpecContainer{
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
				minimalInput := &KubernetesSolrOperator{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesSolrOperator",
					Metadata: &shared.CloudResourceMetadata{
						Name: "minimal-solr-operator",
					},
					Spec: &KubernetesSolrOperatorSpec{
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "solr-operator-system",
							},
						},
						Container: &KubernetesSolrOperatorSpecContainer{},
					},
				}
				err := protovalidate.Validate(minimalInput)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("with custom resource limits", func() {
			ginkgo.It("should not return a validation error", func() {
				customInput := &KubernetesSolrOperator{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesSolrOperator",
					Metadata: &shared.CloudResourceMetadata{
						Name: "custom-resources-operator",
					},
					Spec: &KubernetesSolrOperatorSpec{
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "solr-operator-system",
							},
						},
						Container: &KubernetesSolrOperatorSpecContainer{
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
				selectorInput := &KubernetesSolrOperator{
					ApiVersion: "kubernetes.planton.dev/v1",
					Kind:       "KubernetesSolrOperator",
					Metadata: &shared.CloudResourceMetadata{
						Name: "selector-operator",
					},
					Spec: &KubernetesSolrOperatorSpec{
						TargetCluster: &kubernetes.KubernetesClusterSelector{
							ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
							ClusterName: "selector-cluster",
						},
						Namespace: &foreignkeyv1.StringValueOrRef{
							LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{
								Value: "solr-operator-system",
							},
						},
						Container: &KubernetesSolrOperatorSpecContainer{},
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
	})
})
